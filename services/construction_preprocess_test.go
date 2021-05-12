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

package services

import (
	"context"
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContructionAPIServicer() *ConstructionAPIService {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	cdc := app.MakeCodec()
	return NewConstructionAPIService(cfg, mockClient, cdc)
}

func float64ToPtr(value float64) *float64 {
	return &value
}

func strToPtr(value string) *string {
	return &value
}

func validConstructionPreprocessRequest() *types.ConstructionPreprocessRequest {
	defaultOps := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
			Amount:              &types.Amount{Value: "-5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			RelatedOperations:   []*types.OperationIdentifier{{Index: 0}},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"},
			Amount:              &types.Amount{Value: "5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
	}

	return &types.ConstructionPreprocessRequest{
		Operations: defaultOps,
		Metadata:   make(map[string]interface{}),
	}
}

func TestConstructionPreprocess_NoOperations(t *testing.T) {
	servicer := setupContructionAPIServicer()

	ctx := context.Background()
	response, err := servicer.ConstructionPreprocess(ctx,
		&types.ConstructionPreprocessRequest{},
	)
	assert.Nil(t, response)
	require.NotNil(t, err)

	assert.Equal(t, ErrNoOperations, err)

	ctx = context.Background()
	response, err = servicer.ConstructionPreprocess(ctx,
		&types.ConstructionPreprocessRequest{
			Operations: []*types.Operation{},
		},
	)
	assert.Nil(t, response)
	require.NotNil(t, err)

	assert.Equal(t, ErrNoOperations, err)
}

func TestConstructionPreprocess_SuggestedFeeMultiplier(t *testing.T) {
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		suggestedFeeMultiplier *float64
		expectedFeeMultiplier  float64
	}{
		{
			suggestedFeeMultiplier: nil,
			expectedFeeMultiplier:  1,
		},
		{
			suggestedFeeMultiplier: float64ToPtr(0.0000001),
			expectedFeeMultiplier:  0.0000001,
		},
		{
			suggestedFeeMultiplier: float64ToPtr(0.55),
			expectedFeeMultiplier:  0.55,
		},
		{
			suggestedFeeMultiplier: float64ToPtr(1),
			expectedFeeMultiplier:  1,
		},
		{
			suggestedFeeMultiplier: float64ToPtr(2),
			expectedFeeMultiplier:  2,
		},
		{
			suggestedFeeMultiplier: float64ToPtr(10),
			expectedFeeMultiplier:  10,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		request := validConstructionPreprocessRequest()

		request.SuggestedFeeMultiplier = tc.suggestedFeeMultiplier

		response, err := servicer.ConstructionPreprocess(ctx, request)
		require.Nil(t, err)

		actualMutliplier, ok := response.Options["suggested_fee_multiplier"].(float64)
		require.True(t, ok)

		assert.InDelta(t, tc.expectedFeeMultiplier, actualMutliplier, 0.0000001)
	}
}

func TestConstructionPreprocess_Memo(t *testing.T) {
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		memo *string
	}{
		{
			memo: nil,
		},
		{
			memo: strToPtr(""),
		},
		{
			memo: strToPtr("some memo for tx"),
		},
		{
			memo: strToPtr("a memo that is pretty long, longer than most memos"),
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		request := validConstructionPreprocessRequest()

		if tc.memo != nil {
			request.Metadata["memo"] = *tc.memo
		}

		response, err := servicer.ConstructionPreprocess(ctx, request)
		require.Nil(t, err)

		actualMemo, ok := response.Options["memo"].(string)
		require.True(t, ok)

		var expectedMemo string
		if tc.memo != nil {
			expectedMemo = *tc.memo
		} else {
			expectedMemo = ""
		}
		assert.Equal(t, expectedMemo, actualMemo)
	}
}

func TestConstructionPreprocess_MaxFee(t *testing.T) {
	cdc := app.MakeCodec()
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		maxFee         []*types.Amount
		expectedMaxFee sdk.Coins
		expectedErr    *types.Error
	}{
		{
			maxFee:         nil,
			expectedMaxFee: nil,
		},
		{
			maxFee:      []*types.Amount{{Value: "1.000", Currency: kava.Currencies["ukava"]}},
			expectedErr: ErrInvalidCurrencyAmount,
		},
		{
			maxFee:      []*types.Amount{{Value: "1000000", Currency: &types.Currency{Symbol: "BNB", Decimals: 8}}},
			expectedErr: ErrUnsupportedCurrency,
		},
		{
			maxFee:      []*types.Amount{{Value: "1000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 7}}},
			expectedErr: ErrUnsupportedCurrency,
		},
		{
			maxFee:         []*types.Amount{{Value: "1000000", Currency: kava.Currencies["ukava"]}},
			expectedMaxFee: sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(1000000), Denom: "ukava"}),
		},
		{
			maxFee:         []*types.Amount{{Value: "500000", Currency: kava.Currencies["ukava"]}},
			expectedMaxFee: sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(500000), Denom: "ukava"}),
		},
		{
			maxFee:         []*types.Amount{{Value: "600001", Currency: kava.Currencies["hard"]}},
			expectedMaxFee: sdk.NewCoins(sdk.Coin{Amount: sdk.NewInt(600001), Denom: "hard"}),
		},
		{
			maxFee: []*types.Amount{
				{Value: "100001", Currency: kava.Currencies["ukava"]},
				{Value: "200002", Currency: kava.Currencies["hard"]},
				{Value: "300003", Currency: kava.Currencies["usdx"]},
			},
			expectedMaxFee: sdk.NewCoins(
				sdk.Coin{Amount: sdk.NewInt(100001), Denom: "ukava"},
				sdk.Coin{Amount: sdk.NewInt(200002), Denom: "hard"},
				sdk.Coin{Amount: sdk.NewInt(300003), Denom: "usdx"},
			),
		},
	}

	for _, tc := range testCases {
		request := validConstructionPreprocessRequest()
		request.MaxFee = tc.maxFee

		ctx := context.Background()
		response, err := servicer.ConstructionPreprocess(ctx, request)

		if tc.expectedErr == nil {
			require.Nil(t, err)
		} else {
			assert.Nil(t, response)
			assert.Equal(t, tc.expectedErr, err)
			continue
		}

		if tc.expectedMaxFee != nil {
			actualMaxFee, ok := response.Options["max_fee"].(string)
			require.True(t, ok)

			var coins sdk.Coins
			err := cdc.UnmarshalJSON([]byte(actualMaxFee), &coins)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedMaxFee, coins)
		} else {
			assert.Nil(t, response.Options["max_fee"])
		}
	}
}

func TestConstructionPreprocess_UnclearOperations(t *testing.T) {
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		invalidOperations []*types.Operation
	}{
		{
			invalidOperations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                kava.TransferOpType,
					Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
					Amount:              &types.Amount{Value: "-1", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
			},
		},
		{
			invalidOperations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                kava.TransferOpType,
					Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
					Amount:              &types.Amount{Value: "-1", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					Type:                kava.TransferOpType,
					Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
					Amount:              &types.Amount{Value: "-1", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 2},
					Type:                kava.TransferOpType,
					Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
					Amount:              &types.Amount{Value: "-1", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
			},
		},
		{
			invalidOperations: []*types.Operation{
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 0},
					Type:                kava.TransferOpType,
					Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
					Amount:              &types.Amount{Value: "-10000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
				{
					OperationIdentifier: &types.OperationIdentifier{Index: 1},
					Type:                "not a transfer",
					Account:             &types.AccountIdentifier{Address: "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"},
					Amount:              &types.Amount{Value: "10000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
				},
			},
		},
	}

	for _, tc := range testCases {
		request := &types.ConstructionPreprocessRequest{
			Operations: tc.invalidOperations,
		}

		ctx := context.Background()
		response, err := servicer.ConstructionPreprocess(ctx, request)
		require.Nil(t, response)

		assert.Equal(t, ErrUnclearIntent.Code, err.Code)
		assert.Equal(t, ErrUnclearIntent.Message, err.Message)
	}
}

func TestConstructionPreprocess_TransferOperations(t *testing.T) {
	cdc := app.MakeCodec()
	servicer := setupContructionAPIServicer()

	fromAddress := "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"
	toAddress := "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"
	amount := "5000001"

	operations := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: fromAddress},
			Amount:              &types.Amount{Value: "-" + amount, Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: toAddress},
			Amount:              &types.Amount{Value: amount, Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
	}

	request := &types.ConstructionPreprocessRequest{
		Operations: operations,
	}

	ctx := context.Background()
	response, rerr := servicer.ConstructionPreprocess(ctx, request)
	require.Nil(t, rerr)

	encodedMsgs, ok := response.Options["msgs"].(string)
	require.True(t, ok)

	var msgs []sdk.Msg
	err := cdc.UnmarshalJSON([]byte(encodedMsgs), &msgs)
	require.NoError(t, err)

	fromAddr, err := sdk.AccAddressFromBech32(fromAddress)
	require.NoError(t, err)

	toAddr, err := sdk.AccAddressFromBech32(toAddress)
	require.NoError(t, err)

	coinAmount, ok := sdk.NewIntFromString(amount)
	require.True(t, ok)

	expectedMsgs := []sdk.Msg{
		bank.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   toAddr,
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", coinAmount)),
		},
	}

	assert.Equal(t, expectedMsgs, msgs)
	require.Equal(t, 1, len(response.RequiredPublicKeys))
	assert.Equal(t, "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea", response.RequiredPublicKeys[0].Address)
}
