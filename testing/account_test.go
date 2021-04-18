// +build integration
// Copyright 2021 Kava Labs, Inc.
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

package testing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountBalanceOffline(t *testing.T) {
	if config.Mode.String() == "online" {
		t.Skip("skipping account offline test")
	}

	ctx := context.Background()

	_, rosettaErr, err := client.AccountAPI.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAccountAddress,
		},
	})
	require.Error(t, err)
	require.NotNil(t, rosettaErr)

	err = asserter.Error(rosettaErr)
	require.NoError(t, err)

	assert.Equal(t, int32(1), rosettaErr.Code)
	assert.Equal(t, "Endpoint unavailable offline", rosettaErr.Message)
}

func TestAccountBalanceOnline(t *testing.T) {
	if config.Mode.String() == "offline" {
		t.Skip("skipping account online test")
	}

	ctx := context.Background()

	accountBalance, rosettaErr, err := client.AccountAPI.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAccountAddress,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, rosettaErr)

	err = asserter.AccountBalanceResponse(&types.PartialBlockIdentifier{Index: &accountBalance.BlockIdentifier.Index}, accountBalance)
	require.NoError(t, err)

	accountBalanceByIndex, rosettaErr, err := client.AccountAPI.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAccountAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &accountBalance.BlockIdentifier.Index,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, rosettaErr)
	require.Equal(t, accountBalance, accountBalanceByIndex)

	accountBalanceByHash, rosettaErr, err := client.AccountAPI.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		AccountIdentifier: &types.AccountIdentifier{
			Address: testAccountAddress,
		},
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: &accountBalance.BlockIdentifier.Hash,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, rosettaErr)
	require.Equal(t, accountBalance, accountBalanceByHash)

	// TODO: return height and fetch block time
	// to pass to spendable coins
	account, err := GetAccount(testAccountAddress)
	require.NoError(t, err)
	spendableCoins := account.SpendableCoins(time.Now())

	for _, amount := range accountBalance.Balances {
		rosettaSymbol := amount.Currency.Symbol
		assert.Equal(t, strings.ToUpper(rosettaSymbol), rosettaSymbol)

		kavaSymbol := strings.ToLower(rosettaSymbol)
		if kavaSymbol == "kava" {
			kavaSymbol = "ukava"
		}

		spendableAmount := spendableCoins.AmountOf(kavaSymbol)
		assert.Equal(t, amount.Value, spendableAmount)

		decimals := amount.Currency.Decimals
		if kavaSymbol == "ukava" || kavaSymbol == "hard" || kavaSymbol == "usdx" {
			assert.Equal(t, int32(6), decimals)
		} else {
			assert.Equal(t, int32(8), decimals)
		}
	}
}
