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

	"github.com/kava-labs/rosetta-kava/configuration"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// ConstructionAPIService implements the server.ConstructionAPIServicer interface.
type ConstructionAPIService struct {
	config *configuration.Configuration
	client Client
	cdc    *codec.Codec
}

// NewConstructionAPIService creates a new instance of a ConstructionAPIService.
func NewConstructionAPIService(
	cfg *configuration.Configuration,
	client Client,
	cdc *codec.Codec,
) *ConstructionAPIService {
	return &ConstructionAPIService{
		config: cfg,
		client: client,
		cdc:    cdc,
	}
}

// ConstructionMetadata implements the /construction/metadata endpoint.
func (s *ConstructionAPIService) ConstructionMetadata(
	ctx context.Context,
	request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	if s.config.Mode != configuration.Online {
		return nil, ErrUnavailableOffline
	}

	return nil, ErrUnimplemented
}

// ConstructionPayloads implements the /construction/payloads endpoint.
func (s *ConstructionAPIService) ConstructionPayloads(
	ctx context.Context,
	request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, ErrUnimplemented
}

// ConstructionCombine implements the /construction/combine
// endpoint.
func (s *ConstructionAPIService) ConstructionCombine(
	ctx context.Context,
	request *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	return nil, ErrUnimplemented
}

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	return nil, ErrUnimplemented
}
