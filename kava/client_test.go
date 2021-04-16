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

package kava

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	mocks "github.com/kava-labs/rosetta-kava/mocks/tendermint"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func TestCosmosSDKConfig(t *testing.T) {
	config := sdk.GetConfig()

	coinType := config.GetCoinType()
	assert.Equal(t, uint32(459), coinType)

	prefix := config.GetBech32AccountAddrPrefix()
	assert.Equal(t, "kava", prefix)

	prefix = config.GetBech32ValidatorAddrPrefix()
	assert.Equal(t, "kavavaloper", prefix)

	prefix = config.GetBech32ConsensusAddrPrefix()
	assert.Equal(t, "kavavalcons", prefix)

	prefix = config.GetBech32AccountPubPrefix()
	assert.Equal(t, "kavapub", prefix)

	prefix = config.GetBech32ConsensusPubPrefix()
	assert.Equal(t, "kavavalconspub", prefix)

	assert.PanicsWithValue(t, "Config is sealed", func() { config.SetCoinType(459) })
}

func TestStatus(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	rpcErr := errors.New("unable to contact node")
	mockRPCClient.On(
		"Status",
	).Return(
		nil,
		rpcErr,
	).Once()

	currentBlock,
		currentTime,
		genesisBlock,
		syncStatus,
		peers,
		err := client.Status(ctx)
	assert.Nil(t, currentBlock)
	assert.Equal(t, int64(-1), currentTime)
	assert.Nil(t, genesisBlock)
	assert.Nil(t, syncStatus)
	assert.Nil(t, peers)
	assert.Equal(t, rpcErr, err)

	latestBlockHashStr := "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75"
	latestBlockHash, err := hex.DecodeString(latestBlockHashStr)
	require.NoError(t, err)
	latestBlockTime, err := time.Parse(time.RFC3339Nano, "2021-04-08T15:13:25.837676922Z")
	require.NoError(t, err)

	earliestBlockHashStr := "ADB03E823AFC5F12DC02D984A7E1E0EC47E84FC323005B82FB0B3A9DC8F045B7"
	earliestBlockHash, err := hex.DecodeString(earliestBlockHashStr)
	require.NoError(t, err)
	earliestBlockTime, err := time.Parse(time.RFC3339Nano, "2021-04-08T15:00:00Z")
	require.NoError(t, err)

	syncInfo := ctypes.SyncInfo{
		LatestBlockHash:     bytes.HexBytes(latestBlockHash),
		LatestBlockHeight:   int64(100),
		LatestBlockTime:     latestBlockTime,
		EarliestBlockHash:   bytes.HexBytes(earliestBlockHash),
		EarliestBlockHeight: int64(0),
		EarliestBlockTime:   earliestBlockTime,
		CatchingUp:          false,
	}

	mockRPCClient.On(
		"Status",
	).Return(
		&ctypes.ResultStatus{
			NodeInfo:      p2p.DefaultNodeInfo{},
			SyncInfo:      syncInfo,
			ValidatorInfo: ctypes.ValidatorInfo{},
		},
		nil,
	)

	mockRPCClient.On(
		"NetInfo",
	).Return(
		nil,
		rpcErr,
	).Once()

	currentBlock,
		currentTime,
		genesisBlock,
		syncStatus,
		peers,
		err = client.Status(ctx)
	assert.Nil(t, currentBlock)
	assert.Equal(t, int64(-1), currentTime)
	assert.Nil(t, genesisBlock)
	assert.Nil(t, syncStatus)
	assert.Nil(t, peers)
	assert.Equal(t, rpcErr, err)

	tmPeer := ctypes.Peer{
		NodeInfo: p2p.DefaultNodeInfo{
			DefaultNodeID: "e5d74b3f06226fb0798509e36021e81b7bce934d",
			Moniker:       "kava-node",
			Network:       "kava-7",
			Version:       "0.33.9",
			ListenAddr:    "tcp://192.168.1.1:26656",
		},
		IsOutbound: false,
		RemoteIP:   "192.168.1.1",
	}

	tmPeers := []ctypes.Peer{tmPeer}
	mockRPCClient.On(
		"NetInfo",
	).Return(
		&ctypes.ResultNetInfo{
			Peers: tmPeers,
		},
		nil,
	).Once()

	currentBlock,
		currentTime,
		genesisBlock,
		syncStatus,
		peers,
		err = client.Status(ctx)
	require.NoError(t, err)

	assert.Equal(t, &types.BlockIdentifier{
		Index: int64(100),
		Hash:  latestBlockHashStr,
	}, currentBlock)
	assert.Equal(t, latestBlockTime.UnixNano()/int64(time.Millisecond), currentTime)
	assert.Equal(t, &types.BlockIdentifier{
		Index: int64(0),
		Hash:  earliestBlockHashStr,
	}, genesisBlock)

	currentIndex := int64(100)
	targetIndex := int64(100)
	synced := true
	assert.Equal(t, &types.SyncStatus{
		CurrentIndex: &currentIndex,
		TargetIndex:  &targetIndex,
		Synced:       &synced,
	}, syncStatus)
	assert.Equal(t, []*types.Peer{
		{
			PeerID: string(tmPeer.NodeInfo.DefaultNodeID),
			Metadata: map[string]interface{}{
				"Moniker":    tmPeer.NodeInfo.Moniker,
				"Network":    tmPeer.NodeInfo.Network,
				"Version":    tmPeer.NodeInfo.Version,
				"ListenAddr": tmPeer.NodeInfo.ListenAddr,
				"IsOutbound": tmPeer.IsOutbound,
				"RemoteIP":   tmPeer.RemoteIP,
			},
		},
	}, peers)

	mockRPCClient.AssertExpectations(t)
}

func TestBalance(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	accountResponse, err := client.Balance(
		ctx,
		&types.AccountIdentifier{},
		&types.PartialBlockIdentifier{},
		[]*types.Currency{},
	)
	assert.Nil(t, accountResponse)
	assert.Error(t, err)
}
