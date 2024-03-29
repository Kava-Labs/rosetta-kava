// Code generated by mockery v2.12.3. DO NOT EDIT.

package mocks

import (
	context "context"

	kava "github.com/kava-labs/rosetta-kava/kava"
	mock "github.com/stretchr/testify/mock"

	tenderminttypes "github.com/tendermint/tendermint/types"

	types "github.com/cosmos/cosmos-sdk/types"
)

// BalanceServiceFactory is an autogenerated mock type for the BalanceServiceFactory type
type BalanceServiceFactory struct {
	mock.Mock
}

// Execute provides a mock function with given fields: ctx, addr, blockHeader
func (_m *BalanceServiceFactory) Execute(ctx context.Context, addr types.AccAddress, blockHeader *tenderminttypes.Header) (kava.AccountBalanceService, error) {
	ret := _m.Called(ctx, addr, blockHeader)

	var r0 kava.AccountBalanceService
	if rf, ok := ret.Get(0).(func(context.Context, types.AccAddress, *tenderminttypes.Header) kava.AccountBalanceService); ok {
		r0 = rf(ctx, addr, blockHeader)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(kava.AccountBalanceService)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, types.AccAddress, *tenderminttypes.Header) error); ok {
		r1 = rf(ctx, addr, blockHeader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewBalanceServiceFactoryT interface {
	mock.TestingT
	Cleanup(func())
}

// NewBalanceServiceFactory creates a new instance of BalanceServiceFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBalanceServiceFactory(t NewBalanceServiceFactoryT) *BalanceServiceFactory {
	mock := &BalanceServiceFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
