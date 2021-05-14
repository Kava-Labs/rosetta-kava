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
	"fmt"

	"github.com/kava-labs/rosetta-kava/configuration"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var requiredOptions = []string{
	"msgs",
	"memo",
	"gas_adjustment",
	"suggested_fee_multiplier",
}

type options struct {
	msgs                     []sdk.Msg
	memo                     string
	gas_adjustment           float64
	suggested_fee_multiplier float64
	max_fee                  sdk.Coins
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if s.config.Mode != configuration.Online {
		return nil, ErrUnavailableOffline
	}

	for _, option := range requiredOptions {
		if _, ok := request.Options[option]; !ok {
			return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("no %s provided", option))
		}
	}

	rawMsgs, ok := request.Options["msgs"].(string)
	if !ok {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "msgs"))
	}

	var msgs []sdk.Msg
	err := s.cdc.UnmarshalJSON([]byte(rawMsgs), &msgs)
	if err != nil {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "msgs"))
	}

	_, ok = request.Options["memo"].(string)
	if !ok {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "memo"))
	}

	_, ok = request.Options["gas_adjustment"].(float64)
	if !ok {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "gas_adjustment"))
	}

	_, ok = request.Options["suggested_fee_multiplier"].(float64)
	if !ok {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "suggested_fee_multiplier"))
	}

	rawMaxFee, ok := request.Options["max_fee"].(string)
	if !ok {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "max_fee"))
	}

	var maxFee sdk.Coins
	err = s.cdc.UnmarshalJSON([]byte(rawMaxFee), &maxFee)
	if err != nil {
		return nil, wrapErr(ErrInvalidOptions, fmt.Errorf("invalid value for %s", "max_fee"))
	}

	return nil, nil
}
