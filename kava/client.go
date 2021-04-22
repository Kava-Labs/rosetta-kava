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
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	kava "github.com/kava-labs/kava/app"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func init() {
	// bootstrap cosmos-sdk config for kava chain
	kavaConfig := sdk.GetConfig()
	kava.SetBech32AddressPrefixes(kavaConfig)
	kava.SetBip44CoinType(kavaConfig)
	kavaConfig.Seal()
}

// Client implements services.Client interface for communicating with the kava chain
type Client struct {
	rpc RPCClient
}

// NewClient initialized a new Client with the provided rpc client
func NewClient(rpc RPCClient) (*Client, error) {
	return &Client{
		rpc: rpc,
	}, nil
}

// Status fetches latest status from a kava node and returns the results
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

	// TODO: update when indexer is implemented
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
	// TODO: update when indexer is implemented
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

// Balance fetches and returns the account balance for an account
func (c *Client) Balance(
	ctx context.Context,
	accountIdentifier *types.AccountIdentifier,
	blockIdentifier *types.PartialBlockIdentifier,
	currencies []*types.Currency,
) (*types.AccountBalanceResponse, error) {
	addr, err := sdk.AccAddressFromBech32(accountIdentifier.Address)
	if err != nil {
		return nil, err
	}

	var block *ctypes.ResultBlock
	switch {
	case blockIdentifier == nil:
		block, err = c.rpc.Block(nil)
	case blockIdentifier.Index != nil:
		block, err = c.rpc.Block(blockIdentifier.Index)
	case blockIdentifier.Hash != nil:
		hashBytes, decodeErr := hex.DecodeString(*blockIdentifier.Hash)
		if decodeErr != nil {
			return nil, decodeErr
		}
		block, err = c.rpc.BlockByHash(hashBytes)
	}
	if err != nil {
		return nil, err
	}

	acc, err := c.rpc.Account(addr, block.Block.Header.Height)
	if err != nil {
		return nil, err
	}

	spendableCoins := acc.SpendableCoins(block.Block.Header.Time)
	var currencyLookup map[string]*types.Currency

	if currencies == nil {
		currencyLookup = Currencies
	} else {
		currencyLookup = make(map[string]*types.Currency)

		for _, currency := range currencies {
			denom, ok := Denoms[currency.Symbol]

			if ok {
				currencyLookup[denom] = Currencies[denom]
			}
		}
	}

	balances := []*types.Amount{}
	for _, coin := range spendableCoins {
		currency, ok := currencyLookup[coin.Denom]
		if !ok {
			continue
		}

		balances = append(balances, &types.Amount{
			Value:    coin.Amount.String(),
			Currency: currency,
		})
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: block.Block.Header.Height,
			Hash:  block.BlockID.Hash.String(),
		},
		Balances: balances,
	}, nil
}

func (c *Client) Block(
	ctx context.Context,
	blockIdentifier *types.PartialBlockIdentifier,
) (*types.BlockResponse, error) {
	var (
		block *ctypes.ResultBlock
		err   error
	)

	switch {
	case blockIdentifier == nil:
		block, err = c.rpc.Block(nil)
	case blockIdentifier.Index != nil:
		block, err = c.rpc.Block(blockIdentifier.Index)
	case blockIdentifier.Hash != nil:
		hashBytes, decodeErr := hex.DecodeString(*blockIdentifier.Hash)
		if decodeErr != nil {
			return nil, decodeErr
		}
		block, err = c.rpc.BlockByHash(hashBytes)
	}
	if err != nil {
		return nil, err
	}

	height := block.Block.Header.Height
	identifier := &types.BlockIdentifier{
		Index: height,
		Hash:  block.BlockID.Hash.String(),
	}

	var parentIdentifier *types.BlockIdentifier
	if height == 1 {
		parentIdentifier = identifier
	} else {
		parentIdentifier = &types.BlockIdentifier{
			Index: height - 1,
			Hash:  block.Block.Header.LastBlockID.Hash.String(),
		}
	}

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       identifier,
			ParentBlockIdentifier: parentIdentifier,
			Timestamp:             block.Block.Header.Time.UnixNano() / int64(1e6),
		},
	}, nil
}
