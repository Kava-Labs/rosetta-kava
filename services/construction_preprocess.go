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

	"github.com/coinbase/rosetta-sdk-go/types"
)

var defaultSuggestedFeeMultiplier = float64(1)

// ConstructionPreprocess implements the /construction/preprocess
// endpoint.
func (s *ConstructionAPIService) ConstructionPreprocess(
	ctx context.Context,
	request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	if len(request.Operations) == 0 {
		return nil, ErrNoOperations
	}

	if request.SuggestedFeeMultiplier == nil {
		request.SuggestedFeeMultiplier = &defaultSuggestedFeeMultiplier
	}

	var memo string
	if rawMemo, exists := request.Metadata["memo"]; exists {
		if parsedMemo, ok := rawMemo.(string); ok {
			memo = parsedMemo
		}
	}

	return &types.ConstructionPreprocessResponse{
		Options: map[string]interface{}{
			"suggested_fee_multiplier": *request.SuggestedFeeMultiplier,
			"memo":                     memo,
		},
	}, nil
}
