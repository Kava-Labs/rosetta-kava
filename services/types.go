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

package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// Client is used services to get blockchain
// data and submit transactions.
type Client interface {
	Status(context.Context) (
		*types.BlockIdentifier,
		int64,
		*types.BlockIdentifier,
		*types.SyncStatus,
		[]*types.Peer,
		error,
	)

	Balance(
		context.Context,
		*types.AccountIdentifier,
		*types.PartialBlockIdentifier,
		[]*types.Currency,
	) (*types.AccountBalanceResponse, error)

	Block(
		context.Context,
		*types.PartialBlockIdentifier,
	) (*types.BlockResponse, error)

	PostTx(txBytes []byte) (
		*types.TransactionIdentifier,
		error,
	)
}
