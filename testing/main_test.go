//go:build integration
// +build integration

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

package testing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	rclient "github.com/coinbase/rosetta-sdk-go/client"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttpclient "github.com/cometbft/cometbft/rpc/client/http"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	kava "github.com/kava-labs/kava/app"
	"github.com/kava-labs/rosetta-kava/configuration"
	router "github.com/kava-labs/rosetta-kava/server"
)

// Rosetta Server
var config *configuration.Configuration
var server *httptest.Server

// Rosetta Client
var client *rclient.APIClient

// Tendermint RPC
var cdc codec.Codec
var interfaceRegistry codectypes.InterfaceRegistry
var rpc rpcclient.Client

// Test Settings
var testAccountAddress string

// TestMain loads integration env and runs test
func TestMain(m *testing.M) {
	configLoader := &configuration.EnvLoader{}

	var err error
	config, err = configuration.LoadConfig(configLoader)
	if err != nil {
		fmt.Println(fmt.Errorf("%w: unable to load configuration", err))
		os.Exit(1)
	}

	if config.Mode.String() != os.Getenv("MODE") {
		fmt.Println("MODE was not loaded correctly")
		os.Exit(1)
	}

	if config.NetworkIdentifier.Network != os.Getenv("NETWORK") {
		fmt.Println("NETWORK was not loaded correctly")
		os.Exit(1)
	}

	handler, err := router.NewRouter(config)
	if err != nil {
		fmt.Println(fmt.Errorf("%w: unable to initialize router", err))
		os.Exit(1)
	}

	server = httptest.NewServer(handler)
	defer server.Close()

	clientConfig := rclient.NewConfiguration(
		server.URL,
		"kava-test-client",
		&http.Client{
			Timeout: 10 * time.Second,
		},
	)

	encodingConfig := kava.MakeEncodingConfig()
	cdc = encodingConfig.Marshaler
	interfaceRegistry = encodingConfig.InterfaceRegistry

	client = rclient.NewAPIClient(clientConfig)

	rpc, err = rpchttpclient.New(config.KavaRPCURL, "/websocket")
	if err != nil {
		fmt.Println(fmt.Errorf("%w: could not initialize http client", err))
		os.Exit(1)
	}

	testAccountAddress = os.Getenv("TEST_KAVA_ADDRESS")
	if testAccountAddress == "" {
		testAccountAddress = "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	}

	os.Exit(m.Run())
}

// GetAccount gets an account
func GetAccount(address string, height int64) (authtypes.AccountI, error) {
	addr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	bz, err := cdc.Marshal(&authtypes.QueryAccountRequest{Address: addr.String()})
	if err != nil {
		return nil, err
	}

	path := "/cosmos.auth.v1beta1.Query/Account"
	opts := rpcclient.ABCIQueryOptions{Height: height, Prove: false}

	result, err := ParseABCIResult(rpc.ABCIQueryWithOptions(context.Background(), path, bz, opts))
	if err != nil {
		return nil, err
	}

	var resp authtypes.QueryAccountResponse
	err = cdc.Unmarshal(result, &resp)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = interfaceRegistry.UnpackAny(resp.Account, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetGalance returns the owned coins of an account at a specified height
func GetBalance(address sdktypes.AccAddress, height int64) (sdktypes.Coins, error) {
	path := "/cosmos.bank.v1beta1.Query/AllBalances"
	opts := rpcclient.ABCIQueryOptions{Height: height, Prove: false}
	request := banktypes.QueryAllBalancesRequest{
		Address:    address.String(),
		Pagination: &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	}

	totalBalances := sdk.NewCoins()

	for {
		bz, err := cdc.Marshal(&request)
		if err != nil {
			return nil, err
		}

		result, err := ParseABCIResult(rpc.ABCIQueryWithOptions(context.Background(), path, bz, opts))
		if err != nil {
			return nil, err
		}

		var resp banktypes.QueryAllBalancesResponse
		err = cdc.Unmarshal(result, &resp)
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
