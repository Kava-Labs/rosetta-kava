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
	"fmt"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	kava "github.com/kava-labs/kava/app"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Client implements services.Client interface for communicating with the kava chain
type Client struct {
	rpc            RPCClient
	cdc            *codec.Codec
	balanceFactory BalanceServiceFactory
}

// NewClient initialized a new Client with the provided rpc client
func NewClient(rpc RPCClient, balanceServiceFactory BalanceServiceFactory) (*Client, error) {
	cdc := kava.MakeCodec()

	return &Client{
		rpc:            rpc,
		cdc:            cdc,
		balanceFactory: balanceServiceFactory,
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

	block, err := c.getBlockResult(blockIdentifier)
	if err != nil {
		return nil, err
	}

	balanceService, err := c.balanceFactory(addr, &block.Block.Header)
	if err != nil {
		return nil, err
	}

	coins, err := balanceService.GetCoinsForSubAccount(accountIdentifier.SubAccount)
	if err != nil {
		return nil, err
	}

	balances := c.getBalancesAndFilterByCurrency(coins, currencies)

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{
			Index: block.Block.Header.Height,
			Hash:  block.BlockID.Hash.String(),
		},
		Balances: balances,
	}, nil
}

func (c *Client) getBalancesAndFilterByCurrency(
	coins sdk.Coins,
	currencies []*types.Currency,
) []*types.Amount {
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

	for denom, currency := range currencyLookup {
		value := coins.AmountOf(denom)

		balances = append(balances, &types.Amount{
			Value:    value.String(),
			Currency: currency,
		})
	}

	return balances
}

// Block returns rosetta block for an index or hash
func (c *Client) Block(
	ctx context.Context,
	blockIdentifier *types.PartialBlockIdentifier,
) (*types.BlockResponse, error) {
	block, err := c.getBlockResult(blockIdentifier)
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

	deliverResults, err := c.rpc.BlockResults(&height)
	if err != nil {
		return nil, err
	}

	transactions := c.getTransactionsForBlock(block, deliverResults)

	return &types.BlockResponse{
		Block: &types.Block{
			BlockIdentifier:       identifier,
			ParentBlockIdentifier: parentIdentifier,
			Timestamp:             block.Block.Header.Time.UnixNano() / int64(1e6),
			Transactions:          transactions,
		},
	}, nil
}

// getBlockResult returns the specified block by Index or Hash. If the
// block identifier is not provided, then the latest block is returned
func (c *Client) getBlockResult(blockIdentifier *types.PartialBlockIdentifier) (block *ctypes.ResultBlock, err error) {
	switch {
	case blockIdentifier == nil:
		// fetch the latest block by passing (*int64)(nil) to tendermint rpc
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

	return
}

func (c *Client) getTransactionsForBlock(
	resultBlock *ctypes.ResultBlock,
	resultBlockResults *ctypes.ResultBlockResults,
) []*types.Transaction {
	// returns transactions -- this will be number of txs + begin/end block (if there)
	transactions := []*types.Transaction{}

	// TODO: vesting account operations (must be before being blocker operations)
	//	add them before begin blocker ops, and pass updated index to op method

	beginBlockOps := EventsToOperations(
		stringifyEvents(resultBlockResults.BeginBlockEvents),
		0, // TODO: update index to be after vesting ops
	)

	if len(beginBlockOps) > 0 {
		transactions = append(transactions, &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: BeginBlockTxHash(resultBlock.BlockID.Hash),
			},
			Operations: beginBlockOps,
		})
	}

	// transaction loop
	for i, rawTx := range resultBlock.Block.Data.Txs {
		hash := strings.ToUpper(hex.EncodeToString(rawTx.Hash()))

		var tx authtypes.StdTx
		err := c.cdc.UnmarshalBinaryLengthPrefixed(rawTx, &tx)
		if err != nil {
			panic(fmt.Sprintf(
				"unable to unmarshal transaction at index %d of block %d: %s",
				i, resultBlock.Block.Header.Height, err,
			))
		}

		operations := c.getOperationsForTransaction(&tx, resultBlockResults.TxsResults[i])

		transactions = append(transactions, &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: hash,
			},
			Operations: operations,
		})
	}

	endBlockOps := EventsToOperations(
		stringifyEvents(resultBlockResults.EndBlockEvents),
		0,
	)

	if len(endBlockOps) > 0 {
		transactions = append(transactions, &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: EndBlockTxHash(resultBlock.BlockID.Hash),
			},
			Operations: endBlockOps,
		})
	}

	return transactions
}

func (c *Client) getOperationsForTransaction(
	tx *authtypes.StdTx,
	result *abci.ResponseDeliverTx,
) []*types.Operation {
	var status string

	if result.Code == abci.CodeTypeOK {
		status = SuccessStatus
	} else {
		status = FailureStatus
	}

	logs, err := sdk.ParseABCILogs(result.Log)
	if err != nil {
		logs = sdk.ABCIMessageLogs{}
	}

	return TxToOperations(tx, logs, &status)
}

func stringifyEvents(events []abci.Event) sdk.StringEvents {
	res := make(sdk.StringEvents, 0, len(events))

	for _, e := range events {
		res = append(res, sdk.StringifyEvent(e))
	}

	return res
}
