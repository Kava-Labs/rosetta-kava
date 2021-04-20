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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func makeTestAccount(t *testing.T) *authtypes.BaseAccount {
	addr, err := sdk.AccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
	require.NoError(t, err)

	return &authtypes.BaseAccount{
		Address: addr,
		Coins: sdk.NewCoins(
			sdk.NewCoin("ukava", sdk.NewInt(100)),
			sdk.NewCoin("hard", sdk.NewInt(200)),
			sdk.NewCoin("usdx", sdk.NewInt(300)),
			sdk.NewCoin("bnb", sdk.NewInt(10)),
			sdk.NewCoin("btcb", sdk.NewInt(1)),
			sdk.NewCoin("busd", sdk.NewInt(1000)),
		),
		AccountNumber: 2,
		Sequence:      5,
	}
}

func makeTestVestingAccount(t *testing.T, endTime time.Time) *vestingtypes.DelayedVestingAccount {
	baseAccount := makeTestAccount(t)
	vestingAccount := vestingtypes.NewDelayedVestingAccount(baseAccount, endTime.Unix())

	return vestingAccount
}

func getBalance(balances []*types.Amount, symbol string) *types.Amount {
	for _, balance := range balances {
		if balance.Currency.Symbol == symbol {
			return balance
		}
	}

	return nil
}

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

func TestBalance_InvalidAddress(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	invalidAcc := &types.AccountIdentifier{
		Address: "invalid",
	}

	accountResponse, err := client.Balance(ctx, invalidAcc, nil, nil)
	assert.Nil(t, accountResponse)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestBalance_LatestBlock(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	testAccount := makeTestAccount(t)

	acc := &types.AccountIdentifier{
		Address: testAccount.Address.String(),
	}

	block := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()

	hashBytes, err := hex.DecodeString(block.Hash)
	require.NoError(t, err)

	mockRPCClient.On("Block", (*int64)(nil)).Return(
		nil,
		errors.New("error getting block"),
	).Once()

	accountResponse, err := client.Balance(ctx, acc, nil, nil)
	assert.Nil(t, accountResponse)
	assert.EqualError(t, err, "error getting block")

	mockRPCClient.On("Block", (*int64)(nil)).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   blockTime,
				},
			},
		},
		nil,
	)

	mockRPCClient.On("Account", testAccount.Address, block.Index).Return(
		nil,
		errors.New("account error"),
	).Once()

	accountResponse, err = client.Balance(ctx, acc, nil, nil)
	assert.Nil(t, accountResponse)
	assert.EqualError(t, err, "account error")

	mockRPCClient.On("Account", testAccount.Address, block.Index).Return(
		testAccount,
		nil,
	).Once()

	accountResponse, err = client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)

	// block must be latest block
	assert.Equal(t, block, accountResponse.BlockIdentifier)
	// must have balances set
	assert.Greater(t, len(accountResponse.Balances), 0)

	for _, amount := range accountResponse.Balances {
		denom := Denoms[amount.Currency.Symbol]
		require.NotEmpty(t, denom)

		assert.Equal(t, amount.Currency, Currencies[denom])

		spendableCoins := testAccount.SpendableCoins(blockTime)
		assert.Equal(t, amount.Value, spendableCoins.AmountOf(denom).String())
	}

	mockRPCClient.AssertExpectations(t)
}

func TestBalance_VestingAccount(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	testVestingAccount := makeTestVestingAccount(t, time.Now())
	acc := &types.AccountIdentifier{
		Address: testVestingAccount.Address.String(),
	}

	block := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	hashBytes, err := hex.DecodeString(block.Hash)
	require.NoError(t, err)

	blockTime := time.Unix(testVestingAccount.EndTime, 0).Add(-8 * time.Second)
	vestingBlockTime := time.Unix(testVestingAccount.EndTime, 0).Add(8 * time.Second)

	mockRPCClient.On("Account", testVestingAccount.Address, block.Index).Return(
		testVestingAccount,
		nil,
	)

	mockRPCClient.On("Block", (*int64)(nil)).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   blockTime,
				},
			},
		},
		nil,
	).Once()

	beforeVestingResponse, err := client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)

	for _, amount := range beforeVestingResponse.Balances {
		denom := Denoms[amount.Currency.Symbol]
		require.NotEmpty(t, denom)

		assert.Equal(t, amount.Currency, Currencies[denom])

		spendableCoins := testVestingAccount.SpendableCoins(blockTime)
		assert.Equal(t, amount.Value, spendableCoins.AmountOf(denom).String())
	}

	mockRPCClient.On("Block", (*int64)(nil)).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   vestingBlockTime,
				},
			},
		},
		nil,
	).Once()

	afterVestingResponse, err := client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)

	for _, amount := range afterVestingResponse.Balances {
		denom := Denoms[amount.Currency.Symbol]
		require.NotEmpty(t, denom)

		assert.Equal(t, amount.Currency, Currencies[denom])

		spendableCoins := testVestingAccount.SpendableCoins(vestingBlockTime)
		assert.Equal(t, amount.Value, spendableCoins.AmountOf(denom).String())
	}

	assert.NotEqual(t, beforeVestingResponse, afterVestingResponse)
}

func TestBalance_BlockFilter(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	testAccount := makeTestAccount(t)

	block := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(block.Hash)
	require.NoError(t, err)

	mockRPCClient.On("Account", testAccount.Address, block.Index).Return(
		testAccount,
		nil,
	)

	acc := &types.AccountIdentifier{Address: testAccount.Address.String()}

	mockRPCClient.On("Block", &block.Index).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   blockTime,
				},
			},
		},
		nil,
	).Once()

	blockFilter := &types.PartialBlockIdentifier{Index: &block.Index}
	accountResponse, err := client.Balance(ctx, acc, blockFilter, nil)
	require.NoError(t, err)
	assert.Equal(t, block, accountResponse.BlockIdentifier)

	mockRPCClient.On("Block", &block.Index).Return(
		nil,
		errors.New("some block index error"),
	).Once()

	accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
	assert.Nil(t, accountResponse)
	assert.EqualError(t, err, "some block index error")

	mockRPCClient.On("BlockByHash", hashBytes).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   blockTime,
				},
			},
		},
		nil,
	).Once()

	blockFilter = &types.PartialBlockIdentifier{Hash: &block.Hash}
	accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
	require.NoError(t, err)
	assert.Equal(t, block, accountResponse.BlockIdentifier)

	mockRPCClient.On("BlockByHash", hashBytes).Return(
		nil,
		errors.New("some block hash error"),
	).Once()

	accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
	assert.Nil(t, accountResponse)
	assert.EqualError(t, err, "some block hash error")

	invalidHash := "invalid hash"
	blockFilter = &types.PartialBlockIdentifier{Hash: &invalidHash}
	accountResponse, err = client.Balance(ctx, acc, blockFilter, nil)
	assert.Nil(t, accountResponse)
	assert.Contains(t, err.Error(), "invalid byte")
}

func TestBalance_CurrencyFilter(t *testing.T) {
	ctx := context.Background()
	mockRPCClient := &mocks.Client{}
	client, err := NewClient(mockRPCClient)
	require.NoError(t, err)

	testAccount := makeTestAccount(t)

	block := &types.BlockIdentifier{
		Index: 100,
		Hash:  "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75",
	}
	blockTime := time.Now()
	hashBytes, err := hex.DecodeString(block.Hash)
	require.NoError(t, err)

	mockRPCClient.On("Block", (*int64)(nil)).Return(
		&ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: hashBytes,
			},
			Block: &tmtypes.Block{
				Header: tmtypes.Header{
					Height: block.Index,
					Time:   blockTime,
				},
			},
		},
		nil,
	)

	mockRPCClient.On("Account", testAccount.Address, block.Index).Return(
		testAccount,
		nil,
	)

	// test all coins returned by default
	acc := &types.AccountIdentifier{Address: testAccount.Address.String()}
	accountResponse, err := client.Balance(ctx, acc, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 3)
	assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
	assert.NotNil(t, getBalance(accountResponse.Balances, "HARD"))
	assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))

	// test single coin filter
	filter := []*types.Currency{
		Currencies["ukava"],
	}
	accountResponse, err = client.Balance(ctx, acc, nil, filter)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 1)
	assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
	assert.Nil(t, getBalance(accountResponse.Balances, "HARD"))
	assert.Nil(t, getBalance(accountResponse.Balances, "USDX"))

	// test multi coin filter
	filter = []*types.Currency{
		Currencies["ukava"],
		Currencies["usdx"],
	}
	accountResponse, err = client.Balance(ctx, acc, nil, filter)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 2)
	assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
	assert.Nil(t, getBalance(accountResponse.Balances, "HARD"))
	assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))

	// test all coin filter
	filter = []*types.Currency{
		Currencies["ukava"],
		Currencies["hard"],
		Currencies["usdx"],
	}
	accountResponse, err = client.Balance(ctx, acc, nil, filter)
	require.NoError(t, err)
	assert.Equal(t, len(accountResponse.Balances), 3)
	assert.NotNil(t, getBalance(accountResponse.Balances, "KAVA"))
	assert.NotNil(t, getBalance(accountResponse.Balances, "HARD"))
	assert.NotNil(t, getBalance(accountResponse.Balances, "USDX"))
}
