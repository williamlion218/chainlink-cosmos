package client

import (
	"fmt"
	"os"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-cosmos/pkg/cosmos/params"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestMain(m *testing.M) {
	// these are hardcoded in test_helpers.go.
	params.InitCosmosSdk(
		/* bech32Prefix= */ "wasm",
		/* token= */ "cosm",
	)
	code := m.Run()
	os.Exit(code)
}

func TestErrMatch(t *testing.T) {
	errStr := "rpc error: code = InvalidArgument desc = failed to execute message; message index: 0: Error parsing into type my_first_contract::msg::ExecuteMsg: unknown variant `blah`, expected `increment` or `reset`: execute wasm contract failed: invalid request"
	m := failedMsgIndexRe.FindStringSubmatch(errStr)
	require.Equal(t, 2, len(m))
	assert.Equal(t, m[1], "0")

	errStr = "rpc error: code = InvalidArgument desc = failed to execute message; message index: 10: Error parsing into type my_first_contract::msg::ExecuteMsg: unknown variant `blah`, expected `increment` or `reset`: execute wasm contract failed: invalid request"
	m = failedMsgIndexRe.FindStringSubmatch(errStr)
	require.Equal(t, 2, len(m))
	assert.Equal(t, m[1], "10")

	errStr = "rpc error: code = InvalidArgument desc = failed to execute message; message index: 10000: Error parsing into type my_first_contract::msg::ExecuteMsg: unknown variant `blah`, expected `increment` or `reset`: execute wasm contract failed: invalid request"
	m = failedMsgIndexRe.FindStringSubmatch(errStr)
	require.Equal(t, 2, len(m))
	assert.Equal(t, m[1], "10000")
}

func TestBatchSim(t *testing.T) {
	accounts, testdir, tendermintURL := SetupLocalCosmosNode(t, "42", "ucosm")

	lggr, logs := logger.TestObserved(t, zap.WarnLevel)
	tc, err := NewClient(
		"42",
		tendermintURL,
		DefaultTimeout,
		lggr)
	require.NoError(t, err)
	assertLogsLen := func(t *testing.T, l int) func() {
		return func() { assert.Len(t, logs.TakeAll(), l) }
	}

	contract := DeployTestContract(t, tendermintURL, "42", "ucosm", accounts[0], accounts[0], tc, testdir, "../testdata/my_first_contract.wasm")
	var succeed sdk.Msg = &wasmtypes.MsgExecuteContract{Sender: accounts[0].Address.String(), Contract: contract.String(), Msg: []byte(`{"reset":{"count":5}}`)}
	var fail sdk.Msg = &wasmtypes.MsgExecuteContract{Sender: accounts[0].Address.String(), Contract: contract.String(), Msg: []byte(`{"blah":{"count":5}}`)}

	t.Run("single success", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 0))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: succeed}}, sn)
		require.NoError(t, err)
		require.Equal(t, 1, len(res.Succeeded))
		assert.Equal(t, int64(1), res.Succeeded[0].ID)
		assert.Equal(t, 0, len(res.Failed))
	})

	t.Run("single failure", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 1))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: fail}}, sn)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res.Succeeded))
		require.Equal(t, 1, len(res.Failed))
		assert.Equal(t, int64(1), res.Failed[0].ID)
	})

	t.Run("multi failure", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 2))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: succeed}, {ID: int64(2), Msg: fail}, {ID: int64(3), Msg: fail}}, sn)
		require.NoError(t, err)
		require.Equal(t, 1, len(res.Succeeded))
		assert.Equal(t, int64(1), res.Succeeded[0].ID)
		require.Equal(t, 2, len(res.Failed))
		assert.Equal(t, int64(2), res.Failed[0].ID)
		assert.Equal(t, int64(3), res.Failed[1].ID)
	})

	t.Run("multi succeed", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 1))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: succeed}, {ID: int64(2), Msg: succeed}, {ID: int64(3), Msg: fail}}, sn)
		require.NoError(t, err)
		assert.Equal(t, 2, len(res.Succeeded))
		assert.Equal(t, 1, len(res.Failed))
	})

	t.Run("all succeed", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 0))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: succeed}, {ID: int64(2), Msg: succeed}, {ID: int64(3), Msg: succeed}}, sn)
		require.NoError(t, err)
		assert.Equal(t, 3, len(res.Succeeded))
		assert.Equal(t, 0, len(res.Failed))
	})

	t.Run("all fail", func(t *testing.T) {
		ctx := tests.Context(t)
		_, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		t.Cleanup(assertLogsLen(t, 3))
		res, err := tc.BatchSimulateUnsigned(ctx, []SimMsg{{ID: int64(1), Msg: fail}, {ID: int64(2), Msg: fail}, {ID: int64(3), Msg: fail}}, sn)
		require.NoError(t, err)
		assert.Equal(t, 0, len(res.Succeeded))
		assert.Equal(t, 3, len(res.Failed))
	})
}

func TestCosmosClient(t *testing.T) {
	minGasPrice := sdk.NewDecCoinFromDec("ucosm", defaultCoin)
	// Local only for now, could maybe run on CI if we install terrad there?
	accounts, testdir, tendermintURL := SetupLocalCosmosNode(t, "42", "ucosm")
	lggr := logger.Sugared(logger.Test(t))
	tc, err := NewClient(
		"42",
		tendermintURL,
		DefaultTimeout,
		lggr)
	require.NoError(t, err)
	gpe := NewFixedGasPriceEstimator(map[string]sdk.DecCoin{"ucosm": sdk.NewDecCoinFromDec("ucosm", sdk.MustNewDecFromStr("0.01"))}, lggr)
	contract := DeployTestContract(t, tendermintURL, "42", "ucosm", accounts[0], accounts[0], tc, testdir, "../testdata/my_first_contract.wasm")

	t.Run("send tx between accounts", func(t *testing.T) {
		ctx := tests.Context(t)
		// Assert balance before
		b, err := tc.Balance(ctx, accounts[1].Address, "ucosm")
		require.NoError(t, err)
		assert.Equal(t, "100000000", b.Amount.String())

		// Send a ucosm from one account to another and ensure balances update
		an, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		fund := banktypes.NewMsgSend(accounts[0].Address, accounts[1].Address, sdk.NewCoins(sdk.NewInt64Coin("ucosm", 1)))
		gasLimit, err := tc.SimulateUnsigned(ctx, []sdk.Msg{fund}, sn)
		require.NoError(t, err)
		gasPrices, err := gpe.GasPrices()
		require.NoError(t, err)
		txBytes, err := tc.CreateAndSign([]sdk.Msg{fund}, an, sn, gasLimit.GasInfo.GasUsed, DefaultGasLimitMultiplier, gasPrices["ucosm"], accounts[0].PrivateKey, 0)
		require.NoError(t, err)
		_, err = tc.Simulate(ctx, txBytes)
		require.NoError(t, err)
		resp, err := tc.Broadcast(ctx, txBytes, txtypes.BroadcastMode_BROADCAST_MODE_SYNC)
		require.NoError(t, err)
		tx, success := AwaitTxCommitted(t, tc, resp.TxResponse.TxHash)
		require.True(t, success)
		require.Equal(t, types.CodeTypeOK, tx.TxResponse.Code)

		// Assert balance changed
		b, err = tc.Balance(ctx, accounts[1].Address, "ucosm")
		require.NoError(t, err)
		assert.Equal(t, "100000001", b.Amount.String())

		// Invalid tx should error
		_, err = tc.Tx(ctx, "1234")
		require.Error(t, err)

		// Ensure we can read back the tx with Query
		tr, err := tc.TxsEvents(ctx, []string{fmt.Sprintf("tx.height=%v", tx.TxResponse.Height)}, nil)
		require.NoError(t, err)
		assert.Equal(t, 1, len(tr.TxResponses))
		assert.Equal(t, tx.TxResponse.TxHash, tr.TxResponses[0].TxHash)
		// And also Tx
		getTx, err := tc.Tx(ctx, tx.TxResponse.TxHash)
		require.NoError(t, err)
		assert.Equal(t, getTx.TxResponse.TxHash, tx.TxResponse.TxHash)
	})

	t.Run("can get height", func(t *testing.T) {
		// Check getting the height works
		latestBlock, err := tc.LatestBlock(tests.Context(t))
		require.NoError(t, err)
		assert.True(t, latestBlock.SdkBlock.Header.Height > 1)
	})

	t.Run("contract event querying", func(t *testing.T) {
		ctx := tests.Context(t)
		// Query initial contract state
		count, err := tc.ContractState(
			ctx,
			contract,
			[]byte(`{"get_count":{}}`),
		)
		require.NoError(t, err)
		assert.Equal(t, `{"count":0}`, string(count))
		// Query invalid state should give an error
		count, err = tc.ContractState(
			ctx,
			contract,
			[]byte(`{"blah":{}}`),
		)
		require.Error(t, err)
		require.Nil(t, count)

		// Change the contract state
		rawMsg := &wasmtypes.MsgExecuteContract{
			Sender:   accounts[0].Address.String(),
			Contract: contract.String(),
			Msg:      []byte(`{"reset":{"count":5}}`),
			Funds:    sdk.Coins{},
		}
		an, sn, err := tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		gasPrices, err := gpe.GasPrices()
		require.NoError(t, err)
		resp1, err := tc.SignAndBroadcast(ctx, []sdk.Msg{rawMsg}, an, sn, gasPrices["ucosm"], accounts[0].PrivateKey, txtypes.BroadcastMode_BROADCAST_MODE_SYNC)
		require.NoError(t, err)
		tx1, success := AwaitTxCommitted(t, tc, resp1.TxResponse.TxHash)
		require.True(t, success)
		require.Equal(t, types.CodeTypeOK, tx1.TxResponse.Code)

		// Do it again so there are multiple executions
		rawMsg = &wasmtypes.MsgExecuteContract{
			Sender:   accounts[0].Address.String(),
			Contract: contract.String(),
			Msg:      []byte(`{"reset":{"count":4}}`),
			Funds:    sdk.Coins{},
		}
		an, sn, err = tc.Account(ctx, accounts[0].Address)
		require.NoError(t, err)
		resp2, err := tc.SignAndBroadcast(ctx, []sdk.Msg{rawMsg}, an, sn, gasPrices["ucosm"], accounts[0].PrivateKey, txtypes.BroadcastMode_BROADCAST_MODE_SYNC)
		require.NoError(t, err)
		tx2, success := AwaitTxCommitted(t, tc, resp2.TxResponse.TxHash)
		require.True(t, success)
		require.Equal(t, types.CodeTypeOK, tx2.TxResponse.Code)

		// Observe changed contract state
		count, err = tc.ContractState(
			ctx,
			contract,
			[]byte(`{"get_count":{}}`),
		)
		require.NoError(t, err)
		assert.Equal(t, `{"count":4}`, string(count))

		// Check events querying works
		// TxEvents sorts in a descending manner, so latest txes are first
		ev, err := tc.TxsEvents(ctx, []string{"wasm.action='reset'", fmt.Sprintf("wasm._contract_address='%s'", contract.String())}, nil)
		require.NoError(t, err)
		require.Equal(t, 2, len(ev.TxResponses))
		foundContract := false
		for _, event := range ev.TxResponses[0].Logs[0].Events {
			if event.Type != "wasm" {
				continue
			}
			isResetAction := false
			for _, attr := range event.Attributes {
				if attr.Key != "action" {
					continue
				}
				isResetAction = attr.Value == "reset"
				break
			}
			if !isResetAction {
				continue
			}
			for _, attr := range event.Attributes {
				if attr.Key == "_contract_address" {
					assert.Equal(t, contract.String(), attr.Value)
					foundContract = true
				}
			}
		}
		assert.True(t, foundContract)

		// Ensure the height filtering works
		ev, err = tc.TxsEvents(ctx, []string{fmt.Sprintf("tx.height=%d", tx2.TxResponse.Height), "wasm.action='reset'", fmt.Sprintf("wasm._contract_address='%s'", contract.String())}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, len(ev.TxResponses))
		ev, err = tc.TxsEvents(ctx, []string{fmt.Sprintf("tx.height=%d", tx1.TxResponse.Height), "wasm.action='reset'", fmt.Sprintf("wasm._contract_address='%s'", contract)}, nil)
		require.NoError(t, err)
		require.Equal(t, 1, len(ev.TxResponses))
		for _, ev := range ev.TxResponses[0].Logs[0].Events {
			if ev.Type == "wasm-reset" {
				for _, attr := range ev.Attributes {
					t.Log(attr.Key, attr.Value)
				}
			}
		}
	})

	t.Run("gasprice", func(t *testing.T) {
		rawMsg := &wasmtypes.MsgExecuteContract{
			Sender:   accounts[0].Address.String(),
			Contract: contract.String(),
			Msg:      []byte(`{"reset":{"count":5}}`),
			Funds:    sdk.Coins{},
		}
		const expCodespace = sdkerrors.RootCodespace
		gasPrices, err := gpe.GasPrices()
		require.NoError(t, err)
		for _, tt := range []struct {
			name     string
			gasPrice sdk.DecCoin
			expCode  uint32
		}{
			{
				"zero",
				sdk.NewInt64DecCoin("ucosm", 0),
				sdkerrors.ErrInsufficientFee.ABCICode(),
			},
			{
				"below-min",
				sdk.NewDecCoinFromDec("ucosm", sdk.NewDecWithPrec(1, 4)),
				sdkerrors.ErrInsufficientFee.ABCICode(),
			},
			{
				"min",
				minGasPrice,
				0,
			},
			{
				"recommended",
				gasPrices["ucosm"],
				0,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				ctx := tests.Context(t)
				t.Log("Gas price:", tt.gasPrice)
				an, sn, err := tc.Account(ctx, accounts[0].Address)
				require.NoError(t, err)
				resp, err := tc.SignAndBroadcast(ctx, []sdk.Msg{rawMsg}, an, sn, tt.gasPrice, accounts[0].PrivateKey, txtypes.BroadcastMode_BROADCAST_MODE_SYNC)
				require.NotNil(t, resp)

				if tt.expCode == 0 {
					require.NoError(t, err)
					tx, success := AwaitTxCommitted(t, tc, resp.TxResponse.TxHash)
					require.True(t, success)
					require.Equal(t, types.CodeTypeOK, tx.TxResponse.Code)
					require.Equal(t, "", tx.TxResponse.Codespace)
					require.Equal(t, tt.expCode, tx.TxResponse.Code)
					require.Equal(t, resp.TxResponse.TxHash, tx.TxResponse.TxHash)
					t.Log("Fee:", tx.Tx.GetFee())
					t.Log("Height:", tx.TxResponse.Height)
				} else {
					require.Error(t, err)
					require.Equal(t, expCodespace, resp.TxResponse.Codespace)
					require.Equal(t, tt.expCode, resp.TxResponse.Code)
				}
			})
		}
	})
}
