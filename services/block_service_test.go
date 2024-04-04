// Copyright 2021 Kava Labs, Inc.
// Copyright 2020 Coinbase, Inc.
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
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	tmstate "github.com/cometbft/cometbft/state"
)

func TestBlockService_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	servicer := NewBlockAPIService(cfg, mockClient)
	ctx := context.Background()

	block, err := servicer.Block(ctx, &types.BlockRequest{})
	assert.Nil(t, block)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	blockTransaction, err := servicer.BlockTransaction(ctx, &types.BlockTransactionRequest{})
	assert.Nil(t, blockTransaction)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)
	assert.Equal(t, ErrUnavailableOffline.Message, err.Message)

	mockClient.AssertExpectations(t)
}

func TestBlockService_Online(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Online,
	}
	mockClient := &mocks.Client{}
	servicer := NewBlockAPIService(cfg, mockClient)
	ctx := context.Background()

	blockResponse := &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier: &types.BlockIdentifier{
				Index: 100,
				Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
			},
		},
	}

	blockIdentifier := &types.PartialBlockIdentifier{
		Index: &blockResponse.Block.BlockIdentifier.Index,
	}

	mockClient.On(
		"Block",
		ctx,
		blockIdentifier,
	).Return(
		blockResponse,
		nil,
	).Once()

	block, err := servicer.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	})
	require.Nil(t, err)
	assert.Equal(t, blockResponse, block)

	blockIdentifier = &types.PartialBlockIdentifier{
		Hash: &blockResponse.Block.BlockIdentifier.Hash,
	}

	mockClient.On(
		"Block",
		ctx,
		blockIdentifier,
	).Return(
		blockResponse,
		nil,
	).Once()

	block, err = servicer.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	})
	require.Nil(t, err)
	assert.Equal(t, blockResponse, block)

	kavaErr := errors.New("some client error")
	mockClient.On(
		"Block",
		ctx,
		blockIdentifier,
	).Return(
		nil,
		kavaErr,
	).Once()

	block, err = servicer.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	})
	assert.Nil(t, block)
	assert.Equal(t, ErrKava.Code, err.Code)
	assert.Equal(t, ErrKava.Message, err.Message)
	// errors are not retriable
	assert.Equal(t, ErrKava.Retriable, false)
	assert.Equal(t, kavaErr.Error(), err.Details["context"])

	blockTransaction, err := servicer.BlockTransaction(ctx, &types.BlockTransactionRequest{})
	assert.Nil(t, blockTransaction)
	assert.Equal(t, ErrUnimplemented.Code, err.Code)
	assert.Equal(t, ErrUnimplemented.Message, err.Message)

	abciErr := tmstate.ErrNoABCIResponsesForHeight{Height: 10001}
	kavaErr = tmrpctypes.RPCInternalError(tmrpctypes.JSONRPCIntID(1), abciErr).Error
	mockClient.On(
		"Block",
		ctx,
		blockIdentifier,
	).Return(
		nil,
		kavaErr,
	).Once()

	block, err = servicer.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	})
	assert.Nil(t, block)
	// if block results fail to fetch this is retriable
	assert.Equal(t, err.Retriable, true, "expected could not find results error to be retriable")

	kavaErr = errors.New("some non-retriable client error")
	mockClient.On(
		"Block",
		ctx,
		blockIdentifier,
	).Return(
		nil,
		kavaErr,
	).Once()

	_, err = servicer.Block(ctx, &types.BlockRequest{
		NetworkIdentifier: networkIdentifier,
		BlockIdentifier:   blockIdentifier,
	})
	// errors are not retriable
	assert.Equal(t, err.Retriable, false, "expected error to not be retriable")

	mockClient.AssertExpectations(t)
}
