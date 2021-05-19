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
	"sort"
	"strings"
	"testing"

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
	require.Nil(t, rosettaErr)

	err = asserter.AccountBalanceResponse(&types.PartialBlockIdentifier{Index: &accountBalance.BlockIdentifier.Index}, accountBalance)
	require.NoError(t, err)

	sort.Slice(accountBalance.Balances, func(i, j int) bool {
		return accountBalance.Balances[i].Currency.Symbol < accountBalance.Balances[j].Currency.Symbol
	})

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
	require.Nil(t, rosettaErr)
	sort.Slice(accountBalanceByIndex.Balances, func(i, j int) bool {
		return accountBalanceByIndex.Balances[i].Currency.Symbol < accountBalanceByIndex.Balances[j].Currency.Symbol
	})
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
	require.Nil(t, rosettaErr)
	sort.Slice(accountBalanceByHash.Balances, func(i, j int) bool {
		return accountBalanceByHash.Balances[i].Currency.Symbol < accountBalanceByHash.Balances[j].Currency.Symbol
	})
	require.Equal(t, accountBalance, accountBalanceByHash)

	account, err := GetAccount(testAccountAddress, accountBalance.BlockIdentifier.Index)
	require.NoError(t, err)
	ownedCoins := account.GetCoins()

	for _, amount := range accountBalance.Balances {
		rosettaSymbol := amount.Currency.Symbol
		assert.Equal(t, strings.ToUpper(rosettaSymbol), rosettaSymbol)

		kavaSymbol := strings.ToLower(rosettaSymbol)
		if kavaSymbol == "kava" {
			kavaSymbol = "ukava"
		}

		ownedAmount := ownedCoins.AmountOf(kavaSymbol)
		assert.Equal(t, amount.Value, ownedAmount.String())

		decimals := amount.Currency.Decimals
		if kavaSymbol == "ukava" || kavaSymbol == "hard" || kavaSymbol == "usdx" {
			assert.Equal(t, int32(6), decimals)
		} else {
			assert.Equal(t, int32(8), decimals)
		}
	}
}
