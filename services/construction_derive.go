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
	"github.com/coinbase/rosetta-sdk-go/errors"
	"github.com/coinbase/rosetta-sdk-go/types"
)

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructionAPIService) ConstructionDerive(ctx context.Context, request *types.ConstructionDeriveRequest) (*types.ConstructionDeriveResponse, *types.Error) {
	account, err := on.client.AccountIdentifierFromPublicKey(request.PublicKey)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}
	return &types.ConstructionDeriveResponse{
		AccountIdentifier: account,
		Metadata:          nil,
	}, nil
}