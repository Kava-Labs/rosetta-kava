// +build integration
// Copyright 2021 Kava Labs, Inc.
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

package testing

import (
	"context"
	"testing"

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockOffline(t *testing.T) {
	if config.Mode.String() == "online" {
		t.Skip("skipping block offline test")
	}

	ctx := context.Background()

	blockIndex := int64(1)
	request := &types.BlockRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &blockIndex,
		},
	}

	_, rosettaErr, err := client.BlockAPI.Block(ctx, request)
	require.Error(t, err)
	require.NotNil(t, rosettaErr)

	err = asserter.Error(rosettaErr)
	require.NoError(t, err)

	assert.Equal(t, int32(1), rosettaErr.Code)
	assert.Equal(t, "Endpoint unavailable offline", rosettaErr.Message)
}

func TestBlockOnline(t *testing.T) {
	if config.Mode.String() == "offline" {
		t.Skip("skipping block online test")
	}

	ctx := context.Background()

	asserter, err := asserter.NewServer(
		kava.OperationTypes,
		kava.HistoricalBalanceSupported,
		[]*types.NetworkIdentifier{config.NetworkIdentifier},
		kava.CallMethods,
		kava.IncludeMempoolCoins,
	)
	require.NoError(t, err)

	networkStatus, rosettaErr, err := client.NetworkAPI.NetworkStatus(
		ctx,
		&types.NetworkRequest{
			NetworkIdentifier: config.NetworkIdentifier,
		})
	require.NoError(t, err)
	require.Nil(t, rosettaErr)

	resultBlock, err := rpc.Block(&networkStatus.CurrentBlockIdentifier.Index)
	require.NoError(t, err)

	currentBlock := networkStatus.CurrentBlockIdentifier
	request := &types.BlockRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Index: &resultBlock.Block.Header.Height,
		},
	}

	blockResponseByIndex, rosettaErr, err := client.BlockAPI.Block(ctx, request)
	require.NoError(t, err)
	require.Nil(t, rosettaErr)

	err = asserter.Block(blockResponseByIndex.Block)
	assert.NoError(t, err)

	assert.Equal(t, &resultBlock.Block.Header.Height, blockResponseByIndex.Block.BlockIdentifier.Index)
	assert.Equal(t, resultBlock.BlockID.Hash.String(), blockResponseByIndex.Block.BlockIdentifier.Hash)

	request = &types.BlockRequest{
		NetworkIdentifier: config.NetworkIdentifier,
		BlockIdentifier: &types.PartialBlockIdentifier{
			Hash: &currentBlock.Hash,
		},
	}

	blockResponseByHash, rosettaErr, err := client.BlockAPI.Block(ctx, request)
	require.NoError(t, err)
	require.Nil(t, rosettaErr)

	err = asserter.Block(blockResponseByHash.Block)
	assert.NoError(t, err)

	assert.Equal(t, &resultBlock.Block.Header.Height, blockResponseByHash.Block.BlockIdentifier.Index)
	assert.Equal(t, resultBlock.BlockID.Hash.String(), blockResponseByHash.Block.BlockIdentifier.Hash)

	assert.Equal(t, blockResponseByHash, blockResponseByIndex)
}
