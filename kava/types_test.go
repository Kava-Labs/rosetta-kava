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

	// bnb, busd, btcb, xrpb, hbtc not defined
	for _, denom := range []string{"bnb", "busd", "btcb", "xrpb", "hbtc"} {
		assert.Nil(t, Currencies[denom])
	}
}
