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
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

var (
	networkIdentifier = &types.NetworkIdentifier{
		Blockchain: kava.Blockchain,
		Network:    "kava-testnet-1",
	}

	expectedNetworkOptions = &types.NetworkOptionsResponse{
		Version: &types.Version{
			RosettaVersion:    types.RosettaAPIVersion,
			NodeVersion:       kava.NodeVersion,
			MiddlewareVersion: &configuration.MiddlewareVersion,
		},
		Allow: &types.Allow{
			OperationStatuses:       kava.OperationStatuses,
			OperationTypes:          kava.OperationTypes,
			Errors:                  Errors,
			HistoricalBalanceLookup: kava.HistoricalBalanceSupported,
			CallMethods:             kava.CallMethods,
			BalanceExemptions:       kava.BalanceExemptions,
			MempoolCoins:            kava.IncludeMempoolCoins,
		},
	}
)

func TestNetworkEndpoints_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode:              configuration.Offline,
		NetworkIdentifier: networkIdentifier,
	}

	mockClient := &mocks.Client{}
	servicer := NewNetworkAPIService(cfg, mockClient)
	ctx := context.Background()

	networkList, err := servicer.NetworkList(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			networkIdentifier,
		},
	}, networkList)

	networkOptions, err := servicer.NetworkOptions(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, expectedNetworkOptions, networkOptions)

	networkStatus, err := servicer.NetworkStatus(ctx, nil)
	assert.Nil(t, networkStatus)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	mockClient.AssertExpectations(t)
}

func TestNetworkEndpoints_Online(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode:              configuration.Online,
		NetworkIdentifier: networkIdentifier,
	}
	mockClient := &mocks.Client{}
	servicer := NewNetworkAPIService(cfg, mockClient)
	ctx := context.Background()

	networkList, err := servicer.NetworkList(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, &types.NetworkListResponse{
		NetworkIdentifiers: []*types.NetworkIdentifier{
			networkIdentifier,
		},
	}, networkList)

	networkOptions, err := servicer.NetworkOptions(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, expectedNetworkOptions, networkOptions)

	currentBlock := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	currentTime := int64(1000000000000)
	genesisBlock := &types.BlockIdentifier{
		Index: 1,
		Hash:  "ADB03E823AFC5F12DC02D984A7E1E0EC47E84FC323005B82FB0B3A9DC8F045B7",
	}
	syncStatus := &types.SyncStatus{}
	peers := []*types.Peer{
		{
			PeerID: "e5d74b3f06226fb0798509e36021e81b7bce934d",
		},
	}

	mockClient.On(
		"Status",
		ctx,
	).Return(
		currentBlock,
		currentTime,
		genesisBlock,
		syncStatus,
		peers,
		nil,
	).Once()

	networkRequest := &types.NetworkRequest{
		NetworkIdentifier: networkIdentifier,
	}
	networkStatus, err := servicer.NetworkStatus(ctx, networkRequest)
	assert.Nil(t, err)

	assert.Equal(t, &types.NetworkStatusResponse{
		CurrentBlockIdentifier: currentBlock,
		CurrentBlockTimestamp:  currentTime,
		GenesisBlockIdentifier: genesisBlock,
		SyncStatus:             syncStatus,
		Peers:                  peers,
	}, networkStatus)

	kavaErr := errors.New("some client error")
	mockClient.On(
		"Status",
		ctx,
	).Return(
		nil,
		int64(-1),
		nil,
		nil,
		nil,
		kavaErr,
	).Once()

	networkStatus, err = servicer.NetworkStatus(ctx, networkRequest)
	assert.Nil(t, networkStatus)
	assert.NotNil(t, err)
	assert.Equal(t, ErrKava.Code, err.Code)
	assert.Equal(t, ErrKava.Message, err.Message)
	assert.Equal(t, kavaErr.Error(), err.Details["context"])

	mockClient.AssertExpectations(t)
}
