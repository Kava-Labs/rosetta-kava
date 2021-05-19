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
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	kava "github.com/kava-labs/kava/app"
	pkgerrors "github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/bytes"
	tmrpcclient "github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// HTTPClient extends the tendermint http client to enable finding blocks by hash
type HTTPClient struct {
	*tmhttp.HTTP
	caller *tmclient.Client
	cdc    *codec.Codec
}

// NewHTTPClient returns a new HTTPClient with BlockByHash capabilities
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
	// set codec for tendermint rpc
	cdc := rpc.Codec()
	ctypes.RegisterAmino(cdc)
	rpc.SetCodec(cdc)

	// codec for cosmos-sdk/app level (Account, etc)
	kavaCdc := kava.MakeCodec()

	return &HTTPClient{
		HTTP:   http,
		caller: rpc,
		cdc:    kavaCdc,
	}, nil
}

// BlockByHash fetches a block by it's hash value and return the resulting block
func (c *HTTPClient) BlockByHash(hash []byte) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	_, err := c.caller.Call("block_by_hash", map[string]interface{}{"hash": hash}, result)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "BlockByHash")
	}
	return result, nil
}

// Account returns the Account for a given address
func (c *HTTPClient) Account(addr sdk.AccAddress, height int64) (authexported.Account, error) {
	bz, err := c.cdc.MarshalJSON(authtypes.NewQueryAccountParams(addr))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", authtypes.QuerierRoute, authtypes.QueryAccount)

	data, err := c.abciQuery(path, bz, height)
	if err != nil {
		return nil, err
	}

	var account authexported.Account
	err = c.cdc.UnmarshalJSON(data, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Delegations returns the delegations for an acc address
func (c *HTTPClient) Delegations(addr sdk.AccAddress, height int64) (staking.DelegationResponses, error) {
	bz, err := c.cdc.MarshalJSON(staking.NewQueryDelegatorParams(addr))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", staking.QuerierRoute, staking.QueryDelegatorDelegations)

	data, err := c.abciQuery(path, bz, height)
	if err != nil {
		return nil, err
	}

	var delegations staking.DelegationResponses
	err = c.cdc.UnmarshalJSON(data, &delegations)
	if err != nil {
		return nil, err
	}

	return delegations, nil
}

// UnbondingDelegations returns the unbonding delegations for an address
func (c *HTTPClient) UnbondingDelegations(addr sdk.AccAddress, height int64) (staking.UnbondingDelegations, error) {
	bz, err := c.cdc.MarshalJSON(staking.NewQueryDelegatorParams(addr))
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("custom/%s/%s", staking.QuerierRoute, staking.QueryDelegatorUnbondingDelegations)

	data, err := c.abciQuery(path, bz, height)
	if err != nil {
		return nil, err
	}

	var unbondingDelegations staking.UnbondingDelegations
	err = c.cdc.UnmarshalJSON(data, &unbondingDelegations)
	if err != nil {
		return nil, err
	}

	return unbondingDelegations, nil
}

func (c *HTTPClient) abciQuery(path string, data bytes.HexBytes, height int64) ([]byte, error) {
	opts := tmrpcclient.ABCIQueryOptions{Height: height, Prove: false}
	result, err := c.ABCIQueryWithOptions(path, data, opts)
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
