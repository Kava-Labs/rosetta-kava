package kava

import (
	"regexp"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	staking "github.com/cosmos/cosmos-sdk/x/staking"
	tmtypes "github.com/tendermint/tendermint/types"
)

const stakingDenom = "ukava"

var unknownAddress = regexp.MustCompile("unknown address")

// AccountBalanceService provides an interface fetch a balance from an account subtype
type AccountBalanceService interface {
	GetCoinsForSubAccount(
		subAccount *types.SubAccountIdentifier,
	) (sdk.Coins, error)
}

// BalanceServiceFactory provides an interface for creating a balance service for specifc a account and block
type BalanceServiceFactory func(addr sdk.AccAddress, blockHeader *tmtypes.Header) (AccountBalanceService, error)

// NewRPCBalanceFactory returns a balance service factory that uses an RPCClient to get an accounts balance
func NewRPCBalanceFactory(rpc RPCClient) BalanceServiceFactory {
	return func(addr sdk.AccAddress, blockHeader *tmtypes.Header) (AccountBalanceService, error) {
		acc, err := rpc.Account(addr, blockHeader.Height)
		if err != nil {
			if unknownAddress.MatchString(err.Error()) {
				return &nullBalance{}, nil
			}

			return nil, err
		}

		switch acc := acc.(type) {
		case vestingexported.VestingAccount:
			return &rpcVestingBalance{rpc: rpc, vacc: acc, blockHeader: blockHeader}, nil
		default:
			return &rpcBaseBalance{rpc: rpc, acc: acc, blockHeader: blockHeader}, nil
		}
	}
}

type nullBalance struct {
}

func (b *nullBalance) GetCoinsForSubAccount(subAccount *types.SubAccountIdentifier) (coins sdk.Coins, err error) {
	return sdk.Coins{}, nil
}

type rpcBaseBalance struct {
	rpc         RPCClient
	acc         authexported.Account
	blockHeader *tmtypes.Header
}

func (b *rpcBaseBalance) GetCoinsForSubAccount(subAccount *types.SubAccountIdentifier) (coins sdk.Coins, err error) {
	if subAccount == nil {
		coins = b.acc.GetCoins()
		return
	}

	switch subAccount.Address {
	case AccLiquid:
		coins = b.acc.SpendableCoins(b.blockHeader.Time)
	case AccLiquidDelegated:
		coins, err = b.totalDelegated()
	case AccLiquidUnbonding:
		coins, err = b.totalUnbondingDelegations()
	default:
		coins = sdk.Coins{}
	}

	return
}

func (b *rpcBaseBalance) totalDelegated() (sdk.Coins, error) {
	delegations, err := b.rpc.Delegations(b.acc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumDelegations(delegations), nil
}

func (b *rpcBaseBalance) totalUnbondingDelegations() (sdk.Coins, error) {
	unbondingDelegations, err := b.rpc.UnbondingDelegations(b.acc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumUnbondingDelegations(unbondingDelegations), nil
}

type rpcVestingBalance struct {
	rpc         RPCClient
	vacc        vestingexported.VestingAccount
	blockHeader *tmtypes.Header
}

func (b *rpcVestingBalance) GetCoinsForSubAccount(subAccount *types.SubAccountIdentifier) (coins sdk.Coins, err error) {
	if subAccount == nil {
		coins = b.vacc.GetCoins()
		return
	}

	switch subAccount.Address {
	case AccLiquid:
		coins = b.vacc.SpendableCoins(b.blockHeader.Time)
	case AccVesting:
		coins = b.vacc.GetVestingCoins(b.blockHeader.Time)
	case AccLiquidDelegated:
		coins, _, err = b.delegated()
	case AccVestingDelegated:
		_, coins, err = b.delegated()
	case AccLiquidUnbonding:
		coins, _, err = b.unbonding()
	case AccVestingUnbonding:
		_, coins, err = b.unbonding()
	default:
		coins = sdk.Coins{}
	}

	return
}

// delegated returns liquid and vesting coins that are staked
func (b *rpcVestingBalance) delegated() (sdk.Coins, sdk.Coins, error) {
	delegatedCoins, err := b.totalDelegated()
	if err != nil {
		return nil, nil, err
	}
	unbondingCoins, err := b.totalUnbondingDelegations()
	if err != nil {
		return nil, nil, err
	}

	delegated := delegatedCoins.AmountOf(stakingDenom)
	unbonding := unbondingCoins.AmountOf(stakingDenom)
	totalStaked := delegated.Add(unbonding)
	delegatedFree := b.vacc.GetDelegatedFree().AmountOf(stakingDenom)

	// total number of staked and unbonding tokens considered to be liquid
	totalFree := sdk.MinInt(totalStaked, delegatedFree)
	// any coins that are not considered liquid, are vesting up to a maximum of delegated
	stakedVesting := sdk.MinInt(totalStaked.Sub(totalFree), delegated)
	// staked free coins are left over
	stakedFree := delegated.Sub(stakedVesting)

	liquidCoins := sdk.NewCoins(newKavaCoin(stakedFree))
	vestingCoins := sdk.NewCoins(newKavaCoin(stakedVesting))
	return liquidCoins, vestingCoins, nil
}

// unbonding returns liquid and vesting coins that are unbonding
func (b *rpcVestingBalance) unbonding() (sdk.Coins, sdk.Coins, error) {
	unbondingCoins, err := b.totalUnbondingDelegations()
	if err != nil {
		return nil, nil, err
	}

	unbonding := unbondingCoins.AmountOf(stakingDenom)
	delegatedFree := b.vacc.GetDelegatedFree().AmountOf(stakingDenom)

	unbondingFree := sdk.MinInt(delegatedFree, unbonding)
	unbondingVesting := unbonding.Sub(unbondingFree)

	liquidCoins := sdk.NewCoins(newKavaCoin(unbondingFree))
	vestingCoins := sdk.NewCoins(newKavaCoin(unbondingVesting))
	return liquidCoins, vestingCoins, nil
}

func (b *rpcVestingBalance) totalDelegated() (sdk.Coins, error) {
	delegations, err := b.rpc.Delegations(b.vacc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumDelegations(delegations), nil
}

func (b *rpcVestingBalance) totalUnbondingDelegations() (sdk.Coins, error) {
	unbondingDelegations, err := b.rpc.UnbondingDelegations(b.vacc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumUnbondingDelegations(unbondingDelegations), nil
}

func sumDelegations(delegations staking.DelegationResponses) sdk.Coins {
	coins := sdk.Coins{}
	for _, d := range delegations {
		coins = coins.Add(d.Balance)
	}

	return coins
}

func sumUnbondingDelegations(unbondingDelegations staking.UnbondingDelegations) sdk.Coins {
	totalBalance := sdk.ZeroInt()
	for _, u := range unbondingDelegations {
		for _, e := range u.Entries {
			totalBalance = totalBalance.Add(e.Balance)
		}
	}

	if totalBalance.GT(sdk.ZeroInt()) {
		return sdk.NewCoins(newKavaCoin(totalBalance))
	}

	return sdk.Coins{}
}

func newKavaCoin(amount sdk.Int) sdk.Coin {
	return sdk.Coin{Denom: stakingDenom, Amount: amount}
}
