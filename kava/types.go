// Copyright 2021 Kava Labs, Inc.
// Copyright 2020 Coinbase, Inc.
//
// Derived from github.com/coinbase/rosetta-ethereum@f81889b
//
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
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	tmclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	// NodeVersion is the version of kvd we are using
	NodeVersion = "v0.14.1"
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

	// AccDelegated represents delgated spendable coins
	AccLiquidDelegated = "liquid_delegated"
	// AccUnbonding represents unbonding spendable coins
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
	BalanceExemptions = []*types.BalanceExemption{}
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
	"usdx": &types.Currency{
		Symbol:   "USDX",
		Decimals: 6,
	},
}

// Denoms represents rosetta symbol to kava denom conversion
var Denoms = map[string]string{
	"KAVA": "ukava",
	"HARD": "hard",
	"USDX": "usdx",
}

// RPCClient represents a tendermint http client with ability to get block by hash
type RPCClient interface {
	tmclient.Client

	BlockByHash([]byte) (*ctypes.ResultBlock, error)
	Account(addr sdk.AccAddress, height int64) (authexported.Account, error)
}
