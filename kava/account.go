package kava

import (
	"context"
	"regexp"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingexported "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const stakingDenom = "ukava"

var unknownAddress = regexp.MustCompile("unknown address")

// AccountBalanceService provides an interface fetch a balance from an account subtype
type AccountBalanceService interface {
	GetCoinsAndSequenceForSubAccount(
		ctx context.Context,
		subAccount *types.SubAccountIdentifier,
	) (sdk.Coins, uint64, error)
}

// BalanceServiceFactory provides an interface for creating a balance service for specifc a account and block
type BalanceServiceFactory func(ctx context.Context, addr sdk.AccAddress, blockHeader *tmtypes.Header) (AccountBalanceService, error)

// NewRPCBalanceFactory returns a balance service factory that uses an RPCClient to get an accounts balance
func NewRPCBalanceFactory(rpc RPCClient) BalanceServiceFactory {
	return func(ctx context.Context, addr sdk.AccAddress, blockHeader *tmtypes.Header) (AccountBalanceService, error) {
		acc, err := rpc.Account(ctx, addr, blockHeader.Height)
		if err != nil {
			if unknownAddress.MatchString(err.Error()) {
				return &nullBalance{acc: acc}, nil
			}

			return nil, err
		}

		bal, err := rpc.Balance(ctx, addr, blockHeader.Height)
		if err != nil {
			return nil, err
		}

		switch acc := acc.(type) {
		case vestingexported.VestingAccount:
			return &rpcVestingBalance{rpc: rpc, vacc: acc, bal: bal, blockHeader: blockHeader}, nil
		default:
			return &rpcBaseBalance{rpc: rpc, acc: acc, bal: bal, blockHeader: blockHeader}, nil
		}
	}
}

type nullBalance struct {
	acc authtypes.AccountI
}

func (b *nullBalance) GetCoinsAndSequenceForSubAccount(ctx context.Context, subAccount *types.SubAccountIdentifier) (coins sdk.Coins, sequence uint64, err error) {
	if b.acc != nil {
		sequence = b.acc.GetSequence()
	}
	coins = sdk.Coins{}

	return
}

type rpcBaseBalance struct {
	rpc         RPCClient
	acc         authtypes.AccountI
	bal         sdk.Coins
	blockHeader *tmtypes.Header
}

func (b *rpcBaseBalance) GetCoinsAndSequenceForSubAccount(ctx context.Context, subAccount *types.SubAccountIdentifier) (coins sdk.Coins, sequence uint64, err error) {
	sequence = b.acc.GetSequence()

	if subAccount == nil {
		coins = b.bal
		return
	}

	switch subAccount.Address {
	case AccLiquid:
		coins = b.bal
	case AccLiquidDelegated:
		coins, err = b.totalDelegated(ctx)
	case AccLiquidUnbonding:
		coins, err = b.totalUnbondingDelegations(ctx)
	default:
		coins = sdk.Coins{}
	}

	return
}

func (b *rpcBaseBalance) totalDelegated(ctx context.Context) (sdk.Coins, error) {
	delegations, err := b.rpc.Delegations(ctx, b.acc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumDelegations(delegations), nil
}

func (b *rpcBaseBalance) totalUnbondingDelegations(ctx context.Context) (sdk.Coins, error) {
	unbondingDelegations, err := b.rpc.UnbondingDelegations(ctx, b.acc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumUnbondingDelegations(unbondingDelegations), nil
}

type rpcVestingBalance struct {
	rpc         RPCClient
	vacc        vestingexported.VestingAccount
	bal         sdk.Coins
	blockHeader *tmtypes.Header
}

func (b *rpcVestingBalance) GetCoinsAndSequenceForSubAccount(ctx context.Context, subAccount *types.SubAccountIdentifier) (coins sdk.Coins, sequence uint64, err error) {
	sequence = b.vacc.GetSequence()

	if subAccount == nil {
		coins = b.bal
		return
	}

	switch subAccount.Address {
	case AccLiquid:
		// TODO: this doesn't seem correct?  Can be negative??
		coins = b.bal.Sub(b.vacc.LockedCoins(b.blockHeader.Time))
		//coins = b.vacc.SpendableCoins(b.blockHeader.Time)
	case AccVesting:
		coins = b.vacc.GetVestingCoins(b.blockHeader.Time)
	case AccLiquidDelegated:
		coins, _, err = b.delegated(ctx)
	case AccVestingDelegated:
		_, coins, err = b.delegated(ctx)
	case AccLiquidUnbonding:
		coins, _, err = b.unbonding(ctx)
	case AccVestingUnbonding:
		_, coins, err = b.unbonding(ctx)
	default:
		coins = sdk.Coins{}
	}

	return
}

// delegated returns liquid and vesting coins that are staked
func (b *rpcVestingBalance) delegated(ctx context.Context) (sdk.Coins, sdk.Coins, error) {
	delegatedCoins, err := b.totalDelegated(ctx)
	if err != nil {
		return nil, nil, err
	}
	unbondingCoins, err := b.totalUnbondingDelegations(ctx)
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
func (b *rpcVestingBalance) unbonding(ctx context.Context) (sdk.Coins, sdk.Coins, error) {
	unbondingCoins, err := b.totalUnbondingDelegations(ctx)
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

func (b *rpcVestingBalance) totalDelegated(ctx context.Context) (sdk.Coins, error) {
	delegations, err := b.rpc.Delegations(ctx, b.vacc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumDelegations(delegations), nil
}

func (b *rpcVestingBalance) totalUnbondingDelegations(ctx context.Context) (sdk.Coins, error) {
	unbondingDelegations, err := b.rpc.UnbondingDelegations(ctx, b.vacc.GetAddress(), b.blockHeader.Height)
	if err != nil {
		return nil, err
	}

	return sumUnbondingDelegations(unbondingDelegations), nil
}

func sumDelegations(delegations stakingtypes.DelegationResponses) sdk.Coins {
	coins := sdk.Coins{}
	for _, d := range delegations {
		coins = coins.Add(d.Balance)
	}

	return coins
}

func sumUnbondingDelegations(unbondingDelegations stakingtypes.UnbondingDelegations) sdk.Coins {
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
