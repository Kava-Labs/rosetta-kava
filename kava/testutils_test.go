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

package kava_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/require"
)

func newTestAccount(t *testing.T) *authtypes.BaseAccount {
	addr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	return &authtypes.BaseAccount{
		Address: addr,
		Coins: sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(100)),
			sdk.NewCoin("hard", sdk.NewInt(200)),
			sdk.NewCoin("usdx", sdk.NewInt(300)),
			sdk.NewCoin("bnb", sdk.NewInt(10)),
			sdk.NewCoin("btcb", sdk.NewInt(1)),
			sdk.NewCoin("busd", sdk.NewInt(1000)),
		),
		AccountNumber: 2,
		Sequence:      5,
	}
}

func newEmptyTestAccount(t *testing.T) *authtypes.BaseAccount {
	addr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	return &authtypes.BaseAccount{
		Address:       addr,
		Coins:         sdk.NewCoins(),
		AccountNumber: 3,
		Sequence:      6,
	}
}

func newPartialTestAccount(t *testing.T) *authtypes.BaseAccount {
	addr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	return &authtypes.BaseAccount{
		Address: addr,
		Coins: sdk.NewCoins(
			sdk.NewCoin("hard", sdk.NewInt(10)),
		),
		AccountNumber: 4,
		Sequence:      7,
	}
}

func newTestVestingAccount(t *testing.T, endTime time.Time) *vestingtypes.DelayedVestingAccount {
	baseAccount := newTestAccount(t)
	vestingAccount := vestingtypes.NewDelayedVestingAccount(baseAccount, endTime.Unix())

	return vestingAccount
}

func getBalance(balances []*types.Amount, symbol string) *types.Amount {
	for _, balance := range balances {
		if balance.Currency.Symbol == symbol {
			return balance
		}
	}

	return nil
}

func generateDefaultCoins() sdk.Coins {
	denoms := []string{
		// native
		"ukava", "hard", "usdx",
		// not native
		"bnb", "busd", "btcb",
	}

	return generateCoins(denoms)
}

func generateCoins(denoms []string) sdk.Coins {
	coins := sdk.Coins{}

	for _, denom := range denoms {
		coins = append(coins, sdk.Coin{
			Denom:  denom,
			Amount: sdk.NewInt(int64(rand.Intn(1000 * 1e6))),
		})
	}

	return coins.Sort()
}
