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
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructionAPIService) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	curveType := request.PublicKey.CurveType

	if curveType != types.Secp256k1 {
		return nil, ErrUnsupportedCurveType
	}

	if len(request.PublicKey.Bytes) == 0 {
		return nil, wrapErr(ErrPublicKeyNil, errors.New("nil public key"))
	}

	pubKey, err := btcec.ParsePubKey(request.PublicKey.Bytes, btcec.S256())
	if err != nil {
		return nil, wrapErr(ErrInvalidPublicKey, err)
	}

	var tmPubKey secp256k1.PubKeySecp256k1
	serializedPubKey := pubKey.SerializeCompressed()

	copy(tmPubKey[:], serializedPubKey)

	addressBytes := tmPubKey.Address().Bytes()
	accountAddress := sdk.AccAddress(addressBytes).String()

	response := &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: accountAddress,
		},
	}

	return response, nil
}
