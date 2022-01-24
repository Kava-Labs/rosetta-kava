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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/tendermint/tendermint/crypto"
)

var requiredMetadata = []string{
	"signers",
	"gas_price",
	"gas_wanted",
	"memo",
}

type metadata struct {
	signers   []signerInfo
	gasWanted uint64
	gasPrice  float64
	memo      string
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	metadata, err := validateAndParseMetadata(request.Metadata)
	if err != nil {
		return nil, wrapErr(ErrInvalidMetadata, err)
	}

	txBuilder := s.encodingConfig.TxConfig.NewTxBuilder()

	msgs, rerr := parseOperationMsgs(request.Operations)
	if rerr != nil {
		return nil, rerr
	}
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	feeAmount := sdk.NewInt(int64(math.Ceil(metadata.gasPrice * float64(metadata.gasWanted))))
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("ukava", feeAmount)))
	txBuilder.SetGasLimit(metadata.gasWanted)
	txBuilder.SetMemo(metadata.memo)

	tx := txBuilder.GetTx()

	var sigsV2 []signing.SignatureV2
	for i, signer := range metadata.signers {
		if i >= len(request.PublicKeys) {
			return nil, ErrMissingPublicKey
		}
		// TODO: validate curve type
		pubKey := secp256k1.PubKey{Key: request.PublicKeys[i].Bytes}

		signatureData := signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		}
		sigV2 := signing.SignatureV2{
			PubKey:   &pubKey,
			Data:     &signatureData,
			Sequence: signer.AccountSequence,
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	txBytes, err := s.encodingConfig.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}
	payloads := []*types.SigningPayload{}
	for i, signer := range metadata.signers {
		addr, rerr := getAddressFromPublicKey(request.PublicKeys[i])
		if rerr != nil {
			return nil, rerr
		}

		signerData := authsigning.SignerData{
			ChainID:       request.NetworkIdentifier.Network,
			AccountNumber: signer.AccountNumber,
			Sequence:      signer.AccountSequence,
		}

		signBytes, err := s.encodingConfig.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, tx)
		if err != nil {
			return nil, wrapErr(ErrInvalidTx, err)
		}

		payloads = append(payloads, &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{Address: addr.String()},
			Bytes:             crypto.Sha256(signBytes),
			SignatureType:     types.Ecdsa,
		})
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads:            payloads,
	}, nil
}

func validateAndParseMetadata(meta map[string]interface{}) (*metadata, error) {
	for _, name := range requiredMetadata {
		if _, ok := meta[name]; !ok {
			return nil, fmt.Errorf("no %s provided", name)
		}
	}

	rawSigners, ok := meta["signers"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "signers")
	}
	var signers []signerInfo
	err := json.Unmarshal([]byte(rawSigners), &signers)
	if err != nil {
		return nil, fmt.Errorf("invalid value for signers: %w", err)
	}

	gasPrice, ok := meta["gas_price"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "gas_price")
	}

	gasWanted, ok := meta["gas_wanted"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "gas_wanted")
	}

	memo, ok := meta["memo"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "memo")
	}

	return &metadata{
		signers:   signers,
		gasPrice:  gasPrice,
		gasWanted: uint64(gasWanted),
		memo:      memo,
	}, nil
}
