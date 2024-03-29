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
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"

	sdkmath "cosmossdk.io/math"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var simPubKey = make([]byte, secp256k1.PubKeySize)

var requiredOptions = []string{
	"tx_body",
	"gas_adjustment",
	"suggested_fee_multiplier",
}

type options struct {
	txBody                 *tx.TxBody
	gasAdjustment          float64
	suggestedFeeMultiplier float64
	maxFee                 sdk.Coins
}

type signerInfo struct {
	AccountNumber   uint64 `json:"account_number"`
	AccountSequence uint64 `json:"account_sequence"`
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if s.config.Mode != configuration.Online {
		return nil, ErrUnavailableOffline
	}

	options, err := validateAndParseOptions(s.encodingConfig.Marshaler, request.Options)
	if err != nil {
		return nil, wrapErr(ErrInvalidOptions, err)
	}

	var msgs []sdk.Msg
	for _, any := range options.txBody.Messages {
		val := any.GetCachedValue()
		if val == nil {
			return nil, wrapErr(ErrKava, errors.New("error decoding messages"))
		}
		msg, ok := val.(sdk.Msg)
		if !ok {
			return nil, wrapErr(ErrKava, errors.New("error decoding messages"))
		}
		msgs = append(msgs, msg)
	}

	var signers []signerInfo
	var sigsV2 []signing.SignatureV2
	seenSigners := make(map[string]bool)
	for _, msg := range msgs {
		for _, signerAddr := range msg.GetSigners() {
			addr := signerAddr.String()

			if !seenSigners[addr] {
				seenSigners[addr] = true

				acc, err := s.client.Account(ctx, signerAddr)
				if err != nil {
					return nil, wrapErr(ErrKava, err)
				}

				signers = append(signers, signerInfo{
					AccountNumber:   acc.GetAccountNumber(),
					AccountSequence: acc.GetSequence(),
				})

				sdkpubkey := secp256k1.PubKey{Key: simPubKey}

				signatureData := signing.SingleSignatureData{
					SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
					Signature: nil,
				}
				sigV2 := signing.SignatureV2{
					PubKey:   &sdkpubkey,
					Data:     &signatureData,
					Sequence: acc.GetSequence(),
				}

				sigsV2 = append(sigsV2, sigV2)
			}
		}
	}

	encodedSigners, err := json.Marshal(signers)
	if err != nil {
		return nil, wrapErr(ErrKava, err)
	}

	txBuilder := s.encodingConfig.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}
	txBuilder.SetMemo(options.txBody.Memo)
	err = txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	tx := txBuilder.GetTx()

	gasWanted, err := s.client.EstimateGas(ctx, tx, options.gasAdjustment)
	if err != nil {
		return nil, wrapErr(ErrKava, err)
	}

	gasPrice := gasPriceFromMultiplier(options.suggestedFeeMultiplier)
	feeAmount := gasPrice * float64(gasWanted)
	suggestedFeeAmount := sdkmath.NewInt(int64(math.Ceil(feeAmount)))

	if !options.maxFee.Empty() && suggestedFeeAmount.GT(options.maxFee.AmountOf("ukava")) {
		suggestedFeeAmount = options.maxFee.AmountOf("ukava")
		gasPrice = float64(suggestedFeeAmount.Int64()) / float64(gasWanted)
	}

	return &types.ConstructionMetadataResponse{
		Metadata: map[string]interface{}{
			"signers":    string(encodedSigners),
			"gas_wanted": gasWanted,
			"gas_price":  gasPrice,
			"memo":       options.txBody.Memo,
		},
		SuggestedFee: []*types.Amount{
			{
				Value:    suggestedFeeAmount.String(),
				Currency: kava.Currencies["ukava"],
			},
		},
	}, nil
}

func validateAndParseOptions(cdc codec.Codec, opts map[string]interface{}) (*options, error) {
	for _, option := range requiredOptions {
		if _, ok := opts[option]; !ok {
			return nil, fmt.Errorf("no %s provided", option)
		}
	}

	rawTxBody, ok := opts["tx_body"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "tx_body")
	}

	var txBody tx.TxBody
	err := cdc.UnmarshalJSON([]byte(rawTxBody), &txBody)
	if err != nil {
		return nil, fmt.Errorf("invalid value for %s", "tx_body")
	}

	gasAdjustment, ok := opts["gas_adjustment"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "gas_adjustment")
	}

	suggestedFeeMultiplier, ok := opts["suggested_fee_multiplier"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "suggested_fee_multiplier")
	}

	var maxFee sdk.Coins
	if maxFeeOpt, ok := opts["max_fee"]; ok {
		rawMaxFee, ok := maxFeeOpt.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value for %s", "max_fee")
		}

		err = json.Unmarshal([]byte(rawMaxFee), &maxFee)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s", "max_fee")
		}
	}

	return &options{
		txBody:                 &txBody,
		gasAdjustment:          gasAdjustment,
		suggestedFeeMultiplier: suggestedFeeMultiplier,
		maxFee:                 maxFee,
	}, nil
}

func gasPriceFromMultiplier(multiplier float64) float64 {
	if multiplier <= 0 {
		return 0.001
	}

	if multiplier < 1 {
		return multiplier*0.004 + 0.001
	}

	if multiplier < 2 {
		return (multiplier-1)*0.045 + 0.005
	}

	if multiplier < 3 {
		return (multiplier-2)*0.2 + 0.05
	}

	return 0.25
}
