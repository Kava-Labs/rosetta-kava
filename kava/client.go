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
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	kava "github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

var noBlockResultsForHeight = regexp.MustCompile(`could not find results for height #(\d+)`)

// Client implements services.Client interface for communicating with the kava chain
type Client struct {
	rpc            RPCClient
	encodingConfig params.EncodingConfig
	balanceFactory BalanceServiceFactory
}

// NewClient initialized a new Client with the provided rpc client
func NewClient(rpc RPCClient, balanceServiceFactory BalanceServiceFactory) (*Client, error) {
	encodingConfig := kava.MakeEncodingConfig()

	return &Client{
		rpc:            rpc,
		encodingConfig: encodingConfig,
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
	resultStatus, err := c.rpc.Status(ctx)
	if err != nil {
		return nil, int64(-1), nil, nil, nil, err
	}
	resultNetInfo, err := c.rpc.NetInfo(ctx)
	if err != nil {
		return nil, int64(-1), nil, nil, nil, err
	}
	block, _, err := c.getBlockResult(ctx, nil)
	if err != nil {
		return nil, int64(-1), nil, nil, nil, err
	}

	syncInfo := resultStatus.SyncInfo
	tmPeers := resultNetInfo.Peers

	currentBlock := &types.BlockIdentifier{
		Index: block.Block.Header.Height,
		Hash:  block.BlockID.Hash.String(),
	}
	currentTime := block.Block.Header.Time.UnixNano() / int64(time.Millisecond)

	genesisBlock := &types.BlockIdentifier{
		Index: syncInfo.EarliestBlockHeight,
		Hash:  syncInfo.EarliestBlockHash.String(),
	}

	synced := !syncInfo.CatchingUp
	syncStatus := &types.SyncStatus{
		CurrentIndex: &currentBlock.Index,
		TargetIndex:  &currentBlock.Index,
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

// Account returns the account for the provided address at the latest block height
func (c *Client) Account(ctx context.Context, address sdk.AccAddress) (authtypes.AccountI, error) {
	account, err := c.rpc.Account(ctx, address, 0)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// EstimateGas returns a gas wanted estimate from a tx with a provided adjustment
func (c *Client) EstimateGas(ctx context.Context, tx authsigning.Tx, adjustment float64) (uint64, error) {
	simResp, err := c.rpc.SimulateTx(ctx, tx)
	if err != nil {
		return 0, err
	}

	gas := math.Round(float64(simResp.GasUsed) * (1 + adjustment))

	return uint64(gas), nil
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

	block, _, err := c.getBlockResult(ctx, blockIdentifier)
	if err != nil {
		return nil, err
	}

	balanceService, err := c.balanceFactory(ctx, addr, &block.Block.Header)
	if err != nil {
		return nil, err
	}

	coins, err := balanceService.GetCoinsForSubAccount(ctx, accountIdentifier.SubAccount)
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
	block, deliverResults, err := c.getBlockResult(ctx, blockIdentifier)
	if err != nil {
		return nil, err
	}

	height := block.Block.Header.Height
	if blockIdentifier != nil && blockIdentifier.Index != nil && *blockIdentifier.Index != height {
		return nil, fmt.Errorf("requested index %d does not match returned index %d", *blockIdentifier.Index, height)
	}
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
func (c *Client) getBlockResult(ctx context.Context, blockIdentifier *types.PartialBlockIdentifier) (block *ctypes.ResultBlock, results *ctypes.ResultBlockResults, err error) {
	switch {
	case blockIdentifier == nil:
		// fetch the latest block by passing (*int64)(nil) to tendermint rpc
		results, err = c.rpc.BlockResults(ctx, nil)

		if err != nil {
			// if tendermint returns a no results error, then we must request the previous block
			// since the block has not been fully committed
			if matches := noBlockResultsForHeight.FindStringSubmatch(err.Error()); matches != nil {
				var height int64

				if height, err = strconv.ParseInt(matches[1], 10, 64); err != nil {
					return
				}
				height = height - 1

				results, err = c.rpc.BlockResults(ctx, &height)
				if err != nil {
					return
				}
			} else {
				return
			}
		}

		block, err = c.rpc.Block(ctx, &results.Height)
	case blockIdentifier.Hash != nil:
		hashBytes, decodeErr := hex.DecodeString(*blockIdentifier.Hash)
		if decodeErr != nil {
			return nil, nil, decodeErr
		}
		block, err = c.rpc.BlockByHash(ctx, hashBytes)
	case blockIdentifier.Index != nil:
		block, err = c.rpc.Block(ctx, blockIdentifier.Index)
	}

	if err != nil {
		return
	}

	if results == nil {
		results, err = c.rpc.BlockResults(ctx, &block.Block.Header.Height)
	}

	return
}

func (c *Client) getTransactionsForBlock(
	resultBlock *ctypes.ResultBlock,
	resultBlockResults *ctypes.ResultBlockResults,
) []*types.Transaction {
	// returns transactions -- this will be number of txs + begin/end block (if there)
	eventOpStatus := SuccessStatus
	transactions := []*types.Transaction{}

	beginBlockOps := EventsToOperations(
		stringifyEvents(resultBlockResults.BeginBlockEvents),
		&eventOpStatus,
		0,
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

		tx, err := c.encodingConfig.TxConfig.TxDecoder()(rawTx)
		if err != nil {
			panic(fmt.Sprintf(
				"unable to unmarshal transaction at index %d of block %d: %s",
				i, resultBlock.Block.Header.Height, err,
			))
		}

		sigTx, ok := tx.(authsigning.Tx)
		if !ok {
			panic(fmt.Sprintf(
				"unable to cast transaction at index %d of block %d: %s",
				i, resultBlock.Block.Header.Height, err,
			))
		}

		operations := c.getOperationsForTransaction(sigTx, resultBlockResults.TxsResults[i])
		metadata := c.getMetadataForTransaction(resultBlockResults.TxsResults[i])

		transactions = append(transactions, &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{
				Hash: hash,
			},
			Operations: operations,
			Metadata:   metadata,
		})
	}

	endBlockOps := EventsToOperations(
		stringifyEvents(resultBlockResults.EndBlockEvents),
		&eventOpStatus,
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
	tx authsigning.Tx,
	result *abci.ResponseDeliverTx,
) []*types.Operation {
	opStatus := SuccessStatus
	feeStatus := SuccessStatus

	if result.Code != abci.CodeTypeOK {
		opStatus = FailureStatus
	}

	if result.Codespace == sdkerrors.RootCodespace {
		switch result.Code {
		case sdkerrors.ErrInvalidSequence.ABCICode(), sdkerrors.ErrInsufficientFee.ABCICode(), sdkerrors.ErrWrongSequence.ABCICode():
			feeStatus = FailureStatus
		// For unauthorized and insufficient funds, we must check events in order to know if fee was paid or or not paid
		case sdkerrors.ErrUnauthorized.ABCICode(), sdkerrors.ErrInsufficientFunds.ABCICode():
			feeStatus = FailureStatus

			if containsFee(tx, result) {
				feeStatus = SuccessStatus
			}
		}
	}

	logs, err := sdk.ParseABCILogs(result.Log)
	if err != nil {
		logs = sdk.ABCIMessageLogs{}
	}

	return TxToOperations(tx, logs, &feeStatus, &opStatus)
}

func (c *Client) getMetadataForTransaction(
	result *abci.ResponseDeliverTx,
) map[string]interface{} {
	metadata := make(map[string]interface{})

	if result.Code != abci.CodeTypeOK {
		metadata["log"] = result.Log
	}

	return metadata
}

func stringifyEvents(events []abci.Event) sdk.StringEvents {
	res := make(sdk.StringEvents, 0, len(events))

	for _, e := range events {
		res = append(res, sdk.StringifyEvent(e))
	}

	return res
}

// PostTx broadcasts a transaction and returns an error if it does not get into mempool
func (c *Client) PostTx(ctx context.Context, txBytes []byte) (*types.TransactionIdentifier, error) {
	txRes, err := c.rpc.BroadcastTxSync(ctx, tmtypes.Tx(txBytes))
	if err != nil {
		return nil, err
	}

	if txRes.Code != abci.CodeTypeOK {
		return nil, errors.New(txRes.Log)
	}

	return &types.TransactionIdentifier{Hash: txRes.Hash.String()}, nil
}

// IsRetriableError returns true if the error is retriable or temporary and may succeed on new attempt
func IsRetriableError(err error) bool {
	var rpcError *tmrpctypes.RPCError

	if errors.As(err, &rpcError) {
		if noBlockResultsForHeight.MatchString(rpcError.Data) {
			return true
		}
	}

	return false
}

func containsFee(
	tx authsigning.Tx,
	result *abci.ResponseDeliverTx,
) bool {
	// Check transaction events for fee collector, returning true if found
	for _, event := range stringifyEvents(result.Events) {
		if event.Type == banktypes.EventTypeCoinReceived {
			attributes := make(map[string]string)

			for _, attribute := range event.Attributes {
				attributes[attribute.Key] = attribute.Value
			}

			amount, err := sdk.ParseCoinsNormalized(attributes[sdk.AttributeKeyAmount])
			if err != nil {
				panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
			}

			// Fee was paid
			if attributes[banktypes.AttributeKeyReceiver] == feeCollectorAddress.String() && amount.IsEqual(tx.GetFee()) {
				return true
			}
		}
	}

	return false
}
