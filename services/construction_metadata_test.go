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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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
	servicer, _ := setupConstructionAPIServicer()
	servicer.config.Mode = configuration.Online

	requiredOptions := map[string]interface{}{
		"tx_body":                  "{\"messages\": [], \"memo\": \"some memo message\"}",
		"gas_adjustment":           float64(0.2),
		"suggested_fee_multiplier": float64(0),
	}

	for key := range requiredOptions {
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
	encodingConfig := app.MakeEncodingConfig()
	cdc := encodingConfig.Marshaler
	servicer, _ := setupConstructionAPIServicer()
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
		&banktypes.MsgSend{
			FromAddress: fromAddr.String(),
			ToAddress:   toAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", coinAmount)),
		},
	}

	txBody := tx.TxBody{}
	anys, err := convertMsgsToAnys(msgs)
	require.NoError(t, err)
	txBody.Messages = anys
	txBody.Memo = "some memo message"
	encodedTxBody, err := cdc.MarshalJSON(&txBody)
	require.NoError(t, err)

	maxFee := sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))
	encodedMaxFee, err := json.Marshal(maxFee)
	require.NoError(t, err)

	validOptions := map[string]interface{}{
		"tx_body":                  string(encodedTxBody),
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
			name:    "tx_body not a string encoding",
			key:     "tx_body",
			value:   []byte{},
			message: "invalid value for tx_body",
		},
		{
			name:    "tx_body not correct object",
			key:     "tx_body",
			value:   `[{"foo":"bar"}]`,
			message: "invalid value for tx_body",
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

func TestConstructionMetadata_GasAndFee(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	cdc := encodingConfig.Marshaler

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
		&banktypes.MsgSend{
			FromAddress: fromAddr.String(),
			ToAddress:   toAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", coinAmount)),
		},
	}

	testCases := []struct {
		name                   string
		estimatedGas           uint64
		gasAdjustment          float64
		suggestedFeeMultiplier float64
		maxFee                 sdk.Coins
		expectedGasWanted      uint64
		expectedGasPrice       float64
		expectedFeeAmount      sdk.Int
	}{
		{
			name:                   "zero multiplier",
			gasAdjustment:          0.2,
			suggestedFeeMultiplier: 0,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      85000,
			expectedGasPrice:       float64(0.001),
			expectedFeeAmount:      sdk.NewInt(85),
		},
		{
			name:                   "small multiplier",
			gasAdjustment:          0.5,
			suggestedFeeMultiplier: 0.00001,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      100001,
			expectedGasPrice:       float64(0.00100004),
			expectedFeeAmount:      sdk.NewInt(101),
		},
		{
			name:                   "multiplier under 1",
			gasAdjustment:          0.5,
			suggestedFeeMultiplier: 0.5,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      200000,
			expectedGasPrice:       float64(0.003),
			expectedFeeAmount:      sdk.NewInt(600),
		},
		{
			name:                   "multiplier equal to 1",
			gasAdjustment:          0,
			suggestedFeeMultiplier: 1,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      200000,
			expectedGasPrice:       float64(0.005),
			expectedFeeAmount:      sdk.NewInt(1000),
		},
		{
			name:                   "suggested fee is rounded up",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 1,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      200001,
			expectedGasPrice:       float64(0.005),
			expectedFeeAmount:      sdk.NewInt(1001),
		},
		{
			name:                   "multiplier below 2",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 1.6,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      200001,
			expectedGasPrice:       float64(0.032),
			expectedFeeAmount:      sdk.NewInt(6401),
		},
		{
			name:                   "multiplier equal to 2",
			gasAdjustment:          0.8,
			suggestedFeeMultiplier: 2,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      500001,
			expectedGasPrice:       float64(0.05),
			expectedFeeAmount:      sdk.NewInt(25001),
		},
		{
			name:                   "multiplier below 3",
			gasAdjustment:          0.8,
			suggestedFeeMultiplier: 2.1,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      400001,
			expectedGasPrice:       float64(0.07),
			expectedFeeAmount:      sdk.NewInt(28001),
		},
		{
			name:                   "multiplier equal to 3",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      100000,
			expectedGasPrice:       float64(0.25),
			expectedFeeAmount:      sdk.NewInt(25000),
		},
		{
			name:                   "multiplier over 3",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3.9,
			maxFee:                 sdk.Coins{},
			expectedGasWanted:      100000,
			expectedGasPrice:       float64(0.25),
			expectedFeeAmount:      sdk.NewInt(25000),
		},
		{
			name:                   "max fee greater than suggested fee",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3.9,
			maxFee:                 sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(30000))),
			expectedGasWanted:      100000,
			expectedGasPrice:       float64(0.25),
			expectedFeeAmount:      sdk.NewInt(25000),
		},
		{
			name:                   "max fee equal to suggested fee",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3.9,
			maxFee:                 sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(25000))),
			expectedGasWanted:      100000,
			expectedGasPrice:       float64(0.25),
			expectedFeeAmount:      sdk.NewInt(25000),
		},
		{
			name:                   "max fee less than suggested fee",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3.9,
			maxFee:                 sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20000))),
			expectedGasWanted:      100000,
			expectedGasPrice:       float64(0.2),
			expectedFeeAmount:      sdk.NewInt(20000),
		},
		{
			name:                   "max fee less than suggested fee, gas price capped",
			gasAdjustment:          0.1,
			suggestedFeeMultiplier: 3.9,
			maxFee:                 sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(20000))),
			expectedGasWanted:      100001,
			expectedGasPrice:       float64(0.19999800002),
			expectedFeeAmount:      sdk.NewInt(20000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			servicer, mockClient := setupConstructionAPIServicer()
			servicer.config.Mode = configuration.Online
			ctx := context.Background()

			txBody := tx.TxBody{}
			anys, err := convertMsgsToAnys(msgs)
			require.NoError(t, err)
			txBody.Messages = anys
			txBody.Memo = "some memo message"
			encodedTxBody, err := cdc.MarshalJSON(&txBody)
			require.NoError(t, err)

			txBuilder := encodingConfig.TxConfig.NewTxBuilder()
			err = txBuilder.SetMsgs(msgs...)
			require.NoError(t, err)
			txBuilder.SetMemo(txBody.Memo)
			err = txBuilder.SetSignatures([]signing.SignatureV2{}...)
			require.NoError(t, err)
			expectedTx := txBuilder.GetTx()

			encodedMaxFee, err := json.Marshal(tc.maxFee)
			require.NoError(t, err)

			validOptions := map[string]interface{}{
				"tx_body":                  string(encodedTxBody),
				"gas_adjustment":           tc.gasAdjustment,
				"suggested_fee_multiplier": tc.suggestedFeeMultiplier,
				"max_fee":                  string(encodedMaxFee),
			}

			request := &types.ConstructionMetadataRequest{
				Options:    validOptions,
				PublicKeys: []*types.PublicKey{},
			}

			kavaErr := errors.New("some kava error")
			mockClient.On("EstimateGas", ctx, expectedTx, tc.gasAdjustment).Return(uint64(0), kavaErr).Once()
			response, rerr := servicer.ConstructionMetadata(ctx, request)
			assert.Nil(t, response)
			require.NotNil(t, rerr)
			assert.Equal(t, wrapErr(ErrKava, kavaErr), rerr)

			mockClient.On("EstimateGas", ctx, expectedTx, tc.gasAdjustment).Return(tc.expectedGasWanted, nil).Once()
			response, rerr = servicer.ConstructionMetadata(ctx, request)
			require.Nil(t, rerr)

			gasWanted, ok := response.Metadata["gas_wanted"].(uint64)
			require.True(t, ok)

			gasPrice, ok := response.Metadata["gas_price"].(float64)
			require.True(t, ok)

			memo, ok := response.Metadata["memo"].(string)
			require.True(t, ok)

			assert.Equal(t, "some memo message", memo)

			assert.Equal(t, tc.expectedGasWanted, gasWanted)
			assert.InDelta(t, tc.expectedGasPrice, gasPrice, 0.000000000001)

			require.Equal(t, 1, len(response.SuggestedFee))
			coin, rerr := amountToCoin(response.SuggestedFee[0])
			require.Nil(t, rerr)

			assert.Equal(t, "ukava", coin.Denom)
			assert.Equal(t, tc.expectedFeeAmount, coin.Amount)
		})
	}
}

func TestConstructionMetadata_SignerData(t *testing.T) {
	t.Skip()
	servicer, mockClient := setupConstructionAPIServicer()
	servicer.config.Mode = configuration.Online
	ctx := context.Background()

	encodingConfig := app.MakeEncodingConfig()
	cdc := encodingConfig.Marshaler

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
		&banktypes.MsgSend{
			FromAddress: fromAddr.String(),
			ToAddress:   toAddr.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", coinAmount)),
		},
	}

	txBody := tx.TxBody{}
	anys, err := convertMsgsToAnys(msgs)
	require.NoError(t, err)
	txBody.Messages = anys
	txBody.Memo = "some memo message"
	encodedTxBody, err := cdc.MarshalJSON(&txBody)
	require.NoError(t, err)

	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	txBuilder.SetMemo(txBody.Memo)
	expectedTx := txBuilder.GetTx()

	validOptions := map[string]interface{}{
		"tx_body":                  string(encodedTxBody),
		"gas_adjustment":           float64(0.1),
		"suggested_fee_multiplier": float64(1),
	}

	accountPubKey := "AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY"
	pubKeyBytes, err := base64.StdEncoding.DecodeString(accountPubKey)
	require.NoError(t, err)
	accountAddr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	request := &types.ConstructionMetadataRequest{
		Options: validOptions,
		PublicKeys: []*types.PublicKey{
			{
				Bytes:     pubKeyBytes,
				CurveType: types.Secp256k1,
			},
		},
	}

	mockClient.On("EstimateGas", ctx, expectedTx, float64(0.1)).Return(uint64(100000), nil).Once()

	accountErr := errors.New("some client error")
	mockClient.On("Account", ctx, accountAddr).Return(nil, accountErr).Once()

	response, rerr := servicer.ConstructionMetadata(ctx, request)
	assert.Nil(t, response)
	require.Equal(t, wrapErr(ErrKava, accountErr), rerr)

	account := &authtypes.BaseAccount{
		AccountNumber: 10,
		Sequence:      11,
	}

	mockClient.On("Account", ctx, accountAddr).Return(account, nil).Once()
	response, rerr = servicer.ConstructionMetadata(ctx, request)
	assert.Nil(t, rerr)

	signersRaw, ok := response.Metadata["signers"].(string)
	require.True(t, ok)
	var signers []signerInfo
	err = json.Unmarshal([]byte(signersRaw), &signers)
	require.NoError(t, err)

	require.Equal(t, 1, len(signers))
	signer := signers[0]
	assert.Equal(t, account.GetAccountNumber(), signer.AccountNumber)
	assert.Equal(t, account.GetSequence(), signer.AccountSequence)
}
