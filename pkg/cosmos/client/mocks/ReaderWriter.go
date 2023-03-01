// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	client "github.com/smartcontractkit/chainlink-cosmos/pkg/cosmos/client"

	mock "github.com/stretchr/testify/mock"

	query "github.com/cosmos/cosmos-sdk/types/query"

	tmservice "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"

	tx "github.com/cosmos/cosmos-sdk/types/tx"

	types "github.com/cosmos/cosmos-sdk/types"
)

// ReaderWriter is an autogenerated mock type for the ReaderWriter type
type ReaderWriter struct {
	mock.Mock
}

// Account provides a mock function with given fields: address
func (_m *ReaderWriter) Account(address types.AccAddress) (uint64, uint64, error) {
	ret := _m.Called(address)

	var r0 uint64
	if rf, ok := ret.Get(0).(func(types.AccAddress) uint64); ok {
		r0 = rf(address)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 uint64
	if rf, ok := ret.Get(1).(func(types.AccAddress) uint64); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Get(1).(uint64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(types.AccAddress) error); ok {
		r2 = rf(address)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Balance provides a mock function with given fields: addr, denom
func (_m *ReaderWriter) Balance(addr types.AccAddress, denom string) (*types.Coin, error) {
	ret := _m.Called(addr, denom)

	var r0 *types.Coin
	if rf, ok := ret.Get(0).(func(types.AccAddress, string) *types.Coin); ok {
		r0 = rf(addr, denom)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Coin)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress, string) error); ok {
		r1 = rf(addr, denom)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BatchSimulateUnsigned provides a mock function with given fields: msgs, sequence
func (_m *ReaderWriter) BatchSimulateUnsigned(msgs client.SimMsgs, sequence uint64) (*client.BatchSimResults, error) {
	ret := _m.Called(msgs, sequence)

	var r0 *client.BatchSimResults
	if rf, ok := ret.Get(0).(func(client.SimMsgs, uint64) *client.BatchSimResults); ok {
		r0 = rf(msgs, sequence)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.BatchSimResults)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(client.SimMsgs, uint64) error); ok {
		r1 = rf(msgs, sequence)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockByHeight provides a mock function with given fields: height
func (_m *ReaderWriter) BlockByHeight(height int64) (*tmservice.GetBlockByHeightResponse, error) {
	ret := _m.Called(height)

	var r0 *tmservice.GetBlockByHeightResponse
	if rf, ok := ret.Get(0).(func(int64) *tmservice.GetBlockByHeightResponse); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tmservice.GetBlockByHeightResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Broadcast provides a mock function with given fields: txBytes, mode
func (_m *ReaderWriter) Broadcast(txBytes []byte, mode tx.BroadcastMode) (*tx.BroadcastTxResponse, error) {
	ret := _m.Called(txBytes, mode)

	var r0 *tx.BroadcastTxResponse
	if rf, ok := ret.Get(0).(func([]byte, tx.BroadcastMode) *tx.BroadcastTxResponse); ok {
		r0 = rf(txBytes, mode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.BroadcastTxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, tx.BroadcastMode) error); ok {
		r1 = rf(txBytes, mode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContractState provides a mock function with given fields: contractAddress, queryMsg
func (_m *ReaderWriter) ContractState(contractAddress types.AccAddress, queryMsg []byte) ([]byte, error) {
	ret := _m.Called(contractAddress, queryMsg)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(types.AccAddress, []byte) []byte); ok {
		r0 = rf(contractAddress, queryMsg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress, []byte) error); ok {
		r1 = rf(contractAddress, queryMsg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateAndSign provides a mock function with given fields: msgs, account, sequence, gasLimit, gasLimitMultiplier, gasPrice, signer, timeoutHeight
func (_m *ReaderWriter) CreateAndSign(msgs []types.Msg, account uint64, sequence uint64, gasLimit uint64, gasLimitMultiplier float64, gasPrice types.DecCoin, signer cryptotypes.PrivKey, timeoutHeight uint64) ([]byte, error) {
	ret := _m.Called(msgs, account, sequence, gasLimit, gasLimitMultiplier, gasPrice, signer, timeoutHeight)

	var r0 []byte
	if rf, ok := ret.Get(0).(func([]types.Msg, uint64, uint64, uint64, float64, types.DecCoin, cryptotypes.PrivKey, uint64) []byte); ok {
		r0 = rf(msgs, account, sequence, gasLimit, gasLimitMultiplier, gasPrice, signer, timeoutHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.Msg, uint64, uint64, uint64, float64, types.DecCoin, cryptotypes.PrivKey, uint64) error); ok {
		r1 = rf(msgs, account, sequence, gasLimit, gasLimitMultiplier, gasPrice, signer, timeoutHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LatestBlock provides a mock function with given fields:
func (_m *ReaderWriter) LatestBlock() (*tmservice.GetLatestBlockResponse, error) {
	ret := _m.Called()

	var r0 *tmservice.GetLatestBlockResponse
	if rf, ok := ret.Get(0).(func() *tmservice.GetLatestBlockResponse); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tmservice.GetLatestBlockResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignAndBroadcast provides a mock function with given fields: msgs, accountNum, sequence, gasPrice, signer, mode
func (_m *ReaderWriter) SignAndBroadcast(msgs []types.Msg, accountNum uint64, sequence uint64, gasPrice types.DecCoin, signer cryptotypes.PrivKey, mode tx.BroadcastMode) (*tx.BroadcastTxResponse, error) {
	ret := _m.Called(msgs, accountNum, sequence, gasPrice, signer, mode)

	var r0 *tx.BroadcastTxResponse
	if rf, ok := ret.Get(0).(func([]types.Msg, uint64, uint64, types.DecCoin, cryptotypes.PrivKey, tx.BroadcastMode) *tx.BroadcastTxResponse); ok {
		r0 = rf(msgs, accountNum, sequence, gasPrice, signer, mode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.BroadcastTxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.Msg, uint64, uint64, types.DecCoin, cryptotypes.PrivKey, tx.BroadcastMode) error); ok {
		r1 = rf(msgs, accountNum, sequence, gasPrice, signer, mode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Simulate provides a mock function with given fields: txBytes
func (_m *ReaderWriter) Simulate(txBytes []byte) (*tx.SimulateResponse, error) {
	ret := _m.Called(txBytes)

	var r0 *tx.SimulateResponse
	if rf, ok := ret.Get(0).(func([]byte) *tx.SimulateResponse); ok {
		r0 = rf(txBytes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.SimulateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(txBytes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SimulateUnsigned provides a mock function with given fields: msgs, sequence
func (_m *ReaderWriter) SimulateUnsigned(msgs []types.Msg, sequence uint64) (*tx.SimulateResponse, error) {
	ret := _m.Called(msgs, sequence)

	var r0 *tx.SimulateResponse
	if rf, ok := ret.Get(0).(func([]types.Msg, uint64) *tx.SimulateResponse); ok {
		r0 = rf(msgs, sequence)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.SimulateResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.Msg, uint64) error); ok {
		r1 = rf(msgs, sequence)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx provides a mock function with given fields: hash
func (_m *ReaderWriter) Tx(hash string) (*tx.GetTxResponse, error) {
	ret := _m.Called(hash)

	var r0 *tx.GetTxResponse
	if rf, ok := ret.Get(0).(func(string) *tx.GetTxResponse); ok {
		r0 = rf(hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.GetTxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(hash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TxsEvents provides a mock function with given fields: events, paginationParams
func (_m *ReaderWriter) TxsEvents(events []string, paginationParams *query.PageRequest) (*tx.GetTxsEventResponse, error) {
	ret := _m.Called(events, paginationParams)

	var r0 *tx.GetTxsEventResponse
	if rf, ok := ret.Get(0).(func([]string, *query.PageRequest) *tx.GetTxsEventResponse); ok {
		r0 = rf(events, paginationParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tx.GetTxsEventResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string, *query.PageRequest) error); ok {
		r1 = rf(events, paginationParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
