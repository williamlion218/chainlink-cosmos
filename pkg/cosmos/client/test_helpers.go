package client

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-cosmos/pkg/cosmos/testutil"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Account struct {
	Name       string
	PrivateKey cryptotypes.PrivKey
	Address    sdk.AccAddress
}

// 0.001
var minGasPrice = sdk.NewDecCoinFromDec("ucosm", sdk.NewDecWithPrec(1, 3))

// SetupLocalCosmosNode sets up a local terra node via wasmd, and returns pre-funded accounts, the test directory, and the url.
func SetupLocalCosmosNode(t *testing.T, chainID string) ([]Account, string, string) {
	testdir, err := os.MkdirTemp("", "integration-test")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(testdir))
	})
	t.Log(testdir)
	out, err := exec.Command("wasmd", "init", "integration-test", "-o", "--chain-id", chainID, "--home", testdir).Output()
	require.NoError(t, err, string(out))

	p := path.Join(testdir, "config", "app.toml")
	f, err := os.ReadFile(p)
	require.NoError(t, err)
	config, err := toml.Load(string(f))
	require.NoError(t, err)
	// Enable if desired to use lcd endpoints config.Set("api.enable", "true")
	config.Set("minimum-gas-prices", minGasPrice.String())
	require.NoError(t, os.WriteFile(p, []byte(config.String()), 0600))
	// TODO: could also speed up the block mining config

	p = path.Join(testdir, "config", "genesis.json")
	f, err = os.ReadFile(p)
	require.NoError(t, err)

	genesisData := string(f)
	// fix hardcoded token, see
	// https://github.com/CosmWasm/wasmd/blob/develop/docker/setup_wasmd.sh
	// https://github.com/CosmWasm/wasmd/blob/develop/contrib/local/setup_wasmd.sh
	genesisData = strings.ReplaceAll(genesisData, "\"ustake\"", "\"ucosm\"")
	genesisData = strings.ReplaceAll(genesisData, "\"stake\"", "\"ucosm\"")
	require.NoError(t, os.WriteFile(p, []byte(genesisData), 0600))

	// Create 2 test accounts
	var accounts []Account
	for i := 0; i < 2; i++ {
		account := fmt.Sprintf("test%d", i)
		key, err2 := exec.Command("wasmd", "keys", "add", account, "--output", "json", "--keyring-backend", "test", "--home", testdir).CombinedOutput()
		require.NoError(t, err2, string(key))
		var k struct {
			Address  string `json:"address"`
			Mnemonic string `json:"mnemonic"`
		}
		require.NoError(t, json.Unmarshal(key, &k))
		expAcctAddr, err3 := sdk.AccAddressFromBech32(k.Address)
		require.NoError(t, err3)
		privateKey, address, err4 := testutil.CreateKeyFromMnemonic(k.Mnemonic)
		require.NoError(t, err4)
		require.Equal(t, expAcctAddr, address)
		// Give it 100 luna
		out2, err2 := exec.Command("wasmd", "add-genesis-account", k.Address, "100000000ucosm", "--home", testdir).Output() //nolint:gosec
		require.NoError(t, err2, string(out2))
		accounts = append(accounts, Account{
			Name:       account,
			Address:    address,
			PrivateKey: privateKey,
		})
	}
	// Stake 10 luna in first acct
	out, err = exec.Command("wasmd", "gentx", accounts[0].Name, "10000000ucosm", "--chain-id", chainID, "--keyring-backend", "test", "--home", testdir).CombinedOutput() //nolint:gosec
	require.NoError(t, err, string(out))
	out, err = exec.Command("wasmd", "collect-gentxs", "--home", testdir).CombinedOutput()
	require.NoError(t, err, string(out))

	port := mustRandomPort()
	tendermintHost := fmt.Sprintf("127.0.0.1:%d", port)
	tendermintURL := "http://" + tendermintHost
	t.Log(tendermintURL)

	cmd := exec.Command("wasmd", "start", "--home", testdir,
		"--rpc.laddr", "tcp://"+tendermintHost,
		"--rpc.pprof_laddr", "127.0.0.1:0",
		"--grpc.address", "127.0.0.1:0",
		"--grpc-web.address", "127.0.0.1:0",
		"--p2p.laddr", "127.0.0.1:0")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	require.NoError(t, cmd.Start())
	t.Cleanup(func() {
		assert.NoError(t, cmd.Process.Kill())
		if err2 := cmd.Wait(); assert.Error(t, err2) {
			if !assert.Contains(t, err2.Error(), "signal: killed", cmd.ProcessState.String()) {
				t.Log("wasmd stderr:", stdErr.String())
			}
		}
	})

	// Wait for api server to boot
	var ready bool
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		out, err = exec.Command("curl", tendermintURL+"/abci_info").Output() //nolint:gosec
		if err != nil {
			t.Logf("API server not ready yet (attempt %d): %v\n", i+1, err)
			continue
		}
		var a struct {
			Result struct {
				Response struct {
					LastBlockHeight string `json:"last_block_height"`
				} `json:"response"`
			} `json:"result"`
		}
		require.NoError(t, json.Unmarshal(out, &a), string(out))
		if a.Result.Response.LastBlockHeight == "" {
			t.Logf("API server not ready yet (attempt %d)\n", i+1)
			continue
		}
		ready = true
		break
	}
	require.True(t, ready)
	return accounts, testdir, tendermintURL
}

// DeployTestContract deploys a test contract.
func DeployTestContract(t *testing.T, tendermintURL, chainID string, deployAccount, ownerAccount Account, tc *Client, testdir, wasmTestContractPath string) sdk.AccAddress {
	//nolint:gosec
	out, err := exec.Command("wasmd", "tx", "wasm", "store", wasmTestContractPath, "--node", tendermintURL,
		"--from", deployAccount.Name, "--gas", "auto", "--fees", "100000ucosm", "--gas-adjustment", "1.3", "--chain-id", chainID, "--broadcast-mode", "block", "--home", testdir, "--keyring-backend", "test", "--keyring-dir", testdir, "--yes", "--output", "json").CombinedOutput()
	require.NoError(t, err, string(out))
	an, sn, err2 := tc.Account(ownerAccount.Address)
	require.NoError(t, err2)
	r, err3 := tc.SignAndBroadcast([]sdk.Msg{
		&wasmtypes.MsgInstantiateContract{
			Sender: ownerAccount.Address.String(),
			Admin:  "",
			// TODO: this only works for the first code deployment, read code_id from the store invocation above and use the value here.
			CodeID: 1,
			Label:  "testcontract",
			Msg:    []byte(`{"count":0}`),
			Funds:  sdk.Coins{},
		},
	}, an, sn, minGasPrice, ownerAccount.PrivateKey, txtypes.BroadcastMode_BROADCAST_MODE_BLOCK)
	require.NoError(t, err3)
	return GetContractAddr(t, tc, r.TxResponse.TxHash)
}

func GetContractAddr(t *testing.T, tc *Client, deploymentHash string) sdk.AccAddress {
	var deploymentTx *txtypes.GetTxResponse
	var err error
	for try := 0; try < 5; try++ {
		deploymentTx, err = tc.Tx(deploymentHash)
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
	}
	require.NoError(t, err)
	var contractAddr string
	for _, etype := range deploymentTx.TxResponse.Events {
		if etype.Type == "wasm" {
			for _, attr := range etype.Attributes {
				if string(attr.Key) == "_contract_address" {
					contractAddr = string(attr.Value)
				}
			}
		}
	}
	require.NotEqual(t, "", contractAddr)
	contract, err := sdk.AccAddressFromBech32(contractAddr)
	require.NoError(t, err)
	return contract
}

func mustRandomPort() int {
	r, err := rand.Int(rand.Reader, big.NewInt(65535-1023))
	if err != nil {
		panic(fmt.Errorf("unexpected error generating random port: %w", err))
	}
	return int(r.Int64() + 1024)
}