// Code generated by mockery 2.7.4. DO NOT EDIT.

package mocks

import (
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bytes "github.com/tendermint/tendermint/libs/bytes"

	client "github.com/tendermint/tendermint/rpc/client"

	context "context"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	log "github.com/tendermint/tendermint/libs/log"

	mock "github.com/stretchr/testify/mock"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tenderminttypes "github.com/tendermint/tendermint/types"

	types "github.com/cosmos/cosmos-sdk/types"
)

// RPCClient is an autogenerated mock type for the RPCClient type
type RPCClient struct {
	mock.Mock
}

// ABCIInfo provides a mock function with given fields:
func (_m *RPCClient) ABCIInfo() (*coretypes.ResultABCIInfo, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultABCIInfo
	if rf, ok := ret.Get(0).(func() *coretypes.ResultABCIInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIInfo)
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

// ABCIQuery provides a mock function with given fields: path, data
func (_m *RPCClient) ABCIQuery(path string, data bytes.HexBytes) (*coretypes.ResultABCIQuery, error) {
	ret := _m.Called(path, data)

	var r0 *coretypes.ResultABCIQuery
	if rf, ok := ret.Get(0).(func(string, bytes.HexBytes) *coretypes.ResultABCIQuery); ok {
		r0 = rf(path, data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIQuery)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, bytes.HexBytes) error); ok {
		r1 = rf(path, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ABCIQueryWithOptions provides a mock function with given fields: path, data, opts
func (_m *RPCClient) ABCIQueryWithOptions(path string, data bytes.HexBytes, opts client.ABCIQueryOptions) (*coretypes.ResultABCIQuery, error) {
	ret := _m.Called(path, data, opts)

	var r0 *coretypes.ResultABCIQuery
	if rf, ok := ret.Get(0).(func(string, bytes.HexBytes, client.ABCIQueryOptions) *coretypes.ResultABCIQuery); ok {
		r0 = rf(path, data, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultABCIQuery)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, bytes.HexBytes, client.ABCIQueryOptions) error); ok {
		r1 = rf(path, data, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Account provides a mock function with given fields: addr, height
func (_m *RPCClient) Account(addr types.AccAddress, height int64) (authtypes.AccountI, error) {
	ret := _m.Called(addr, height)

	var r0 authtypes.AccountI
	if rf, ok := ret.Get(0).(func(types.AccAddress, int64) authtypes.AccountI); ok {
		r0 = rf(addr, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authtypes.AccountI)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress, int64) error); ok {
		r1 = rf(addr, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Block provides a mock function with given fields: height
func (_m *RPCClient) Block(height *int64) (*coretypes.ResultBlock, error) {
	ret := _m.Called(height)

	var r0 *coretypes.ResultBlock
	if rf, ok := ret.Get(0).(func(*int64) *coretypes.ResultBlock); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlock)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockByHash provides a mock function with given fields: _a0
func (_m *RPCClient) BlockByHash(_a0 []byte) (*coretypes.ResultBlock, error) {
	ret := _m.Called(_a0)

	var r0 *coretypes.ResultBlock
	if rf, ok := ret.Get(0).(func([]byte) *coretypes.ResultBlock); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlock)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockResults provides a mock function with given fields: height
func (_m *RPCClient) BlockResults(height *int64) (*coretypes.ResultBlockResults, error) {
	ret := _m.Called(height)

	var r0 *coretypes.ResultBlockResults
	if rf, ok := ret.Get(0).(func(*int64) *coretypes.ResultBlockResults); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlockResults)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BlockchainInfo provides a mock function with given fields: minHeight, maxHeight
func (_m *RPCClient) BlockchainInfo(minHeight int64, maxHeight int64) (*coretypes.ResultBlockchainInfo, error) {
	ret := _m.Called(minHeight, maxHeight)

	var r0 *coretypes.ResultBlockchainInfo
	if rf, ok := ret.Get(0).(func(int64, int64) *coretypes.ResultBlockchainInfo); ok {
		r0 = rf(minHeight, maxHeight)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlockchainInfo)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64, int64) error); ok {
		r1 = rf(minHeight, maxHeight)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastEvidence provides a mock function with given fields: ev
func (_m *RPCClient) BroadcastEvidence(ev tenderminttypes.Evidence) (*coretypes.ResultBroadcastEvidence, error) {
	ret := _m.Called(ev)

	var r0 *coretypes.ResultBroadcastEvidence
	if rf, ok := ret.Get(0).(func(tenderminttypes.Evidence) *coretypes.ResultBroadcastEvidence); ok {
		r0 = rf(ev)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastEvidence)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tenderminttypes.Evidence) error); ok {
		r1 = rf(ev)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxAsync provides a mock function with given fields: tx
func (_m *RPCClient) BroadcastTxAsync(tx tenderminttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(tx)

	var r0 *coretypes.ResultBroadcastTx
	if rf, ok := ret.Get(0).(func(tenderminttypes.Tx) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tenderminttypes.Tx) error); ok {
		r1 = rf(tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxCommit provides a mock function with given fields: tx
func (_m *RPCClient) BroadcastTxCommit(tx tenderminttypes.Tx) (*coretypes.ResultBroadcastTxCommit, error) {
	ret := _m.Called(tx)

	var r0 *coretypes.ResultBroadcastTxCommit
	if rf, ok := ret.Get(0).(func(tenderminttypes.Tx) *coretypes.ResultBroadcastTxCommit); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTxCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tenderminttypes.Tx) error); ok {
		r1 = rf(tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BroadcastTxSync provides a mock function with given fields: tx
func (_m *RPCClient) BroadcastTxSync(tx tenderminttypes.Tx) (*coretypes.ResultBroadcastTx, error) {
	ret := _m.Called(tx)

	var r0 *coretypes.ResultBroadcastTx
	if rf, ok := ret.Get(0).(func(tenderminttypes.Tx) *coretypes.ResultBroadcastTx); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBroadcastTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tenderminttypes.Tx) error); ok {
		r1 = rf(tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Commit provides a mock function with given fields: height
func (_m *RPCClient) Commit(height *int64) (*coretypes.ResultCommit, error) {
	ret := _m.Called(height)

	var r0 *coretypes.ResultCommit
	if rf, ok := ret.Get(0).(func(*int64) *coretypes.ResultCommit); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultCommit)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsensusParams provides a mock function with given fields: height
func (_m *RPCClient) ConsensusParams(height *int64) (*coretypes.ResultConsensusParams, error) {
	ret := _m.Called(height)

	var r0 *coretypes.ResultConsensusParams
	if rf, ok := ret.Get(0).(func(*int64) *coretypes.ResultConsensusParams); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultConsensusParams)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConsensusState provides a mock function with given fields:
func (_m *RPCClient) ConsensusState() (*coretypes.ResultConsensusState, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultConsensusState
	if rf, ok := ret.Get(0).(func() *coretypes.ResultConsensusState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultConsensusState)
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

// Delegations provides a mock function with given fields: addr, height
func (_m *RPCClient) Delegations(addr types.AccAddress, height int64) (stakingtypes.DelegationResponses, error) {
	ret := _m.Called(addr, height)

	var r0 stakingtypes.DelegationResponses
	if rf, ok := ret.Get(0).(func(types.AccAddress, int64) stakingtypes.DelegationResponses); ok {
		r0 = rf(addr, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(stakingtypes.DelegationResponses)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress, int64) error); ok {
		r1 = rf(addr, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DumpConsensusState provides a mock function with given fields:
func (_m *RPCClient) DumpConsensusState() (*coretypes.ResultDumpConsensusState, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultDumpConsensusState
	if rf, ok := ret.Get(0).(func() *coretypes.ResultDumpConsensusState); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultDumpConsensusState)
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

// Genesis provides a mock function with given fields:
func (_m *RPCClient) Genesis() (*coretypes.ResultGenesis, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultGenesis
	if rf, ok := ret.Get(0).(func() *coretypes.ResultGenesis); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultGenesis)
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

// Health provides a mock function with given fields:
func (_m *RPCClient) Health() (*coretypes.ResultHealth, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultHealth
	if rf, ok := ret.Get(0).(func() *coretypes.ResultHealth); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultHealth)
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

// IsRunning provides a mock function with given fields:
func (_m *RPCClient) IsRunning() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// NetInfo provides a mock function with given fields:
func (_m *RPCClient) NetInfo() (*coretypes.ResultNetInfo, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultNetInfo
	if rf, ok := ret.Get(0).(func() *coretypes.ResultNetInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultNetInfo)
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

// NumUnconfirmedTxs provides a mock function with given fields:
func (_m *RPCClient) NumUnconfirmedTxs() (*coretypes.ResultUnconfirmedTxs, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultUnconfirmedTxs
	if rf, ok := ret.Get(0).(func() *coretypes.ResultUnconfirmedTxs); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultUnconfirmedTxs)
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

// OnReset provides a mock function with given fields:
func (_m *RPCClient) OnReset() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStart provides a mock function with given fields:
func (_m *RPCClient) OnStart() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OnStop provides a mock function with given fields:
func (_m *RPCClient) OnStop() {
	_m.Called()
}

// Quit provides a mock function with given fields:
func (_m *RPCClient) Quit() <-chan struct{} {
	ret := _m.Called()

	var r0 <-chan struct{}
	if rf, ok := ret.Get(0).(func() <-chan struct{}); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan struct{})
		}
	}

	return r0
}

// Reset provides a mock function with given fields:
func (_m *RPCClient) Reset() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetLogger provides a mock function with given fields: _a0
func (_m *RPCClient) SetLogger(_a0 log.Logger) {
	_m.Called(_a0)
}

// SimulateTx provides a mock function with given fields: tx
func (_m *RPCClient) SimulateTx(tx *legacytx.StdTx) (*types.SimulationResponse, error) {
	ret := _m.Called(tx)

	var r0 *types.SimulationResponse
	if rf, ok := ret.Get(0).(func(*legacytx.StdTx) *types.SimulationResponse); ok {
		r0 = rf(tx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.SimulationResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*legacytx.StdTx) error); ok {
		r1 = rf(tx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Start provides a mock function with given fields:
func (_m *RPCClient) Start() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Status provides a mock function with given fields:
func (_m *RPCClient) Status() (*coretypes.ResultStatus, error) {
	ret := _m.Called()

	var r0 *coretypes.ResultStatus
	if rf, ok := ret.Get(0).(func() *coretypes.ResultStatus); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultStatus)
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

// Stop provides a mock function with given fields:
func (_m *RPCClient) Stop() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// String provides a mock function with given fields:
func (_m *RPCClient) String() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Subscribe provides a mock function with given fields: ctx, subscriber, query, outCapacity
func (_m *RPCClient) Subscribe(ctx context.Context, subscriber string, query string, outCapacity ...int) (<-chan coretypes.ResultEvent, error) {
	_va := make([]interface{}, len(outCapacity))
	for _i := range outCapacity {
		_va[_i] = outCapacity[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, subscriber, query)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 <-chan coretypes.ResultEvent
	if rf, ok := ret.Get(0).(func(context.Context, string, string, ...int) <-chan coretypes.ResultEvent); ok {
		r0 = rf(ctx, subscriber, query, outCapacity...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan coretypes.ResultEvent)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, ...int) error); ok {
		r1 = rf(ctx, subscriber, query, outCapacity...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Tx provides a mock function with given fields: hash, prove
func (_m *RPCClient) Tx(hash []byte, prove bool) (*coretypes.ResultTx, error) {
	ret := _m.Called(hash, prove)

	var r0 *coretypes.ResultTx
	if rf, ok := ret.Get(0).(func([]byte, bool) *coretypes.ResultTx); ok {
		r0 = rf(hash, prove)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTx)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]byte, bool) error); ok {
		r1 = rf(hash, prove)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TxSearch provides a mock function with given fields: query, prove, page, perPage, orderBy
func (_m *RPCClient) TxSearch(query string, prove bool, page int, perPage int, orderBy string) (*coretypes.ResultTxSearch, error) {
	ret := _m.Called(query, prove, page, perPage, orderBy)

	var r0 *coretypes.ResultTxSearch
	if rf, ok := ret.Get(0).(func(string, bool, int, int, string) *coretypes.ResultTxSearch); ok {
		r0 = rf(query, prove, page, perPage, orderBy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTxSearch)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, bool, int, int, string) error); ok {
		r1 = rf(query, prove, page, perPage, orderBy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnbondingDelegations provides a mock function with given fields: addr, height
func (_m *RPCClient) UnbondingDelegations(addr types.AccAddress, height int64) (stakingtypes.UnbondingDelegations, error) {
	ret := _m.Called(addr, height)

	var r0 stakingtypes.UnbondingDelegations
	if rf, ok := ret.Get(0).(func(types.AccAddress, int64) stakingtypes.UnbondingDelegations); ok {
		r0 = rf(addr, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(stakingtypes.UnbondingDelegations)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress, int64) error); ok {
		r1 = rf(addr, height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnconfirmedTxs provides a mock function with given fields: limit
func (_m *RPCClient) UnconfirmedTxs(limit int) (*coretypes.ResultUnconfirmedTxs, error) {
	ret := _m.Called(limit)

	var r0 *coretypes.ResultUnconfirmedTxs
	if rf, ok := ret.Get(0).(func(int) *coretypes.ResultUnconfirmedTxs); ok {
		r0 = rf(limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultUnconfirmedTxs)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Unsubscribe provides a mock function with given fields: ctx, subscriber, query
func (_m *RPCClient) Unsubscribe(ctx context.Context, subscriber string, query string) error {
	ret := _m.Called(ctx, subscriber, query)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, subscriber, query)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsubscribeAll provides a mock function with given fields: ctx, subscriber
func (_m *RPCClient) UnsubscribeAll(ctx context.Context, subscriber string) error {
	ret := _m.Called(ctx, subscriber)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, subscriber)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Validators provides a mock function with given fields: height, page, perPage
func (_m *RPCClient) Validators(height *int64, page int, perPage int) (*coretypes.ResultValidators, error) {
	ret := _m.Called(height, page, perPage)

	var r0 *coretypes.ResultValidators
	if rf, ok := ret.Get(0).(func(*int64, int, int) *coretypes.ResultValidators); ok {
		r0 = rf(height, page, perPage)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultValidators)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64, int, int) error); ok {
		r1 = rf(height, page, perPage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
