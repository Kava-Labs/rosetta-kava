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

	"github.com/cometbft/cometbft/libs/bytes"
	tmrpcclient "github.com/cometbft/cometbft/rpc/client"
	tmhttp "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	kava "github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/app/params"
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
	encodingConfig.InterfaceRegistry.RegisterInterface(
		"ibc.lightclients.solomachine.v2.ClientState",
		(*ClientStateI)(nil),
		&ClientState{},
	)

	return &HTTPClient{
		HTTP:           http,
		caller:         rpc,
		cdc:            encodingConfig.Amino,
		encodingConfig: encodingConfig,
	}, nil
}

// Account returns the Account for a given address
func (c *HTTPClient) Account(ctx context.Context, addr sdk.AccAddress, height int64) (authtypes.AccountI, error) {
	bz, err := c.encodingConfig.Marshaler.Marshal(&authtypes.QueryAccountRequest{Address: addr.String()})
	if err != nil {
		return nil, err
	}

	path := "/cosmos.auth.v1beta1.Query/Account"

	data, err := c.abciQuery(ctx, path, bz, height)
	if err != nil {
		return nil, err
	}

	var resp authtypes.QueryAccountResponse
	err = c.encodingConfig.Marshaler.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = c.encodingConfig.InterfaceRegistry.UnpackAny(resp.Account, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Balance returns the Balance for a given address
func (c *HTTPClient) Balance(ctx context.Context, addr sdk.AccAddress, height int64) (sdk.Coins, error) {
	path := "/cosmos.bank.v1beta1.Query/AllBalances"
	totalBalances := sdk.NewCoins()

	request := banktypes.QueryAllBalancesRequest{
		Address:    addr.String(),
		Pagination: &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	}

	for {
		bz, err := c.encodingConfig.Marshaler.Marshal(&request)
		if err != nil {
			return nil, err
		}

		data, err := c.abciQuery(ctx, path, bz, height)
		if err != nil {
			return nil, err
		}

		var resp banktypes.QueryAllBalancesResponse
		err = c.encodingConfig.Marshaler.Unmarshal(data, &resp)
		if err != nil {
			return nil, err
		}

		totalBalances = totalBalances.Add(resp.Balances...)

		if resp.Pagination.NextKey == nil {
			break
		}
		request.Pagination.Key = resp.Pagination.NextKey
	}

	return totalBalances, nil
}

// Delegations returns the delegations for an acc address
func (c *HTTPClient) Delegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.DelegationResponses, error) {
	path := "/cosmos.staking.v1beta1.Query/DelegatorDelegations"
	delegationResponses := stakingtypes.DelegationResponses{}

	request := stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: addr.String(),
		Pagination:    &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	}

	for {
		bz, err := c.encodingConfig.Marshaler.Marshal(&request)
		if err != nil {
			return nil, err
		}

		data, err := c.abciQuery(ctx, path, bz, height)
		if err != nil {
			return nil, err
		}

		var resp stakingtypes.QueryDelegatorDelegationsResponse
		err = c.encodingConfig.Marshaler.Unmarshal(data, &resp)
		if err != nil {
			return nil, err
		}

		delegationResponses = append(delegationResponses, resp.DelegationResponses...)

		if resp.Pagination.NextKey == nil {
			break
		}
		request.Pagination.Key = resp.Pagination.NextKey
	}

	return delegationResponses, nil
}

// UnbondingDelegations returns the unbonding delegations for an address
func (c *HTTPClient) UnbondingDelegations(ctx context.Context, addr sdk.AccAddress, height int64) (stakingtypes.UnbondingDelegations, error) {
	path := "/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations"
	unbondingDelegations := stakingtypes.UnbondingDelegations{}

	request := stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: addr.String(),
		Pagination:    &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	}

	for {
		bz, err := c.encodingConfig.Marshaler.Marshal(&request)
		if err != nil {
			return nil, err
		}

		data, err := c.abciQuery(ctx, path, bz, height)
		if err != nil {
			return nil, err
		}

		var resp stakingtypes.QueryDelegatorUnbondingDelegationsResponse
		err = c.encodingConfig.Marshaler.Unmarshal(data, &resp)
		if err != nil {
			return nil, err
		}

		unbondingDelegations = append(unbondingDelegations, resp.UnbondingResponses...)

		if resp.Pagination.NextKey == nil {
			break
		}
		request.Pagination.Key = resp.Pagination.NextKey
	}

	return unbondingDelegations, nil
}

// SimulateTx simulates a transaction and returns the response containing the gas used and result
func (c *HTTPClient) SimulateTx(ctx context.Context, tx authsigning.Tx) (*sdk.SimulationResponse, error) {
	bz, err := c.encodingConfig.TxConfig.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	data, err := c.abciQuery(ctx, "/app/simulate", bz, 0)
	if err != nil {
		return nil, err
	}

	var simRes sdk.SimulationResponse
	if err := c.encodingConfig.Marshaler.UnmarshalJSON(data, &simRes); err != nil {
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
