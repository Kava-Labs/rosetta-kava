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
	"fmt"
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertOptionsErrorContext(t *testing.T, err *types.Error, context string) {
	require.NotNil(t, err)
	assert.Equal(t, ErrInvalidOptions.Code, err.Code)
	assert.Equal(t, ErrInvalidOptions.Message, err.Message)
	assert.Equal(t, context, err.Details["context"])
}

func TestConstructionMetadata_OptionsValidation_MissingFields(t *testing.T) {
	servicer := setupContructionAPIServicer()
	servicer.config.Mode = configuration.Online

	requiredOptions := map[string]interface{}{
		"msgs":                     "[]",
		"memo":                     "some memo message",
		"gas_adjustment":           float64(0.2),
		"suggested_fee_multiplier": float64(0),
	}

	for key, _ := range requiredOptions {
		ctx := context.Background()

		options := make(map[string]interface{})

		for k, v := range requiredOptions {
			if k != key {
				options[k] = v
			}
		}

		request := &types.ConstructionMetadataRequest{
			Options:    options,
			PublicKeys: []*types.PublicKey{},
		}

		response, err := servicer.ConstructionMetadata(ctx, request)
		assert.Nil(t, response)
		assertOptionsErrorContext(t, err, fmt.Sprintf("no %s provided", key))
	}
}

func TestConstructionMetadata_OptionsValidation_InvalidFields(t *testing.T) {
	cdc := app.MakeCodec()
	servicer := setupContructionAPIServicer()
	servicer.config.Mode = configuration.Online

	fromAddress := "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"
	toAddress := "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"
	amount := "5000001"
	fromAddr, err := sdk.AccAddressFromBech32(fromAddress)
	require.NoError(t, err)
	toAddr, err := sdk.AccAddressFromBech32(toAddress)
	require.NoError(t, err)
	coinAmount, ok := sdk.NewIntFromString(amount)
	require.True(t, ok)

	msgs := []sdk.Msg{
		bank.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   toAddr,
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", coinAmount)),
		},
	}
	encodedMsgs, err := cdc.MarshalJSON(msgs)
	require.NoError(t, err)

	maxFee := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))
	encodedMaxFee, err := cdc.MarshalJSON(maxFee)
	require.NoError(t, err)

	validOptions := map[string]interface{}{
		"msgs":                     string(encodedMsgs),
		"memo":                     "some memo message",
		"gas_adjustment":           float64(0.2),
		"suggested_fee_multiplier": float64(1.2),
		"max_fee":                  string(encodedMaxFee),
	}

	assertOptionError := func(key string, value interface{}, message string) {
		old, ok := validOptions[key]
		require.True(t, ok)
		validOptions[key] = value

		request := &types.ConstructionMetadataRequest{
			Options:    validOptions,
			PublicKeys: []*types.PublicKey{},
		}

		ctx := context.Background()
		resp, rerr := servicer.ConstructionMetadata(ctx, request)
		assert.Nil(t, resp)
		assertOptionsErrorContext(t, rerr, message)

		validOptions[key] = old
	}

	testCases := []struct {
		name    string
		key     string
		value   interface{}
		message string
	}{
		{
			name:    "msgs not a string encoding",
			key:     "msgs",
			value:   []byte{},
			message: "invalid value for msgs",
		},
		{
			name:    "msgs not correct object",
			key:     "msgs",
			value:   `[{"foo":"bar"}]`,
			message: "invalid value for msgs",
		},
		{
			name:    "memo not a string",
			key:     "memo",
			value:   []byte{},
			message: "invalid value for memo",
		},
		{
			name:    "gas adjustment not a float",
			key:     "gas_adjustment",
			value:   "1a",
			message: "invalid value for gas_adjustment",
		},
		{
			name:    "gas adjustment not a float",
			key:     "suggested_fee_multiplier",
			value:   "1a",
			message: "invalid value for suggested_fee_multiplier",
		},
		{
			name:    "max_fee not a string encoding",
			key:     "max_fee",
			value:   []byte{},
			message: "invalid value for max_fee",
		},
		{
			name:    "max_fee not correct object",
			key:     "max_fee",
			value:   `{}`,
			message: "invalid value for max_fee",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assertOptionError(tc.key, tc.value, tc.message)
		})
	}
}
