//go:build integration
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

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkList(t *testing.T) {
	ctx := context.Background()

	networkList, rosettaErr, err := client.NetworkAPI.NetworkList(ctx, &types.MetadataRequest{})
	require.Nil(t, rosettaErr)
	require.NoError(t, err)

	err = asserter.NetworkListResponse(networkList)
	require.NoError(t, err)

	require.Equal(t, 1, len(networkList.NetworkIdentifiers))
	network := networkList.NetworkIdentifiers[0]

	assert.Equal(t, "Kava", network.Blockchain)
	assert.Equal(t, config.NetworkIdentifier.Network, network.Network)
}

func TestNetworkOptions(t *testing.T) {
	ctx := context.Background()

	networkOptions, rosettaErr, err := client.NetworkAPI.NetworkOptions(
		ctx,
		&types.NetworkRequest{
			NetworkIdentifier: config.NetworkIdentifier,
		})
	require.Nil(t, rosettaErr)
	require.NoError(t, err)

	err = asserter.NetworkOptionsResponse(networkOptions)
	require.NoError(t, err)

	assert.NotEmpty(t, networkOptions.Version.RosettaVersion)
	assert.NotEmpty(t, networkOptions.Version.NodeVersion)
	assert.NotEmpty(t, networkOptions.Version.MiddlewareVersion)
}

func TestNetworkStatus(t *testing.T) {
	ctx := context.Background()

	networkStatus, rosettaErr, err := client.NetworkAPI.NetworkStatus(
		ctx,
		&types.NetworkRequest{
			NetworkIdentifier: config.NetworkIdentifier,
		})

	if config.Mode.String() == "online" {
		require.Nil(t, rosettaErr)
		require.NoError(t, err)

		err = asserter.NetworkStatusResponse(networkStatus)
		require.NoError(t, err)

		resultStatus, err := rpc.Status(ctx)
		require.NoError(t, err)

		assert.Equal(t, &types.BlockIdentifier{
			Index: resultStatus.SyncInfo.EarliestBlockHeight,
			Hash:  resultStatus.SyncInfo.EarliestBlockHash.String(),
		}, networkStatus.GenesisBlockIdentifier)
	} else {
		require.Error(t, err)
		require.NotNil(t, rosettaErr)

		err = asserter.Error(rosettaErr)
		require.NoError(t, err)

		assert.Equal(t, int32(1), rosettaErr.Code)
		assert.Equal(t, "Endpoint unavailable offline", rosettaErr.Message)
	}
}
