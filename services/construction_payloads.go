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
	"fmt"
	"math"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
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

	msgs, rerr := parseOperationMsgs(request.Operations)
	if rerr != nil {
		return nil, rerr
	}

	feeAmount := sdk.NewInt(int64(math.Ceil(metadata.gasPrice * float64(metadata.gasWanted))))
	tx := auth.NewStdTx(
		msgs,
		auth.NewStdFee(metadata.gasWanted, sdk.NewCoins(sdk.NewCoin("ukava", feeAmount))),
		[]auth.StdSignature{},
		metadata.memo,
	)

	txBytes, err := s.cdc.MarshalBinaryLengthPrefixed(tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	payloads := []*types.SigningPayload{}
	for i, signer := range metadata.signers {
		if i >= len(request.PublicKeys) {
			return nil, ErrMissingPublicKey
		}

		addr, rerr := getAddressFromPublicKey(request.PublicKeys[0])
		if rerr != nil {
			return nil, rerr
		}

		signBytes := auth.StdSignBytes(
			request.NetworkIdentifier.Network,
			signer.accountNumber,
			signer.accountSequence,
			tx.Fee,
			tx.Msgs,
			tx.Memo,
		)

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

	signers, ok := meta["signers"].([]signerInfo)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "signers")
	}

	gasPrice, ok := meta["gas_price"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid value for %s", "gas_price")
	}

	gasWanted, ok := meta["gas_wanted"].(uint64)
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
		gasWanted: gasWanted,
		memo:      memo,
	}, nil
}
