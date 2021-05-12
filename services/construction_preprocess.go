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

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const defaultSuggestedFeeMultiplier = float64(1)

// ConstructionPreprocess implements the /construction/preprocess
// endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	if len(request.Operations) == 0 {
		return nil, ErrNoOperations
	}

	encodedMaxFee, err := getMaxFeeAndEncodeOption(request.MaxFee, s.cdc)
	if err != nil {
		return nil, err
	}

	options := map[string]interface{}{
		"suggested_fee_multiplier": suggestedMultiplerOrDefault(request.SuggestedFeeMultiplier),
		"memo":                     getMemoFromMetadata(request.Metadata),
	}
	if encodedMaxFee != nil {
		options["max_fee"] = *encodedMaxFee
	}

	return &types.ConstructionPreprocessResponse{
		Options: options,
	}, nil
}

func suggestedMultiplerOrDefault(multiplier *float64) float64 {
	if multiplier == nil {
		return defaultSuggestedFeeMultiplier
	}

	return *multiplier
}

func getMemoFromMetadata(metadata map[string]interface{}) string {
	if rawMemo, exists := metadata["memo"]; exists {
		if memo, ok := rawMemo.(string); ok {
			return memo
		}
	}

	return ""
}

func getMaxFeeAndEncodeOption(amounts []*types.Amount, cdc *codec.Codec) (*string, *types.Error) {
	if len(amounts) == 0 {
		return nil, nil
	}

	var maxFee sdk.Coins
	for _, feeAmount := range amounts {
		value, ok := sdk.NewIntFromString(feeAmount.Value)
		if !ok {
			return nil, ErrInvalidCurrencyAmount
		}

		denom, ok := kava.Denoms[feeAmount.Currency.Symbol]
		if !ok {
			return nil, ErrUnsupportedCurrency
		}

		currency, ok := kava.Currencies[denom]
		if !ok {
			return nil, ErrUnsupportedCurrency
		}

		if currency.Symbol != feeAmount.Currency.Symbol ||
			currency.Decimals != feeAmount.Currency.Decimals {
			return nil, ErrUnsupportedCurrency
		}

		maxFee = maxFee.Add(sdk.NewCoin(denom, value))
	}

	b, err := cdc.MarshalJSON(maxFee)
	if err != nil {
		return nil, wrapErr(ErrKava, err)
	}

	encodedMaxFee := string(b)
	return &encodedMaxFee, nil
}
