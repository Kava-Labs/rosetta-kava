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
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	"github.com/stretchr/testify/assert"
)

func setupConstructionAPIServicer() (*ConstructionAPIService, *mocks.Client) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	encodingConfig := app.MakeEncodingConfig()
	return NewConstructionAPIService(cfg, mockClient, encodingConfig), mockClient
}

func setupConstructionAPIServicerWithEncodingConfig(encodingConfig params.EncodingConfig) (*ConstructionAPIService, *mocks.Client) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	return NewConstructionAPIService(cfg, mockClient, encodingConfig), mockClient
}

func TestConstructionService_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}

	mockClient := &mocks.Client{}
	encodingConfig := app.MakeEncodingConfig()
	servicer := NewConstructionAPIService(cfg, mockClient, encodingConfig)
	ctx := context.Background()

	// Test Metadata
	metadataResponse, err := servicer.ConstructionMetadata(ctx, &types.ConstructionMetadataRequest{})
	assert.Nil(t, metadataResponse)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	// Test Submit
	submitResponse, err := servicer.ConstructionSubmit(ctx, &types.ConstructionSubmitRequest{})
	assert.Nil(t, submitResponse)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	mockClient.AssertExpectations(t)
}
