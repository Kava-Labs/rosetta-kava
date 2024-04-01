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

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cometbft/cometbft/crypto/secp256k1"
)

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructionAPIService) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	addr, err := getAddressFromPublicKey(request.PublicKey)
	if err != nil {
		return nil, err
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: addr.String(),
		},
	}, nil
}

// TODO: use cosmos-sdk/crypto/keys instead of tendermint?
func parsePublicKey(pubKey *types.PublicKey) (secp256k1.PubKey, *types.Error) {
	tmPubKey := make([]byte, secp256k1.PubKeySize)

	if pubKey.CurveType != types.Secp256k1 {
		return tmPubKey, ErrUnsupportedCurveType
	}

	if len(pubKey.Bytes) == 0 {
		return tmPubKey, wrapErr(ErrPublicKeyNil, errors.New("nil public key"))
	}

	pk, err := btcec.ParsePubKey(pubKey.Bytes)
	if err != nil {
		return tmPubKey, wrapErr(ErrInvalidPublicKey, err)
	}

	copy(tmPubKey[:], pk.SerializeCompressed())

	return tmPubKey, nil
}

func getAddressFromPublicKey(pubKey *types.PublicKey) (sdk.AccAddress, *types.Error) {
	tmPubKey, err := parsePublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	return sdk.AccAddress(tmPubKey.Address().Bytes()), nil
}
