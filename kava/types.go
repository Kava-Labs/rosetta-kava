// Copyright 2021 Kava Labs, Inc.
// Copyright 2020 Coinbase, Inc.
//
// Derived from github.com/coinbase/rosetta-ethereum@f81889b //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kava

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmclient "github.com/cometbft/cometbft/rpc/client"
)

const (
	// NodeVersion is the version of kava we are using
	NodeVersion = "v0.16.1"
	// Blockchain is always Kava
	Blockchain = "Kava"
	// HistoricalBalanceSupported is whether historical balance is supported.
	HistoricalBalanceSupported = true
	// IncludeMempoolCoins does not apply to rosetta-kava as it is not UTXO-based.
	IncludeMempoolCoins = false

	// SuccessStatus is the status of any
	// Kava operation considered successful.
	SuccessStatus = "success"
	// FailureStatus is the status of any
	// Kava operation considered unsuccessful.
	FailureStatus = "failure"

	// FeeOpType is used to reference fee operations
	FeeOpType = "fee"
	// TransferOpType is used to reference transfer operations
	TransferOpType = "transfer"
	// MintOpType is used to reference mint operations
	MintOpType = "mint"
	// BurnOpType is used to reference burn operations
	BurnOpType = "burn"

	// AccLiquid represents spendable coins
	AccLiquid = "liquid"
	// AccLiquidDelegated represents delgated spendable coins
	AccLiquidDelegated = "liquid_delegated"
	// AccLiquidUnbonding represents unbonding spendable coins
	AccLiquidUnbonding = "liquid_unbonding"
	// AccVesting represents vesting (non-spendable) coins
	AccVesting = "vesting"
	// AccVestingDelegated represents vesting coins that are delegated
	AccVestingDelegated = "vesting_delegated"
	// AccVestingUnbonding represents vesting coins that are unbonding
	AccVestingUnbonding = "vesting_unbonding"
)

var (
	// OperationTypes are all suppoorted operation types.
	OperationTypes = []string{
		FeeOpType,
		TransferOpType,
		MintOpType,
		BurnOpType,
	}

	// OperationStatuses are all supported operation statuses.
	OperationStatuses = []*types.OperationStatus{
		{
			Status:     SuccessStatus,
			Successful: true,
		},
		{
			Status:     FailureStatus,
			Successful: false,
		},
	}

	// CallMethods are all supported call methods.
	CallMethods = []string{}

	// BalanceExemptions lists sub-accounts that are balance exempt
	BalanceExemptions = []*types.BalanceExemption{
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccLiquid),
			ExemptionType:     types.BalanceDynamic,
		},
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccVesting),
			ExemptionType:     types.BalanceDynamic,
		},
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccLiquidDelegated),
			ExemptionType:     types.BalanceDynamic,
		},
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccVestingDelegated),
			ExemptionType:     types.BalanceDynamic,
		},
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccLiquidUnbonding),
			ExemptionType:     types.BalanceDynamic,
		},
		&types.BalanceExemption{
			SubAccountAddress: strToPtr(AccVestingUnbonding),
			ExemptionType:     types.BalanceDynamic,
		},
	}
)

// Currencies represents supported kava denom to rosetta currencies
var Currencies = map[string]*types.Currency{
	"ukava": &types.Currency{
		Symbol:   "KAVA",
		Decimals: 6,
	},
	"hard": &types.Currency{
		Symbol:   "HARD",
		Decimals: 6,
	},
	"swp": &types.Currency{
		Symbol:   "SWP",
		Decimals: 6,
	},
	"usdx": &types.Currency{
		Symbol:   "USDX",
		Decimals: 6,
	},
}

// Denoms represents rosetta symbol to kava denom conversion
var Denoms = map[string]string{
	"KAVA": "ukava",
	"HARD": "hard",
	"SWP":  "swp",
	"USDX": "usdx",
}

// RPCClient represents a tendermint http client with ability to get block by hash
type RPCClient interface {
	tmclient.Client

	Account(ctx context.Context, addr sdk.AccAddress, height int64) (authtypes.AccountI, error)
	Balance(ctx context.Context, addr sdk.AccAddress, height int64) (sdk.Coins, error)
	Delegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.DelegationResponses, error)
	UnbondingDelegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.UnbondingDelegations, error)
	SimulateTx(ctx context.Context, tx authsigning.Tx) (*sdk.SimulationResponse, error)
}

func strToPtr(s string) *string {
	return &s
}
