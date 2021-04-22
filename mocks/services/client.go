// Code generated by mockery 2.7.4. DO NOT EDIT.

package services

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Balance provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Client) Balance(_a0 context.Context, _a1 *types.AccountIdentifier, _a2 *types.PartialBlockIdentifier, _a3 []*types.Currency) (*types.AccountBalanceResponse, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 *types.AccountBalanceResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.AccountIdentifier, *types.PartialBlockIdentifier, []*types.Currency) *types.AccountBalanceResponse); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.AccountBalanceResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.AccountIdentifier, *types.PartialBlockIdentifier, []*types.Currency) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Block provides a mock function with given fields: _a0, _a1
func (_m *Client) Block(_a0 context.Context, _a1 *types.PartialBlockIdentifier) (*types.BlockResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *types.BlockResponse
	if rf, ok := ret.Get(0).(func(context.Context, *types.PartialBlockIdentifier) *types.BlockResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *types.PartialBlockIdentifier) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Status provides a mock function with given fields: _a0
func (_m *Client) Status(_a0 context.Context) (*types.BlockIdentifier, int64, *types.BlockIdentifier, *types.SyncStatus, []*types.Peer, error) {
	ret := _m.Called(_a0)

	var r0 *types.BlockIdentifier
	if rf, ok := ret.Get(0).(func(context.Context) *types.BlockIdentifier); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.BlockIdentifier)
		}
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(context.Context) int64); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Get(1).(int64)
	}

	var r2 *types.BlockIdentifier
	if rf, ok := ret.Get(2).(func(context.Context) *types.BlockIdentifier); ok {
		r2 = rf(_a0)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*types.BlockIdentifier)
		}
	}

	var r3 *types.SyncStatus
	if rf, ok := ret.Get(3).(func(context.Context) *types.SyncStatus); ok {
		r3 = rf(_a0)
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(*types.SyncStatus)
		}
	}

	var r4 []*types.Peer
	if rf, ok := ret.Get(4).(func(context.Context) []*types.Peer); ok {
		r4 = rf(_a0)
	} else {
		if ret.Get(4) != nil {
			r4 = ret.Get(4).([]*types.Peer)
		}
	}

	var r5 error
	if rf, ok := ret.Get(5).(func(context.Context) error); ok {
		r5 = rf(_a0)
	} else {
		r5 = ret.Error(5)
	}

	return r0, r1, r2, r3, r4, r5
}
