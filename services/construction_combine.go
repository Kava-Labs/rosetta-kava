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

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// ConstructionCombine implements the /construction/combine endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	tx, err := s.encodingConfig.TxConfig.TxDecoder()(txBytes)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	txBuilder, err := s.encodingConfig.TxConfig.WrapTxBuilder(tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	sigsV2, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	for i, signature := range request.Signatures {
		if i >= len(sigsV2) {
			return nil, ErrMissingSignature
		}
		// TODO: verify public key is equal to the set sig v2 public key
		//tmpubkey, err := parsePublicKey(signature.PublicKey)
		//if err != nil {
		//return nil, err
		//}
		//pubkey := secp256k1.PubKey{Key: tmpubkey}

		sigsV2[i].Data = &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signature.Bytes,
		}
	}
	txBuilder.SetSignatures(sigsV2...)

	signedTxBytes, err := s.encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedTxBytes),
	}, nil
}
