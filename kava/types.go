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
)

const (
	// NodeVersion is the version of kvd we are using
	NodeVersion = "v0.14.1"

	// Blockchain is always Kava
	Blockchain = "Kava"

	// HistoricalBalanceSupported is whether historical balance is supported.
	HistoricalBalanceSupported = true

	// SuccessStatus is the status of any
	// Kava operation considered successful.
	SuccessStatus = "SUCCESS"

	// FailureStatus is the status of any
	// Kava operation considered unsuccessful.
	FailureStatus = "FAILURE"

	// IncludeMempoolCoins does not apply to rosetta-kava as it is not UTXO-based.
	IncludeMempoolCoins = false
)

var (
	// OperationTypes are all suppoorted operation types.
	OperationTypes = []string{
		"noop", // TODO: temp to satisfy asserter until we support operations
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
