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
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	kava "github.com/kava-labs/kava/app"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

func init() {
	//
	// bootstrap kava chain config
	//
	// sets a global cosmos sdk for bech32 prefix
	//
	// required before loading config
	//
	kavaConfig := sdk.GetConfig()
	kava.SetBech32AddressPrefixes(kavaConfig)
	kava.SetBip44CoinType(kavaConfig)
	kavaConfig.Seal()
}

// Client implements services.Client interface for communicating with the kava chain
type Client struct {
	rpc rpcclient.Client
}

// NewClient initialized a new Client with the provided rpc client
func NewClient(rpc rpcclient.Client) (*Client, error) {
	return &Client{
		rpc: rpc,
	}, nil
}

func (c *Client) Status(ctx context.Context) (
	*types.BlockIdentifier,
	int64,
	*types.BlockIdentifier,
	*types.SyncStatus,
	[]*types.Peer,
	error,
) {
	resultStatus, err := c.rpc.Status()
	if err != nil {
		return nil, int64(-1), nil, nil, nil, err
	}
	resultNetInfo, err := c.rpc.NetInfo()
	if err != nil {
		return nil, int64(-1), nil, nil, nil, err
	}

	syncInfo := resultStatus.SyncInfo
	tmPeers := resultNetInfo.Peers

	currentBlock := &types.BlockIdentifier{
		Index: syncInfo.LatestBlockHeight,
		Hash:  syncInfo.LatestBlockHash.String(),
	}
	currentTime := syncInfo.LatestBlockTime.UnixNano() / int64(time.Millisecond)

	genesisBlock := &types.BlockIdentifier{
		Index: syncInfo.EarliestBlockHeight,
		Hash:  syncInfo.EarliestBlockHash.String(),
	}

	synced := !syncInfo.CatchingUp
	syncStatus := &types.SyncStatus{
		CurrentIndex: &syncInfo.LatestBlockHeight,
		TargetIndex:  &syncInfo.LatestBlockHeight,
		Synced:       &synced,
	}

	peers := []*types.Peer{}
	for _, tmPeer := range tmPeers {
		peers = append(peers, &types.Peer{
			PeerID: string(tmPeer.NodeInfo.DefaultNodeID),
			Metadata: map[string]interface{}{
				"Moniker":    tmPeer.NodeInfo.Moniker,
				"Network":    tmPeer.NodeInfo.Network,
				"Version":    tmPeer.NodeInfo.Version,
				"ListenAddr": tmPeer.NodeInfo.ListenAddr,
				"IsOutbound": tmPeer.IsOutbound,
				"RemoteIP":   tmPeer.RemoteIP,
			},
		})
	}

	return currentBlock, currentTime, genesisBlock, syncStatus, peers, nil
}
