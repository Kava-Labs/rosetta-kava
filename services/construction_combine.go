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
	"github.com/cosmos/cosmos-sdk/x/auth"
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

	var tx auth.StdTx
	err = s.cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	for _, signature := range request.Signatures {
		pubkey, err := parsePublicKey(signature.PublicKey)
		if err != nil {
			return nil, err
		}

		tx.Signatures = append(tx.Signatures, auth.StdSignature{
			PubKey:    pubkey,
			Signature: signature.Bytes,
		})
	}

	signedTxBytes, err := s.cdc.MarshalBinaryLengthPrefixed(tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedTxBytes),
	}, nil
}
