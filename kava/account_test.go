package kava_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kava-labs/rosetta-kava/kava"
	mocks "github.com/kava-labs/rosetta-kava/kava/mocks"

	sdkmath "cosmossdk.io/math"
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func setupFactory(t *testing.T, blockTime time.Time) (sdk.AccAddress, *tmtypes.Header, *mocks.RPCClient, kava.BalanceServiceFactory) {
	addr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	blockHeader := &tmtypes.Header{
		Height: 100,
		Time:   blockTime,
	}

	mockRPCClient := &mocks.RPCClient{}
	return addr, blockHeader, mockRPCClient, kava.NewRPCBalanceFactory(mockRPCClient)
}

func TestRPCAccountBalance_AccountError(t *testing.T) {
	ctx := context.Background()
	addr, blockHeader, mockRPCClient, serviceFactory := setupFactory(t, time.Now())

	accErr := errors.New("error retrieving account")
	mockRPCClient.On("Account", ctx, addr, blockHeader.Height).Return(nil, accErr)

	service, err := serviceFactory(ctx, addr, blockHeader)

	assert.Nil(t, service)
	assert.EqualError(t, err, accErr.Error())
}

func TestRPCAccountBalance_NullAccount(t *testing.T) {
	ctx := context.Background()
	addr, blockHeader, mockRPCClient, serviceFactory := setupFactory(t, time.Now())

	accErr := errors.New("unknown address kava1abc...")
	mockRPCClient.On("Account", ctx, addr, blockHeader.Height).Return(nil, accErr)

	service, err := serviceFactory(ctx, addr, blockHeader)
	assert.NoError(t, err)

	balance, sequence, err := service.GetCoinsAndSequenceForSubAccount(context.Background(), &types.SubAccountIdentifier{Address: kava.AccLiquid})
	assert.NoError(t, err)

	assert.Equal(t, sdk.Coins{}, balance)
	assert.Equal(t, uint64(0), sequence)
}

func TestRPCAccountBalance_BalanceError(t *testing.T) {
	ctx := context.Background()
	addr, blockHeader, mockRPCClient, serviceFactory := setupFactory(t, time.Now())

	balErr := errors.New("error retrieving balance")
	mockRPCClient.On("Account", ctx, addr, blockHeader.Height).Return(&authtypes.BaseAccount{}, nil)
	mockRPCClient.On("Balance", ctx, addr, blockHeader.Height).Return(nil, balErr)

	service, err := serviceFactory(ctx, addr, blockHeader)

	assert.Nil(t, service)
	assert.EqualError(t, err, balErr.Error())
}

func TestRPCAccountBalance_BaseAccount(t *testing.T) {
	coins := sdk.NewCoins(
		sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)},
		sdk.Coin{Denom: "hard", Amount: sdkmath.NewInt(500000)},
	)

	testCases := []struct {
		name           string
		subType        *types.SubAccountIdentifier
		baseCoins      sdk.Coins
		delegatedCoins []sdk.Coin
		delegatedErr   error
		unbondingCoins []sdk.Coin
		unbondingErr   error
		expectedCoins  sdk.Coins
		expectedErr    error
	}{
		{
			name:          "no subaccount returns owned coins",
			subType:       nil,
			baseCoins:     coins,
			expectedCoins: coins,
		},
		{
			name:          "liquid subaccount returns owned coins",
			subType:       &types.SubAccountIdentifier{Address: kava.AccLiquid},
			baseCoins:     coins,
			expectedCoins: coins,
		},
		{
			name:          "vesting subaccount returns zero coins",
			subType:       &types.SubAccountIdentifier{Address: kava.AccVesting},
			baseCoins:     coins,
			expectedCoins: sdk.Coins{},
		},
		{
			name:      "liquid delgated returns all delgated coins",
			subType:   &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins: coins,
			delegatedCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{
				{Denom: "ukava", Amount: sdkmath.NewInt(1500000)},
			},
		},
		{
			name:         "delegated coins rpc error",
			subType:      &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:    coins,
			delegatedErr: errors.New("some rpc error"),
			expectedErr:  errors.New("some rpc error"),
		},
		{
			name:      "vesting delegated coins returns zero",
			subType:   &types.SubAccountIdentifier{Address: kava.AccVestingDelegated},
			baseCoins: coins,
			delegatedCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{},
		},
		{
			name:      "liquid unbonding coins returns all unbonding tokens",
			subType:   &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins: coins,
			unbondingCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{
				{Denom: "ukava", Amount: sdkmath.NewInt(1500000)},
			},
		},
		{
			name:      "liquid unbonding coins returns all unbonding tokens",
			subType:   &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins: coins,
			unbondingCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{
				{Denom: "ukava", Amount: sdkmath.NewInt(1500000)},
			},
		},
		{
			name:         "unbonding coins rpc error",
			subType:      &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins:    coins,
			unbondingErr: errors.New("some rpc error"),
			expectedErr:  errors.New("some rpc error"),
		},
		{
			name:      "vesting ubonding coins returns zero",
			subType:   &types.SubAccountIdentifier{Address: kava.AccVestingUnbonding},
			baseCoins: coins,
			delegatedCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{},
		},
		{
			name:      "unknown sub type",
			subType:   &types.SubAccountIdentifier{Address: "unknown"},
			baseCoins: coins,
			delegatedCoins: []sdk.Coin{
				{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
			},
			expectedCoins: sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			addr, blockHeader, mockRPCClient, serviceFactory := setupFactory(t, time.Now())

			acc := &authtypes.BaseAccount{
				Address:  addr.String(),
				Sequence: 100,
			}

			mockRPCClient.On("Account", ctx, addr, blockHeader.Height).Return(acc, nil)
			mockRPCClient.On("Balance", ctx, addr, blockHeader.Height).Return(coins, nil)
			balanceService, err := serviceFactory(ctx, addr, blockHeader)
			require.NoError(t, err)

			delegations := stakingtypes.DelegationResponses{}
			for _, dc := range tc.delegatedCoins {
				delegations = append(delegations, stakingtypes.DelegationResponse{
					Balance: dc,
				})
			}

			if tc.delegatedErr == nil {
				mockRPCClient.On("Delegations", ctx, addr, blockHeader.Height).Return(delegations, nil)
			} else {
				mockRPCClient.On("Delegations", ctx, addr, blockHeader.Height).Return(nil, tc.delegatedErr)
			}

			unbondingDelegations := stakingtypes.UnbondingDelegations{}
			for _, dc := range tc.unbondingCoins {
				unbondingDelegations = append(unbondingDelegations, stakingtypes.UnbondingDelegation{
					Entries: []stakingtypes.UnbondingDelegationEntry{
						{
							Balance: dc.Amount,
						},
					},
				})
			}

			if tc.unbondingErr == nil {
				mockRPCClient.On("UnbondingDelegations", ctx, addr, blockHeader.Height).Return(unbondingDelegations, nil)
			} else {
				mockRPCClient.On("UnbondingDelegations", ctx, addr, blockHeader.Height).Return(nil, tc.unbondingErr)
			}

			coins, sequence, err := balanceService.GetCoinsAndSequenceForSubAccount(ctx, tc.subType)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCoins, coins)
				assert.Equal(t, acc.GetSequence(), sequence)
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Equal(t, (sdk.Coins)(nil), coins)
			}

		})
	}
}

func TestRPCAccountBalance_VestingAccount(t *testing.T) {
	// vest over a period of two hours with the current time
	// in the the middle
	vestingStartTime := time.Now().Add(-1 * time.Hour)
	vestingEndTime := vestingStartTime.Add(2 * time.Hour)

	// two vesting periods equally split, once vests are current time
	// and the other an hour after
	vestingPeriods := []vestingtypes.Period{
		{
			Length: 3600,
			Amount: sdk.NewCoins(
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				sdk.Coin{Denom: "hard", Amount: sdkmath.NewInt(250000)},
			),
		},
		{
			Length: 3600,
			Amount: sdk.NewCoins(
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
				sdk.Coin{Denom: "hard", Amount: sdkmath.NewInt(250000)},
			),
		},
	}

	// always the sum of vesting periods
	originalVesting := sdk.NewCoins()
	for _, vp := range vestingPeriods {
		originalVesting = originalVesting.Add(vp.Amount...)
	}

	// add some liquid coins to coins that are vesting
	baseCoins := originalVesting.Add(sdk.NewCoins(
		sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)},
		sdk.Coin{Denom: "hard", Amount: sdkmath.NewInt(250000)},
	)...)

	testCases := []struct {
		name      string
		subType   *types.SubAccountIdentifier
		baseCoins sdk.Coins

		originalVesting  sdk.Coins
		delegatedVesting sdk.Coins
		delegatedFree    sdk.Coins
		delegatedCoins   []sdk.Coin

		delegatedErr   error
		unbondingCoins []sdk.Coin
		unbondingErr   error
		expectedCoins  sdk.Coins
		expectedErr    error

		blockTime time.Time
	}{
		{
			name:            "no subaccount returns owned coins",
			subType:         nil,
			baseCoins:       baseCoins,
			originalVesting: originalVesting,
			expectedCoins:   baseCoins,
			blockTime:       time.Now(),
		},
		{
			name:            "liquid subaccount returns spendable coins",
			subType:         &types.SubAccountIdentifier{Address: kava.AccLiquid},
			baseCoins:       baseCoins,
			originalVesting: originalVesting,
			expectedCoins:   baseCoins.Sub(vestingPeriods[1].Amount...),
			blockTime:       time.Now(),
		},
		{
			name:            "vesting subaccount returns non-spendable coins",
			subType:         &types.SubAccountIdentifier{Address: kava.AccVesting},
			baseCoins:       baseCoins,
			originalVesting: originalVesting,
			expectedCoins:   vestingPeriods[1].Amount,
			blockTime:       time.Now(),
		},
		{
			name:            "liquid delegated - delgatedFree greater than unbonding, no vesting staked",
			subType:         &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:       baseCoins,
			originalVesting: originalVesting,
			delegatedFree:   sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "liquid delegated - delgatedFree greater than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.Coins{},
			blockTime:     time.Now(),
		},
		{
			name:             "liquid delegated - zero unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "vesting delegated - delgatedFree greater than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccVestingDelegated},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(750000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "vesting delegated - delgatedFree less than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccVestingDelegated},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "vesting delegated - zero unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccVestingDelegated},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(0)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "liquid unbonding - delgatedFree greater than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(750000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "liquid unbonding - delgatedFree less than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			blockTime:     time.Now(),
		},
		{
			name:             "vesting unbonding - delgatedFree greater than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccVestingUnbonding},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(500000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.Coins{},
			blockTime:     time.Now(),
		},
		{
			name:             "vesting unbonding - delgatedFree less than unbonding",
			subType:          &types.SubAccountIdentifier{Address: kava.AccVestingUnbonding},
			baseCoins:        baseCoins,
			originalVesting:  originalVesting,
			delegatedFree:    sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			delegatedVesting: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(1000000)}),
			unbondingCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			delegatedCoins: []sdk.Coin{
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
				sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)},
			},
			expectedCoins: sdk.NewCoins(sdk.Coin{Denom: "ukava", Amount: sdkmath.NewInt(250000)}),
			blockTime:     time.Now(),
		},
		{
			name:         "delegated coins rpc error",
			subType:      &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:    baseCoins,
			delegatedErr: errors.New("some rpc error"),
			expectedErr:  errors.New("some rpc error"),
		},
		{
			name:         "delegated coins unbonding rpc error",
			subType:      &types.SubAccountIdentifier{Address: kava.AccLiquidDelegated},
			baseCoins:    baseCoins,
			unbondingErr: errors.New("some rpc error"),
			expectedErr:  errors.New("some rpc error"),
		},
		{
			name:         "unbonding coins rpc error",
			subType:      &types.SubAccountIdentifier{Address: kava.AccLiquidUnbonding},
			baseCoins:    baseCoins,
			unbondingErr: errors.New("some rpc error"),
			expectedErr:  errors.New("some rpc error"),
		},
		{
			name:          "unknown sub type",
			subType:       &types.SubAccountIdentifier{Address: "unknown"},
			baseCoins:     baseCoins,
			expectedCoins: sdk.Coins{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			addr, blockHeader, mockRPCClient, serviceFactory := setupFactory(t, tc.blockTime)

			acc := &vestingtypes.PeriodicVestingAccount{
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount: &authtypes.BaseAccount{
						Address:  addr.String(),
						Sequence: 101,
					},
					OriginalVesting:  tc.originalVesting,
					DelegatedVesting: tc.delegatedVesting,
					DelegatedFree:    tc.delegatedFree,
					EndTime:          vestingEndTime.Unix(),
				},
				StartTime:      vestingStartTime.Unix(),
				VestingPeriods: vestingPeriods,
			}

			mockRPCClient.On("Account", ctx, addr, blockHeader.Height).Return(acc, nil)
			mockRPCClient.On("Balance", ctx, addr, blockHeader.Height).Return(tc.baseCoins, nil)
			balanceService, err := serviceFactory(ctx, addr, blockHeader)
			require.NoError(t, err)

			delegations := stakingtypes.DelegationResponses{}
			for _, dc := range tc.delegatedCoins {
				delegations = append(delegations, stakingtypes.DelegationResponse{
					Balance: dc,
				})
			}

			if tc.delegatedErr == nil {
				mockRPCClient.On("Delegations", ctx, addr, blockHeader.Height).Return(delegations, nil)
			} else {
				mockRPCClient.On("Delegations", ctx, addr, blockHeader.Height).Return(nil, tc.delegatedErr)
			}

			unbondingDelegations := stakingtypes.UnbondingDelegations{}
			for _, dc := range tc.unbondingCoins {
				unbondingDelegations = append(unbondingDelegations, stakingtypes.UnbondingDelegation{
					Entries: []stakingtypes.UnbondingDelegationEntry{
						{
							Balance: dc.Amount,
						},
					},
				})
			}

			if tc.unbondingErr == nil {
				mockRPCClient.On("UnbondingDelegations", ctx, addr, blockHeader.Height).Return(unbondingDelegations, nil)
			} else {
				mockRPCClient.On("UnbondingDelegations", ctx, addr, blockHeader.Height).Return(nil, tc.unbondingErr)
			}

			coins, sequence, err := balanceService.GetCoinsAndSequenceForSubAccount(ctx, tc.subType)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCoins, coins)
				assert.Equal(t, acc.GetSequence(), sequence)
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Equal(t, (sdk.Coins)(nil), coins)
			}
		})
	}
}
