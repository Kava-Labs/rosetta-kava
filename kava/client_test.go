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

package kava_test

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/kava-labs/rosetta-kava/kava"
	mocks "github.com/kava-labs/rosetta-kava/kava/mocks"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	app "github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmstate "github.com/tendermint/tendermint/state"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	latestSyncBlockHashStr = "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75"
	latestSyncBlockTime    = "2021-04-08T15:13:25.837676922Z"
	latestBlockHashStr     = "3A3FB732715054D56313354A8CEB135848363AB97E9323559E419F3D09BA4B31"
	latestBlockTime        = "2021-04-08T15:06:25.345443445Z"
	earliestBlockHashStr   = "ADB03E823AFC5F12DC02D984A7E1E0EC47E84FC323005B82FB0B3A9DC8F045B7"
	earliestBlockTime      = "2021-04-08T15:00:00Z"
)

func newResultStatus(t *testing.T) *ctypes.ResultStatus {
	latestBlockHash, err := hex.DecodeString(latestSyncBlockHashStr)
	require.NoError(t, err)
	latestBlockTime, err := time.Parse(time.RFC3339Nano, latestSyncBlockTime)
	require.NoError(t, err)

	earliestBlockHash, err := hex.DecodeString(earliestBlockHashStr)
	require.NoError(t, err)
	earliestBlockTime, err := time.Parse(time.RFC3339Nano, earliestBlockTime)
	require.NoError(t, err)

	syncInfo := ctypes.SyncInfo{
		LatestBlockHash:     bytes.HexBytes(latestBlockHash),
		LatestBlockHeight:   int64(101),
		LatestBlockTime:     latestBlockTime,
		EarliestBlockHash:   bytes.HexBytes(earliestBlockHash),
		EarliestBlockHeight: int64(0),
		EarliestBlockTime:   earliestBlockTime,
		CatchingUp:          false,
	}

	return &ctypes.ResultStatus{
		NodeInfo:      p2p.DefaultNodeInfo{},
		SyncInfo:      syncInfo,
		ValidatorInfo: ctypes.ValidatorInfo{},
	}
}

func newResultNetInfo() *ctypes.ResultNetInfo {
	tmPeer := ctypes.Peer{
		NodeInfo: p2p.DefaultNodeInfo{
			DefaultNodeID: "e5d74b3f06226fb0798509e36021e81b7bce934d",
			Moniker:       "kava-node",
			Network:       "kava-9",
			Version:       "0.33.9",
			ListenAddr:    "tcp://192.168.1.1:26656",
		},
		IsOutbound: false,
		RemoteIP:   "192.168.1.1",
	}

	tmPeers := []ctypes.Peer{tmPeer}

	return &ctypes.ResultNetInfo{
		Peers: tmPeers,
	}
}

func newBlockWithResult(t *testing.T) (*types.BlockIdentifier, *ctypes.ResultBlock) {
	block := &types.BlockIdentifier{
		Index: 100,
		Hash:  latestBlockHashStr,
	}
	blockTime, err := time.Parse(time.RFC3339Nano, latestBlockTime)
	require.NoError(t, err)
	hashBytes, err := hex.DecodeString(block.Hash)
	require.NoError(t, err)

	resultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: hashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: block.Index,
				Time:   blockTime,
			},
		},
	}

	return block, resultBlock
}

func setupClient(t *testing.T) (*mocks.RPCClient, *mocks.BalanceServiceFactory, *kava.Client) {
	mockRPCClient := &mocks.RPCClient{}
	mockBalanceFactory := &mocks.BalanceServiceFactory{}
	client, err := kava.NewClient(mockRPCClient, mockBalanceFactory.Execute)
	require.NoError(t, err)

	return mockRPCClient, mockBalanceFactory, client
}

func TestStatus(t *testing.T) {
	t.Run("rpc error when getting node status", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		rpcErr := errors.New("unable to contact node")
		mockRPCClient.On("Status", ctx).Return(nil, rpcErr)

		currentBlock, currentTime, genesisBlock, syncStatus, peers, err := client.Status(ctx)

		assert.Nil(t, currentBlock)
		assert.Equal(t, int64(-1), currentTime)
		assert.Nil(t, genesisBlock)
		assert.Nil(t, syncStatus)
		assert.Nil(t, peers)
		assert.Equal(t, rpcErr, err)
	})

	t.Run("rpc error when getting net info for peers", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		mockRPCClient.On("Status", ctx).Return(newResultStatus(t), nil)
		rpcErr := errors.New("unable to contact node")
		mockRPCClient.On("NetInfo", ctx).Return(nil, rpcErr).Once()

		currentBlock, currentTime, genesisBlock, syncStatus, peers, err := client.Status(ctx)

		assert.Nil(t, currentBlock)
		assert.Equal(t, int64(-1), currentTime)
		assert.Nil(t, genesisBlock)
		assert.Nil(t, syncStatus)
		assert.Nil(t, peers)
		assert.Equal(t, rpcErr, err)
	})

	t.Run("rpc error when getting latest block results", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		mockRPCClient.On("Status", ctx).Return(newResultStatus(t), nil)
		mockRPCClient.On("NetInfo", ctx).Return(newResultNetInfo(), nil)

		rpcErr := errors.New("unable to contact node")
		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(nil, rpcErr).Once()

		currentBlock, currentTime, genesisBlock, syncStatus, peers, err := client.Status(ctx)

		assert.Nil(t, currentBlock)
		assert.Equal(t, int64(-1), currentTime)
		assert.Nil(t, genesisBlock)
		assert.Nil(t, syncStatus)
		assert.Nil(t, peers)
		assert.Equal(t, rpcErr, err)
	})

	t.Run("successful response", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		mockRPCClient.On("Status", ctx).Return(newResultStatus(t), nil)
		mockRPCClient.On("NetInfo", ctx).Return(newResultNetInfo(), nil)

		_, mockResultBlock := newBlockWithResult(t)
		mockResultBlockResults := &ctypes.ResultBlockResults{Height: mockResultBlock.Block.Height}

		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(mockResultBlockResults, nil)
		mockRPCClient.On("Block", ctx, &mockResultBlock.Block.Height).Return(mockResultBlock, nil)

		currentBlock, currentTime, genesisBlock, syncStatus, peers, err := client.Status(ctx)
		require.NoError(t, err)

		assert.Equal(t, &types.BlockIdentifier{Index: int64(100), Hash: latestBlockHashStr}, currentBlock)
		latestBlockTime, err := time.Parse(time.RFC3339Nano, latestBlockTime)
		require.NoError(t, err)
		assert.Equal(t, latestBlockTime.UnixNano()/int64(time.Millisecond), currentTime)
		assert.Equal(t, &types.BlockIdentifier{Index: int64(0), Hash: earliestBlockHashStr}, genesisBlock)

		currentIndex := int64(100)
		targetIndex := int64(100)
		synced := true
		assert.Equal(t, &types.SyncStatus{CurrentIndex: &currentIndex, TargetIndex: &targetIndex, Synced: &synced}, syncStatus)

		tmPeer := newResultNetInfo().Peers[0]
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
	})
}

func TestBalance_InvalidAddress(t *testing.T) {
	_, _, client := setupClient(t)

	invalidAcc := &types.AccountIdentifier{Address: "invalid"}

	ctx := context.Background()
	accountResponse, err := client.Balance(ctx, invalidAcc, nil, nil)

	assert.Nil(t, accountResponse)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestBalance_NoFilters(t *testing.T) {
	t.Run("error fetching latest block", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		blockErr := errors.New("error getting block")
		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(nil, blockErr).Once()

		accountResponse, err := client.Balance(ctx, acc, nil, nil)

		assert.Nil(t, accountResponse)
		assert.EqualError(t, err, blockErr.Error())
	})

	t.Run("error getting balance service for account", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, mockBalanceFactory, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		_, resultBlock := newBlockWithResult(t)
		resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(resultBlockResults, nil)
		mockRPCClient.On("Block", ctx, &resultBlock.Block.Height).Return(resultBlock, nil)

		balErr := errors.New("could not find account")
		mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
			nil,
			balErr,
		)

		accountResponse, err := client.Balance(ctx, acc, nil, nil)

		assert.Nil(t, accountResponse)
		assert.EqualError(t, err, balErr.Error())
	})

	t.Run("error getting coins for account", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, mockBalanceFactory, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		_, resultBlock := newBlockWithResult(t)
		resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(resultBlockResults, nil)
		mockRPCClient.On("Block", ctx, &resultBlock.Block.Height).Return(resultBlock, nil)

		mockBalanceService := &mocks.AccountBalanceService{}
		mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
			mockBalanceService,
			nil,
		)
		balErr := errors.New("could not get coins for account")
		mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(nil, uint64(0), balErr)

		accountResponse, err := client.Balance(ctx, acc, nil, nil)
		assert.Nil(t, accountResponse)
		assert.EqualError(t, err, balErr.Error())

	})

	t.Run("successful balance response", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, mockBalanceFactory, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		block, resultBlock := newBlockWithResult(t)
		resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(resultBlockResults, nil)
		mockRPCClient.On("Block", ctx, &resultBlock.Block.Height).Return(resultBlock, nil)

		mockBalanceService := &mocks.AccountBalanceService{}
		mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
			mockBalanceService,
			nil,
		)
		coins := generateDefaultCoins()
		mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(coins, uint64(0), nil)

		accountResponse, err := client.Balance(ctx, acc, nil, nil)
		require.NoError(t, err)

		assert.Equal(t, block, accountResponse.BlockIdentifier)
		assert.Greater(t, len(accountResponse.Balances), 0)

		for _, amount := range accountResponse.Balances {
			denom := kava.Denoms[amount.Currency.Symbol]
			require.NotEmpty(t, denom)

			assert.Equal(t, kava.Currencies[denom], amount.Currency)
			assert.Equal(t, coins.AmountOf(denom).String(), amount.Value)
		}
	})
}

func TestBalance_BlockFilter(t *testing.T) {
	t.Run("filter by block index", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, mockBalanceFactory, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		block, resultBlock := newBlockWithResult(t)
		resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
		mockRPCClient.On("Block", ctx, &block.Index).Return(resultBlock, nil).Once()
		mockRPCClient.On("BlockResults", ctx, &block.Index).Return(resultBlockResults, nil).Once()

		mockBalanceService := &mocks.AccountBalanceService{}
		mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
			mockBalanceService,
			nil,
		)
		coins := generateDefaultCoins()
		mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(coins, uint64(0), nil)

		blockFilter := &types.PartialBlockIdentifier{Index: &block.Index}
		accountResponse, err := client.Balance(ctx, acc, blockFilter, nil)
		require.NoError(t, err)
		assert.Equal(t, block, accountResponse.BlockIdentifier)

		blockErr := errors.New("some block index error")
		mockRPCClient.On("Block", ctx, &block.Index).Return(nil, blockErr).Once()

		accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
		assert.Nil(t, accountResponse)
		assert.EqualError(t, err, blockErr.Error())
	})

	t.Run("filter by block hash", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, mockBalanceFactory, client := setupClient(t)

		testAccount, _ := newTestAccount(t)
		acc := &types.AccountIdentifier{Address: testAccount.Address}
		block, resultBlock := newBlockWithResult(t)
		resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
		mockRPCClient.On("BlockByHash", ctx, []byte(resultBlock.BlockID.Hash)).Return(resultBlock, nil).Once()
		mockRPCClient.On("BlockResults", ctx, &block.Index).Return(resultBlockResults, nil).Once()

		mockBalanceService := &mocks.AccountBalanceService{}
		mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
			mockBalanceService,
			nil,
		)
		coins := generateDefaultCoins()
		mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(coins, uint64(0), nil)

		blockFilter := &types.PartialBlockIdentifier{Hash: &block.Hash}
		accountResponse, err := client.Balance(ctx, acc, blockFilter, nil)
		require.NoError(t, err)
		assert.Equal(t, block, accountResponse.BlockIdentifier)

		blockErr := errors.New("some block index error")
		mockRPCClient.On("BlockByHash", ctx, []byte(resultBlock.BlockID.Hash)).Return(nil, blockErr).Once()

		accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
		assert.Nil(t, accountResponse)
		assert.EqualError(t, err, blockErr.Error())

		invalidHash := "invalid hash"
		blockFilter = &types.PartialBlockIdentifier{Hash: &invalidHash}
		accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
		assert.Nil(t, accountResponse)
		assert.Contains(t, err.Error(), "invalid byte")
	})
}

func TestBalance_CurrencyFilter(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, mockBalanceFactory, client := setupClient(t)

	testAccount, _ := newTestAccount(t)
	acc := &types.AccountIdentifier{Address: testAccount.Address}
	_, resultBlock := newBlockWithResult(t)
	resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
	mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(resultBlockResults, nil)
	mockRPCClient.On("Block", ctx, &resultBlockResults.Height).Return(resultBlock, nil)

	mockBalanceService := &mocks.AccountBalanceService{}
	mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, testAccount.Address), &resultBlock.Block.Header).Return(
		mockBalanceService,
		nil,
	)
	coins := generateDefaultCoins()
	mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(coins, uint64(0), nil)

	t.Run("all supported coins are returned by default", func(t *testing.T) {
		accountResponse, err := client.Balance(ctx, acc, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, len(accountResponse.Balances), 4)
		assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "HARD"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "SWP"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))
	})

	t.Run("filter by single currency", func(t *testing.T) {
		filter := []*types.Currency{
			kava.Currencies["ukava"],
		}
		accountResponse, err := client.Balance(ctx, acc, nil, filter)
		require.NoError(t, err)
		assert.Equal(t, len(accountResponse.Balances), 1)
		assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
		assert.Nil(t, getBalance(accountResponse.Balances, "HARD"))
		assert.Nil(t, getBalance(accountResponse.Balances, "SWP"))
		assert.Nil(t, getBalance(accountResponse.Balances, "USDX"))
	})

	t.Run("filter by multiple currencies", func(t *testing.T) {
		filter := []*types.Currency{
			kava.Currencies["ukava"],
			kava.Currencies["usdx"],
		}
		accountResponse, err := client.Balance(ctx, acc, nil, filter)
		require.NoError(t, err)
		assert.Equal(t, len(accountResponse.Balances), 2)
		assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
		assert.Nil(t, getBalance(accountResponse.Balances, "HARD"))
		assert.Nil(t, getBalance(accountResponse.Balances, "SWP"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))
	})

	t.Run("filter by all supported coins", func(t *testing.T) {
		filter := []*types.Currency{
			kava.Currencies["ukava"],
			kava.Currencies["hard"],
			kava.Currencies["swp"],
			kava.Currencies["usdx"],
		}
		accountResponse, err := client.Balance(ctx, acc, nil, filter)
		require.NoError(t, err)
		assert.Equal(t, len(accountResponse.Balances), 4)
		assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "HARD"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "SWP"))
		assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))
	})
}

func TestBalance_DefaultZeroCurrency(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, mockBalanceFactory, client := setupClient(t)

	emptyTestAccount, emptyAccountBalance := newEmptyTestAccount(t)
	partialTestAccount, partialAccountBalance := newPartialTestAccount(t)

	_, resultBlock := newBlockWithResult(t)
	resultBlockResults := &ctypes.ResultBlockResults{Height: resultBlock.Block.Height}
	mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(resultBlockResults, nil)
	mockRPCClient.On("Block", ctx, &resultBlockResults.Height).Return(resultBlock, nil)

	mockBalanceService := &mocks.AccountBalanceService{}
	mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(emptyAccountBalance, emptyTestAccount.GetSequence(), nil).Once()

	mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, emptyTestAccount.Address), &resultBlock.Block.Header).Return(
		mockBalanceService,
		nil,
	).Once()

	acc := &types.AccountIdentifier{Address: emptyTestAccount.Address}
	accountResponse, err := client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 4)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "KAVA").Value)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "HARD").Value)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "SWP").Value)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "USDX").Value)
	assert.Equal(t, emptyTestAccount.GetSequence(), accountResponse.Metadata["sequence"])

	mockBalanceService.On("GetCoinsAndSequenceForSubAccount", ctx, (*types.SubAccountIdentifier)(nil)).Return(partialAccountBalance, partialTestAccount.GetSequence(), nil).Once()
	mockBalanceFactory.On("Execute", ctx, mustAccAddrFromStr(t, partialTestAccount.Address), &resultBlock.Block.Header).Return(
		mockBalanceService,
		nil,
	).Once()

	// test that partial account returns zero balances for unspecified coins
	acc = &types.AccountIdentifier{Address: partialTestAccount.Address}
	accountResponse, err = client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 4)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "KAVA").Value)
	assert.NotEqual(t, "0", getBalance(accountResponse.Balances, "HARD").Value)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "SWP").Value)
	assert.Equal(t, "0", getBalance(accountResponse.Balances, "USDX").Value)
	assert.Equal(t, partialTestAccount.GetSequence(), accountResponse.Metadata["sequence"])
}

func TestBlock_Info_NoTransactions(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)

	genesisBlockIdentifier := &types.BlockIdentifier{
		Index: 1,
		Hash:  "ADB03E823AFC5F12DC02D984A7E1E0EC47E84FC323005B82FB0B3A9DC8F045B7",
	}
	genesisBlockTime := time.Now().Add(-800 * time.Second)
	genesisHashBytes, err := hex.DecodeString(genesisBlockIdentifier.Hash)
	require.NoError(t, err)

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: 99,
		Hash:  "8EA67B6F7927DB941F86501D1757AC6804C1D21B7A75B9DA3F16A3C81C397E50",
	}
	parentHashBytes, err := hex.DecodeString(parentBlockIdentifier.Hash)
	require.NoError(t, err)

	blockIdentifier := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(blockIdentifier.Hash)
	require.NoError(t, err)

	mockGenesisResultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: genesisHashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: genesisBlockIdentifier.Index,
				Time:   genesisBlockTime,
			},
		},
	}

	mockResultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: hashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: blockIdentifier.Index,
				Time:   blockTime,
				LastBlockID: tmtypes.BlockID{
					Hash: parentHashBytes,
				},
			},
		},
	}

	mockBlockErr := errors.New("some block error")

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(
		mockResultBlock,
		nil,
	).Once()

	mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(
		&ctypes.ResultBlockResults{Height: blockIdentifier.Index},
		nil,
	)

	blockResponse, err := client.Block(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, blockIdentifier, blockResponse.Block.BlockIdentifier)
	assert.Equal(t, parentBlockIdentifier, blockResponse.Block.ParentBlockIdentifier)
	assert.Equal(t, blockTime.UnixNano()/int64(1e6), blockResponse.Block.Timestamp)
	assert.Equal(t, 0, len(blockResponse.Block.Transactions))
	assert.Nil(t, blockResponse.OtherTransactions)

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(
		nil,
		mockBlockErr,
	).Once()

	blockResponse, err = client.Block(ctx, nil)
	assert.Nil(t, blockResponse)
	assert.Equal(t, err, mockBlockErr)

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(
		mockResultBlock,
		nil,
	).Once()
	mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(
		&ctypes.ResultBlockResults{Height: blockIdentifier.Index},
		nil,
	)

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Index: &blockIdentifier.Index,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, blockIdentifier, blockResponse.Block.BlockIdentifier)
	assert.Equal(t, parentBlockIdentifier, blockResponse.Block.ParentBlockIdentifier)
	assert.Equal(t, blockTime.UnixNano()/int64(1e6), blockResponse.Block.Timestamp)
	assert.Equal(t, 0, len(blockResponse.Block.Transactions))
	assert.Nil(t, blockResponse.OtherTransactions)

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(
		nil,
		mockBlockErr,
	).Once()

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Index: &blockIdentifier.Index,
		},
	)
	assert.Nil(t, blockResponse)
	assert.Equal(t, err, mockBlockErr)

	mockRPCClient.On("BlockByHash", ctx, hashBytes).Return(
		mockResultBlock,
		nil,
	).Once()

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Hash: &blockIdentifier.Hash,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, blockIdentifier, blockResponse.Block.BlockIdentifier)
	assert.Equal(t, parentBlockIdentifier, blockResponse.Block.ParentBlockIdentifier)
	assert.Equal(t, blockTime.UnixNano()/int64(1e6), blockResponse.Block.Timestamp)
	assert.Equal(t, 0, len(blockResponse.Block.Transactions))
	assert.Nil(t, blockResponse.OtherTransactions)

	mockRPCClient.On("BlockByHash", ctx, hashBytes).Return(
		nil,
		mockBlockErr,
	).Once()

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Hash: &blockIdentifier.Hash,
		},
	)
	assert.Nil(t, blockResponse)
	assert.Equal(t, err, mockBlockErr)

	mockRPCClient.On("Block", ctx, &genesisBlockIdentifier.Index).Return(
		mockGenesisResultBlock,
		nil,
	).Once()

	mockRPCClient.On("BlockResults", ctx, &genesisBlockIdentifier.Index).Return(
		&ctypes.ResultBlockResults{},
		nil,
	)

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Index: &genesisBlockIdentifier.Index,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, genesisBlockIdentifier, blockResponse.Block.BlockIdentifier)
	assert.Equal(t, genesisBlockIdentifier, blockResponse.Block.ParentBlockIdentifier)
	assert.Equal(t, genesisBlockTime.UnixNano()/int64(1e6), blockResponse.Block.Timestamp)
	assert.Equal(t, 0, len(blockResponse.Block.Transactions))
	assert.Nil(t, blockResponse.OtherTransactions)

	mockRPCClient.On("BlockByHash", ctx, genesisHashBytes).Return(
		mockGenesisResultBlock,
		nil,
	).Once()

	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Hash: &genesisBlockIdentifier.Hash,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, genesisBlockIdentifier, blockResponse.Block.BlockIdentifier)
	assert.Equal(t, genesisBlockIdentifier, blockResponse.Block.ParentBlockIdentifier)
	assert.Equal(t, genesisBlockTime.UnixNano()/int64(1e6), blockResponse.Block.Timestamp)
	assert.Equal(t, 0, len(blockResponse.Block.Transactions))
	assert.Nil(t, blockResponse.OtherTransactions)

	invalidHash := "invalid hash"
	blockResponse, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Hash: &invalidHash,
		},
	)
	assert.Nil(t, blockResponse)
	assert.Contains(t, err.Error(), "invalid byte")
}

func TestBlock_Transactions(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)
	encodingConfig := app.MakeEncodingConfig()

	txBuilder1 := encodingConfig.TxConfig.NewTxBuilder()
	err := txBuilder1.SetMsgs(&banktypes.MsgSend{
		FromAddress: sdk.AccAddress("test from address").String(),
		ToAddress:   sdk.AccAddress("test to address").String(),
		Amount:      sdk.Coins{sdk.NewCoin("ukava", sdk.NewInt(100))},
	})
	require.NoError(t, err)
	txBuilder1.SetGasLimit(100000)
	txBuilder1.SetFeeAmount(sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(5000)}})
	txBuilder1.SetMemo("mock transaction 1")

	var rawMockTx1 tmtypes.Tx
	rawMockTx1, err = encodingConfig.TxConfig.TxEncoder()(txBuilder1.GetTx())
	require.NoError(t, err)
	mockDeliverTx1 := &abci.ResponseDeliverTx{
		Code: 0,
		Log: sdk.ABCIMessageLogs{
			sdk.NewABCIMessageLog(0, "", []sdk.Event{}),
		}.String(),
	}

	txBuilder2 := encodingConfig.TxConfig.NewTxBuilder()
	err = txBuilder2.SetMsgs(&banktypes.MsgSend{
		FromAddress: sdk.AccAddress("test from address").String(),
		ToAddress:   sdk.AccAddress("test to address").String(),
		Amount:      sdk.Coins{sdk.NewCoin("ukava", sdk.NewInt(200))},
	})
	require.NoError(t, err)
	txBuilder2.SetGasLimit(200000)
	txBuilder2.SetFeeAmount(sdk.Coins{sdk.Coin{Denom: "ukava", Amount: sdk.NewInt(5000)}})
	txBuilder2.SetMemo("mock transaction 2")

	var rawMockTx2 tmtypes.Tx
	rawMockTx2, err = encodingConfig.TxConfig.TxEncoder()(txBuilder2.GetTx())
	require.NoError(t, err)
	mockDeliverTx2 := &abci.ResponseDeliverTx{
		Code: 1,
		Log:  "some error message",
	}

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: 99,
		Hash:  "8EA67B6F7927DB941F86501D1757AC6804C1D21B7A75B9DA3F16A3C81C397E50",
	}
	parentHashBytes, err := hex.DecodeString(parentBlockIdentifier.Hash)
	require.NoError(t, err)

	blockIdentifier := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(blockIdentifier.Hash)
	require.NoError(t, err)

	mockRawTransactions := []tmtypes.Tx{rawMockTx1, rawMockTx2}
	mockResultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: hashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: blockIdentifier.Index,
				Time:   blockTime,
				LastBlockID: tmtypes.BlockID{
					Hash: parentHashBytes,
				},
			},
			Data: tmtypes.Data{
				Txs: mockRawTransactions,
			},
		},
	}

	mockDeliverTxs := []*abci.ResponseDeliverTx{mockDeliverTx1, mockDeliverTx2}
	mockResultBlockResults := &ctypes.ResultBlockResults{
		TxsResults:       mockDeliverTxs,
		BeginBlockEvents: []abci.Event{abci.Event{}},
		EndBlockEvents:   []abci.Event{abci.Event{}},
	}

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(mockResultBlock, nil).Once()
	mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(mockResultBlockResults, nil).Once()

	blockResponse, err := client.Block(ctx, &types.PartialBlockIdentifier{Index: &blockIdentifier.Index})
	require.NoError(t, err)
	assert.Equal(t, 2, len(blockResponse.Block.Transactions))

	for i, tx := range blockResponse.Block.Transactions {
		mockRawTx := mockRawTransactions[i]
		mockDeliverTx := mockDeliverTxs[i]

		expectedHash := strings.ToUpper(hex.EncodeToString(mockRawTx.Hash()))
		assert.Equal(t, expectedHash, tx.TransactionIdentifier.Hash)

		assert.Greater(t, len(tx.Operations), 1)

		if mockDeliverTx.Code != 0 {
			expectedLog := mockDeliverTx.Log
			actualLog, ok := tx.Metadata["log"].(string)
			require.True(t, ok)
			assert.Equal(t, expectedLog, actualLog)
		} else {
			_, exists := tx.Metadata["log"]
			assert.False(t, exists)
		}

		for index, operation := range tx.Operations {
			currentIndex := int64(index)
			assert.Equal(t, currentIndex, operation.OperationIdentifier.Index)

			if mockDeliverTx.Code == 0 || operation.Type == kava.FeeOpType {
				assert.Equal(t, kava.SuccessStatus, *operation.Status)
			} else {
				assert.Equal(t, kava.FailureStatus, *operation.Status)
			}

			for _, relatedOperation := range operation.RelatedOperations {
				assert.Greater(t, currentIndex, relatedOperation.Index)
			}
		}
	}

	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(mockResultBlock, nil).Once()
	rpcErr := errors.New("block results error")
	mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(nil, rpcErr).Once()

	blockResponse, err = client.Block(ctx, &types.PartialBlockIdentifier{Index: &blockIdentifier.Index})
	assert.Nil(t, blockResponse)
	assert.Error(t, err)
	assert.Equal(t, rpcErr, err)

	badTx := tmtypes.Tx("invalid tx")
	mockResultBlock.Block.Data.Txs = []tmtypes.Tx{badTx}
	mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(mockResultBlock, nil).Once()
	mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(mockResultBlockResults, nil).Once()

	assert.Panics(t, func() {
		_, _ = client.Block(ctx, &types.PartialBlockIdentifier{Index: &blockIdentifier.Index})
	})
}

func TestBlock_BlockResultsRetry(t *testing.T) {
	parentBlockIdentifier := &types.BlockIdentifier{
		Index: 99,
		Hash:  "8EA67B6F7927DB941F86501D1757AC6804C1D21B7A75B9DA3F16A3C81C397E50",
	}
	parentHashBytes, err := hex.DecodeString(parentBlockIdentifier.Hash)
	require.NoError(t, err)

	blockIdentifier := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(blockIdentifier.Hash)
	require.NoError(t, err)

	mockResultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: hashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: blockIdentifier.Index,
				Time:   blockTime,
				LastBlockID: tmtypes.BlockID{
					Hash: parentHashBytes,
				},
			},
		},
	}

	abciErr := tmstate.ErrNoABCIResponsesForHeight{Height: blockIdentifier.Index + 1}
	rpcErr := tmrpctypes.RPCInternalError(tmrpctypes.JSONRPCIntID(1), abciErr).Error
	mockErr := fmt.Errorf("Block Result: %w", rpcErr)

	t.Run("retries if there are no abci results yet", func(t *testing.T) {
		ctx := context.Background()
		mockRPCClient, _, client := setupClient(t)

		mockRPCClient.On("Block", ctx, &blockIdentifier.Index).Return(
			mockResultBlock,
			nil,
		).Once()

		mockRPCClient.On("BlockResults", ctx, (*int64)(nil)).Return(
			nil,
			mockErr,
		).Once()

		mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(
			&ctypes.ResultBlockResults{Height: blockIdentifier.Index},
			nil,
		).Once()

		_, err = client.Block(ctx, nil)
		require.NoError(t, err)
	})
}

func TestAccount(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)
	addr, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	require.NoError(t, err)

	accErr := errors.New("error retrieving account")
	mockRPCClient.On("Account", ctx, addr, int64(0)).Return(nil, accErr).Once()

	account, err := client.Account(ctx, addr)
	assert.Nil(t, account)
	assert.EqualError(t, err, accErr.Error())

	expectedAccount := &authtypes.BaseAccount{}
	mockRPCClient.On("Account", ctx, addr, int64(0)).Return(expectedAccount, nil).Once()

	ctx = context.Background()
	account, err = client.Account(ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, expectedAccount, account)
}

// TestBlock_HashPriority asserts that a block is looked up by hash if both
// the index and hash are provided.  In addition, we assert that an error
// is thrown if the returned index does not match the queried index.
func TestBlock_HashPriority(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)

	parentBlockIdentifier := &types.BlockIdentifier{
		Index: 99,
		Hash:  "8EA67B6F7927DB941F86501D1757AC6804C1D21B7A75B9DA3F16A3C81C397E50",
	}
	parentHashBytes, err := hex.DecodeString(parentBlockIdentifier.Hash)
	require.NoError(t, err)

	blockIdentifier := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(blockIdentifier.Hash)
	require.NoError(t, err)

	mockResultBlock := &ctypes.ResultBlock{
		BlockID: tmtypes.BlockID{
			Hash: hashBytes,
		},
		Block: &tmtypes.Block{
			Header: tmtypes.Header{
				Height: blockIdentifier.Index,
				Time:   blockTime,
				LastBlockID: tmtypes.BlockID{
					Hash: parentHashBytes,
				},
			},
		},
	}

	// block is only ever looked up by hash and not the index
	mockRPCClient.On("BlockByHash", ctx, hashBytes).Return(
		mockResultBlock,
		nil,
	)
	mockRPCClient.On("BlockResults", ctx, &blockIdentifier.Index).Return(
		&ctypes.ResultBlockResults{},
		nil,
	)

	blockResponse, err := client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Index: &blockIdentifier.Index,
			Hash:  &blockIdentifier.Hash,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, blockIdentifier, blockResponse.Block.BlockIdentifier)

	invalidIndex := blockIdentifier.Index + 1
	_, err = client.Block(
		ctx,
		&types.PartialBlockIdentifier{
			Index: &invalidIndex,
			Hash:  &blockIdentifier.Hash,
		},
	)
	assert.EqualError(t, err, fmt.Sprintf("requested index %d does not match returned index %d", invalidIndex, blockIdentifier.Index))

	mockRPCClient.AssertExpectations(t)
}

func TestEstimateGas(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)

	addr1, err := sdk.AccAddressFromBech32("kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea")
	require.NoError(t, err)
	addr2, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	msgs := []sdk.Msg{
		&banktypes.MsgSend{
			FromAddress: addr1.String(),
			ToAddress:   addr2.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000))),
		},
		&banktypes.MsgSend{
			FromAddress: addr1.String(),
			ToAddress:   addr2.String(),
			Amount:      sdk.NewCoins(sdk.NewCoin("hard", sdk.NewInt(2000000))),
		},
	}

	tx := legacytx.NewStdTx(
		msgs,
		legacytx.StdFee{},
		nil,
		"a memo",
	)
	gasAdjusment := float64(0.1)

	simErr := errors.New("could not simulate tx")
	mockRPCClient.On("SimulateTx", ctx, &tx).Return(nil, simErr).Once()

	gas, err := client.EstimateGas(ctx, &tx, gasAdjusment)
	assert.Equal(t, uint64(0), gas)
	assert.EqualError(t, err, simErr.Error())

	gasUsed := uint64(200000)
	simResp := &sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{
			GasWanted: 100000,
			GasUsed:   gasUsed,
		},
	}

	mockRPCClient.On("SimulateTx", ctx, &tx).Return(simResp, nil).Once()
	ctx = context.Background()
	gas, err = client.EstimateGas(ctx, &tx, gasAdjusment)
	require.Nil(t, err)
	assert.Equal(t, uint64(220000), gas)
}

func TestPostTx(t *testing.T) {
	ctx := context.Background()
	mockRPCClient, _, client := setupClient(t)

	txjson, err := ioutil.ReadFile("test-fixtures/txs/msg-send.json")
	require.NoError(t, err)

	cdc := app.MakeEncodingConfig().Amino
	var stdtx legacytx.StdTx
	err = cdc.UnmarshalJSON(txjson, &stdtx)
	require.NoError(t, err)

	txBytes, err := cdc.MarshalLengthPrefixed(stdtx)
	require.NoError(t, err)

	rpcErr := errors.New("some rpc error")
	mockRPCClient.On("BroadcastTxSync", ctx, tmtypes.Tx(txBytes)).Return(nil, rpcErr).Once()

	response, err := client.PostTx(context.Background(), txBytes)
	assert.Nil(t, response)
	assert.Equal(t, rpcErr, err)

	txResult := &ctypes.ResultBroadcastTx{
		Code: abci.CodeTypeOK,
		Hash: tmtypes.Tx(txBytes).Hash(),
	}
	mockRPCClient.On("BroadcastTxSync", ctx, tmtypes.Tx(txBytes)).Return(txResult, nil).Once()

	response, err = client.PostTx(context.Background(), txBytes)
	require.NoError(t, err)
	assert.Equal(t, "3A3FB732715054D56313354A8CEB135848363AB97E9323559E419F3D09BA4B31", response.Hash)

	txResult = &ctypes.ResultBroadcastTx{
		Code: 4,
		Hash: tmtypes.Tx(txBytes).Hash(),
		Log:  "some tx error",
	}
	mockRPCClient.On("BroadcastTxSync", ctx, tmtypes.Tx(txBytes)).Return(txResult, nil).Once()

	response, err = client.PostTx(context.Background(), txBytes)
	require.Nil(t, response)
	assert.EqualError(t, err, "some tx error")
}
