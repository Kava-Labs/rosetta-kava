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
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// ConstructionHash implements the /construction/hash endpoint.
func (s *ConstructionAPIService) ConstructionHash(
	ctx context.Context,
	request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	bz, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	hash := sha256.Sum256(bz)
	bzHash := hash[:]
	hashString := hex.EncodeToString(bzHash)

	txIdentifier := &types.TransactionIdentifier{Hash: strings.ToUpper(hashString)}
	return &types.TransactionIdentifierResponse{TransactionIdentifier: txIdentifier}, nil
}
