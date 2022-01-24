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

package kava

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrencies(t *testing.T) {
	// all currences have a denom mapping defined
	for expectedDenom, currency := range Currencies {
		denom, ok := Denoms[currency.Symbol]
		if !ok {
			t.Fatalf("Symbol %s missing from Denoms", currency.Symbol)
		}
		assert.Equal(t, expectedDenom, denom)
		assert.Equal(t, currency.Decimals, int32(6))
	}

	// all denoms have a currency mapping
	for symbol, denom := range Denoms {
		currency, ok := Currencies[denom]
		if !ok {
			t.Fatalf("Denom %s missing from Currencies", denom)
		}

		assert.Equal(t, symbol, currency.Symbol)
	}

	// bnb, busd, btcb, xrpb, hbtc, (atom direct from hub) not defined
	for _, denom := range []string{"bnb", "busd", "btcb", "xrpb", "hbtc", "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2"} {
		assert.Nil(t, Currencies[denom])
	}
}
