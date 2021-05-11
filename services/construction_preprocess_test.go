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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContructionAPIServicer() *ConstructionAPIService {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	return NewConstructionAPIService(cfg, mockClient)
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
