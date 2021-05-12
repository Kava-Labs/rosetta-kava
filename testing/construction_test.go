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
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTransferOperations(from string, to string, amount string) []*types.Operation {
	return []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                "transfer",
			Account:             &types.AccountIdentifier{Address: from},
			Amount:              &types.Amount{Value: amount, Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			RelatedOperations:   []*types.OperationIdentifier{{Index: 0}},
			Type:                "transfer",
			Account:             &types.AccountIdentifier{Address: from},
			Amount:              &types.Amount{Value: amount, Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
	}
}

func TestTransfer(t *testing.T) {
	ctx := context.Background()
	cdc := app.MakeCodec()

	// TODO: use /construction/dervice to generate bech32 addresses
	operations := createTransferOperations(
		"kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea", // from
		"kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w", // to
		"1000001", // ukava amount
	)

	suggestedFeeMultipler := float64(0.5)

	preprocessResponse, rosettaErr, err := client.ConstructionAPI.ConstructionPreprocess(ctx,
		&types.ConstructionPreprocessRequest{
			NetworkIdentifier:      config.NetworkIdentifier,
			Operations:             operations,
			MaxFee:                 []*types.Amount{{Value: "500001", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}}},
			SuggestedFeeMultiplier: &suggestedFeeMultipler,
			Metadata:               map[string]interface{}{"memo": "test transfer integration"},
		},
	)

	require.Nil(t, rosettaErr)
	require.NoError(t, err)

	actualSuggestedFeeMultiplier, ok := preprocessResponse.Options["suggested_fee_multiplier"].(float64)
	assert.True(t, ok)
	assert.InDelta(t, suggestedFeeMultipler, actualSuggestedFeeMultiplier, 0.0000001)

	actualMemo, ok := preprocessResponse.Options["memo"].(string)
	assert.True(t, ok)
	assert.Equal(t, "test transfer integration", actualMemo)

	maxFeeEncoded, ok := preprocessResponse.Options["max_fee"].(string)
	assert.True(t, ok)

	var maxFee sdk.Coins
	err = cdc.UnmarshalJSON([]byte(maxFeeEncoded), &maxFee)
	require.NoError(t, err)
	assert.Equal(t, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(500001))), maxFee)
}
