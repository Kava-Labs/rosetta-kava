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
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	kava "github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
	"github.com/tendermint/tendermint/libs/bytes"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// HTTPClient extends the tendermint http client to enable finding blocks by hash
type HTTPClient struct {
	*tmhttp.HTTP
	caller         *tmclient.Client
	cdc            *codec.LegacyAmino
	encodingConfig params.EncodingConfig
}

// NewHTTPClient returns a new HTTPClient with additional capabilities
func NewHTTPClient(remote string) (*HTTPClient, error) {
	client, err := tmclient.DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	http, err := tmhttp.NewWithClient(remote, "/websocket", client)
	if err != nil {
		return nil, err
	}

	rpc, err := tmclient.NewWithHTTPClient(remote, client)
	if err != nil {
		return nil, err
	}

	encodingConfig := kava.MakeEncodingConfig()

	return &HTTPClient{
		HTTP:           http,
		caller:         rpc,
		cdc:            encodingConfig.Amino,
		encodingConfig: encodingConfig,
	}, nil
}

// Account returns the Account for a given address
func (c *HTTPClient) Account(ctx context.Context, addr sdk.AccAddress, height int64) (authtypes.AccountI, error) {
	bz, err := c.cdc.MarshalJSON(authtypes.QueryAccountRequest{Address: addr.String()})
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", authtypes.QuerierRoute, authtypes.QueryAccount)

	data, err := c.abciQuery(ctx, path, bz, height)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = c.cdc.UnmarshalJSON(data, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Balance returns the Balance for a given address
func (c *HTTPClient) Balance(ctx context.Context, addr sdk.AccAddress, height int64) (sdk.Coins, error) {
	// legacy querier does not paginate -- Pagination parameter set to nil to ignore
	bz, err := c.cdc.MarshalJSON(banktypes.NewQueryAllBalancesRequest(addr, nil))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", banktypes.QuerierRoute, banktypes.QueryAllBalances)

	data, err := c.abciQuery(ctx, path, bz, height)
	if err != nil {
		return nil, err
	}

	var balance sdk.Coins
	err = c.cdc.UnmarshalJSON(data, &balance)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// Delegations returns the delegations for an acc address
func (c *HTTPClient) Delegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.DelegationResponses, error) {
	bz, err := c.cdc.MarshalJSON(stakingtypes.NewQueryDelegatorParams(addr))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", stakingtypes.QuerierRoute, stakingtypes.QueryDelegatorDelegations)

	data, err := c.abciQuery(ctx, path, bz, height)
	if err != nil {
		return nil, err
	}

	var delegations stakingtypes.DelegationResponses
	err = c.cdc.UnmarshalJSON(data, &delegations)
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// UnbondingDelegations returns the unbonding delegations for an address
func (c *HTTPClient) UnbondingDelegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.UnbondingDelegations, error) {
	bz, err := c.cdc.MarshalJSON(stakingtypes.NewQueryDelegatorParams(addr))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", stakingtypes.QuerierRoute, stakingtypes.QueryDelegatorUnbondingDelegations)

	data, err := c.abciQuery(ctx, path, bz, height)
	if err != nil {
		return nil, err
	}

	var unbondingDelegations stakingtypes.UnbondingDelegations
	err = c.cdc.UnmarshalJSON(data, &unbondingDelegations)
	if err != nil {
		return nil, err
	}

	return unbondingDelegations, nil
}

// SimulateTx simulates a transaction and returns the response containing the gas used and result
func (c *HTTPClient) SimulateTx(ctx context.Context, tx sdk.Tx) (*sdk.SimulationResponse, error) {
	bz, err := c.encodingConfig.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	data, err := c.abciQuery(ctx, "/app/simulate", bz, 0)
	if err != nil {
		return nil, err
	}

	var simRes sdk.SimulationResponse
	if err := c.encodingConfig.Marshaler.Unmarshal(data, &simRes); err != nil {
		return nil, err
	}
	return &simRes, nil
}

func (c *HTTPClient) abciQuery(ctx context.Context, path string, data bytes.HexBytes, height int64) ([]byte, error) {
	opts := tmrpcclient.ABCIQueryOptions{Height: height, Prove: false}
	result, err := c.ABCIQueryWithOptions(ctx, path, data, opts)
	return ParseABCIResult(result, err)
}

// ParseABCIResult returns the Value of a ABCI Query
func ParseABCIResult(result *ctypes.ResultABCIQuery, err error) ([]byte, error) {
	if err != nil {
		return []byte{}, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return []byte{}, errors.New(resp.Log)
	}

	value := result.Response.GetValue()
	if value == nil {
		return []byte{}, nil
	}

	return value, nil
}
