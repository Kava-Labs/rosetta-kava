// Code generated by mockery 2.7.4. DO NOT EDIT.

package mocks

import (
	cosmos_sdktypes "github.com/cosmos/cosmos-sdk/types"

	mock "github.com/stretchr/testify/mock"

	types "github.com/coinbase/rosetta-sdk-go/types"
)

// AccountBalanceService is an autogenerated mock type for the AccountBalanceService type
type AccountBalanceService struct {
	mock.Mock
}

// GetCoinsForSubAccount provides a mock function with given fields: subAccount
func (_m *AccountBalanceService) GetCoinsForSubAccount(subAccount *types.SubAccountIdentifier) (cosmos_sdktypes.Coins, error) {
	ret := _m.Called(subAccount)

	var r0 cosmos_sdktypes.Coins
	if rf, ok := ret.Get(0).(func(*types.SubAccountIdentifier) cosmos_sdktypes.Coins); ok {
		r0 = rf(subAccount)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cosmos_sdktypes.Coins)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*types.SubAccountIdentifier) error); ok {
		r1 = rf(subAccount)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}