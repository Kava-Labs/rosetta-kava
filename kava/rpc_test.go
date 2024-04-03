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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	app "github.com/kava-labs/kava/app"
	"github.com/kava-labs/rosetta-kava/kava"
	"github.com/tendermint/go-amino"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	jsonrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
	tmtypes "github.com/cometbft/cometbft/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAddr = sdk.MustAccAddressFromBech32("kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq")
var accountNotFound = fmt.Sprintf("rpc error: code = NotFound desc = account %s not found: key not found", testAddr.String())

// abciRequestQuery ensures that height & data can properly
// encode & decode across json rpc boundaries
type abciRequestQuery struct {
	Height string
	Path   string
	Data   bytes.HexBytes
	Prove  bool
}

type jsonRPCHandler func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse

// abciQueryCall provides an expectation and a response for an abci query
// request done over tendermint abci
type abciQueryCall struct {
	expectedQuery abciRequestQuery
	responseQuery abcitypes.ResponseQuery
}

// newABCIQueryHandler creates a handler with mockCalls provided in order
// they should happen.
func newABCIQueryHandler(
	t *testing.T,
	mockCalls []abciQueryCall,
) jsonRPCHandler {
	callIndex := 0

	return func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		if callIndex >= len(mockCalls) {
			t.Errorf("unexpected number of calls")
		}
		defer func() { callIndex++ }()

		require.Equal(t, "abci_query", request.Method)

		var requestQuery abciRequestQuery
		err := json.Unmarshal(request.Params, &requestQuery)
		require.NoError(t, err)

		require.Equal(t, mockCalls[callIndex].expectedQuery, requestQuery)

		result, err := json.Marshal(ctypes.ResultABCIQuery{Response: mockCalls[callIndex].responseQuery})
		require.NoError(t, err)

		return jsonrpctypes.RPCResponse{request.JSONRPC, request.ID, result, nil}
	}
}

func rpcTestServer(
	t *testing.T,
	rpcHandler func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse,
) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var request jsonrpctypes.RPCRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err)

		response := rpcHandler(request)

		b, err := json.Marshal(&response)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(b)
		require.NoError(t, err)
	}))
}

func TestHTTPClient_BlockByHash(t *testing.T) {
	cdc := amino.NewCodec()

	ts := rpcTestServer(t, func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		assert.Equal(t, "block_by_hash", request.Method)

		var params struct {
			Hash string
		}

		err := json.Unmarshal(request.Params, &params)
		require.NoError(t, err)
		hash, err := base64.StdEncoding.DecodeString(params.Hash)
		require.NoError(t, err)

		result := &ctypes.ResultBlock{
			BlockID: tmtypes.BlockID{
				Hash: []byte(hash),
			},
			Block: &tmtypes.Block{},
		}

		data, err := cdc.MarshalJSON(result)
		require.NoError(t, err)

		var response jsonrpctypes.RPCResponse

		if len(hash) == 0 {
			response = jsonrpctypes.RPCResponse{
				JSONRPC: request.JSONRPC,
				ID:      request.ID,
				Error: &jsonrpctypes.RPCError{
					Code:    1,
					Message: "invalid hash",
				},
			}
		} else {
			response = jsonrpctypes.RPCResponse{
				JSONRPC: request.JSONRPC,
				ID:      request.ID,
				Result:  json.RawMessage(data),
			}
		}

		return response
	})

	defer ts.Close()

	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	testHash := []byte("D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75")
	block, err := client.BlockByHash(context.Background(), testHash)
	assert.NoError(t, err)
	assert.Equal(t, testHash, []byte(block.BlockID.Hash))

	testHash = []byte{}
	block, err = client.BlockByHash(context.Background(), testHash)
	assert.Error(t, err)
	assert.Nil(t, block)
	assert.ErrorContains(t, err, "invalid hash")
}

func TestHTTPClient_Account(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	codec := encodingConfig.Marshaler

	testAccount, _ := newTestAccount(t)
	anyAccount, err := codectypes.NewAnyWithValue(testAccount)
	require.NoError(t, err)

	response := authtypes.QueryAccountResponse{
		Account: anyAccount,
	}
	responseData, err := codec.Marshal(&response)
	require.NoError(t, err)

	height := int64(101)
	requestData, err := codec.Marshal(&authtypes.QueryAccountRequest{Address: testAddr.String()})
	require.NoError(t, err)

	requestQuery := abciRequestQuery{
		Height: strconv.FormatInt(height, 10),
		Path:   "/cosmos.auth.v1beta1.Query/Account",
		Data:   requestData,
		Prove:  false,
	}

	fmt.Println(string(requestData))

	responseQuery := abcitypes.ResponseQuery{
		Height: height,
		Value:  responseData,
	}

	mockCalls := []abciQueryCall{
		{
			expectedQuery: requestQuery,
			responseQuery: responseQuery,
		},
	}

	ts := rpcTestServer(t, newABCIQueryHandler(t, mockCalls))
	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	account, err := client.Account(context.Background(), testAddr, height)
	require.NoError(t, err)
	require.Equal(t, testAccount, account)
}

func TestHTTPClient_Account_NotFound(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	codec := encodingConfig.Marshaler

	height := int64(101)
	requestData, err := codec.Marshal(&authtypes.QueryAccountRequest{Address: testAddr.String()})
	require.NoError(t, err)

	requestQuery := abciRequestQuery{
		Height: strconv.FormatInt(height, 10),
		Path:   "/cosmos.auth.v1beta1.Query/Account",
		Data:   requestData,
		Prove:  false,
	}

	responseQuery := abcitypes.ResponseQuery{
		Height: height,
		Code:   22,
		Log:    accountNotFound,
	}

	mockCalls := []abciQueryCall{
		{
			expectedQuery: requestQuery,
			responseQuery: responseQuery,
		},
	}

	ts := rpcTestServer(t, newABCIQueryHandler(t, mockCalls))
	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	account, err := client.Account(context.Background(), testAddr, height)
	require.ErrorContains(t, err, "not found")
	require.Nil(t, account)
}

func TestHTTPClient_Balance(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	codec := encodingConfig.Marshaler

	height := int64(102)
	heightStr := strconv.FormatInt(height, 10)
	queryPath := "/cosmos.bank.v1beta1.Query/AllBalances"

	expectedBalance := sdk.NewCoins(
		sdk.NewCoin("bnb", sdk.NewInt(1000000)),
		sdk.NewCoin("btcb", sdk.NewInt(2000000)),
		sdk.NewCoin("hard", sdk.NewInt(3000000)),
		sdk.NewCoin("swp", sdk.NewInt(4000000)),
		sdk.NewCoin("ukava", sdk.NewInt(5000000)),
	)

	page1Request, err := codec.Marshal(&banktypes.QueryAllBalancesRequest{
		Address:    testAddr.String(),
		Pagination: &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	})
	require.NoError(t, err)
	page1Response, err := codec.Marshal(&banktypes.QueryAllBalancesResponse{
		Balances:   expectedBalance[:2],
		Pagination: &query.PageResponse{NextKey: []byte("request-page-2")},
	})
	require.NoError(t, err)

	page2Request, err := codec.Marshal(&banktypes.QueryAllBalancesRequest{
		Address:    testAddr.String(),
		Pagination: &query.PageRequest{Key: []byte("request-page-2"), Limit: query.DefaultLimit},
	})
	page2Response, err := codec.Marshal(&banktypes.QueryAllBalancesResponse{
		Balances:   expectedBalance[2:4],
		Pagination: &query.PageResponse{NextKey: []byte("request-page-3")},
	})
	require.NoError(t, err)

	page3Request, err := codec.Marshal(&banktypes.QueryAllBalancesRequest{
		Address:    testAddr.String(),
		Pagination: &query.PageRequest{Key: []byte("request-page-3"), Limit: query.DefaultLimit},
	})
	page3Response, err := codec.Marshal(&banktypes.QueryAllBalancesResponse{
		Balances:   expectedBalance[4:],
		Pagination: &query.PageResponse{NextKey: nil},
	})
	require.NoError(t, err)

	mockCalls := []abciQueryCall{
		{abciRequestQuery{heightStr, queryPath, page1Request, false}, abcitypes.ResponseQuery{Height: height, Value: page1Response}},
		{abciRequestQuery{heightStr, queryPath, page2Request, false}, abcitypes.ResponseQuery{Height: height, Value: page2Response}},
		{abciRequestQuery{heightStr, queryPath, page3Request, false}, abcitypes.ResponseQuery{Height: height, Value: page3Response}},
	}

	ts := rpcTestServer(t, newABCIQueryHandler(t, mockCalls))
	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	balance, err := client.Balance(context.Background(), testAddr, height)
	require.NoError(t, err)
	require.Equal(t, expectedBalance, balance)
}

func TestHTTPClient_Delegated(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	codec := encodingConfig.Marshaler

	height := int64(103)
	heightStr := strconv.FormatInt(height, 10)
	queryPath := "/cosmos.staking.v1beta1.Query/DelegatorDelegations"

	expectedDelegationResponses := stakingtypes.DelegationResponses{
		{
			Delegation: stakingtypes.Delegation{
				DelegatorAddress: testAddr.String(),
				ValidatorAddress: "kavavaloper1ppj7c8tqt2e3rzqtmztsmd6ea6u3nz6qggcp5e",
				Shares:           sdk.MustNewDecFromStr("0.000073454065009902"),
			},
			Balance: sdk.NewCoin("ukava", sdk.NewInt(0)),
		},
		{
			Delegation: stakingtypes.Delegation{
				DelegatorAddress: testAddr.String(),
				ValidatorAddress: "kavavaloper1zw8ce44kdqzfu0r2t9qwr75gqdcarclf9fj9lt",
				Shares:           sdk.MustNewDecFromStr("19399980.000000000000000000"),
			},
			Balance: sdk.NewCoin("ukava", sdk.NewInt(19399980)),
		},
		{
			Delegation: stakingtypes.Delegation{
				DelegatorAddress: testAddr.String(),
				ValidatorAddress: "kavavaloper1ffcujj05v6220ccxa6qdnpz3j48ng024ykh2df",
				Shares:           sdk.MustNewDecFromStr("13301323.130293333267920034"),
			},
			Balance: sdk.NewCoin("ukava", sdk.NewInt(13299993)),
		},
		{
			Delegation: stakingtypes.Delegation{
				DelegatorAddress: testAddr.String(),
				ValidatorAddress: "kavavaloper10m3hjapny44txmgr47rf277364htgqpr646cty",
				Shares:           sdk.MustNewDecFromStr("0.908355880844728396"),
			},
			Balance: sdk.NewCoin("ukava", sdk.NewInt(0)),
		},
		{
			Delegation: stakingtypes.Delegation{
				DelegatorAddress: testAddr.String(),
				ValidatorAddress: "kavavaloper1ceun2qqw65qce5la33j8zv8ltyyaqqfctl35n4",
				Shares:           sdk.MustNewDecFromStr("9600953.092623759678779246"),
			},
			Balance: sdk.NewCoin("ukava", sdk.NewInt(9599993)),
		},
	}

	page1Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	})
	require.NoError(t, err)
	page1Response, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsResponse{
		DelegationResponses: expectedDelegationResponses[:2],
		Pagination:          &query.PageResponse{NextKey: []byte("request-page-2")},
	})
	require.NoError(t, err)

	page2Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: []byte("request-page-2"), Limit: query.DefaultLimit},
	})
	page2Response, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsResponse{
		DelegationResponses: expectedDelegationResponses[2:4],
		Pagination:          &query.PageResponse{NextKey: []byte("request-page-3")},
	})
	require.NoError(t, err)

	page3Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: []byte("request-page-3"), Limit: query.DefaultLimit},
	})
	page3Response, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsResponse{
		DelegationResponses: expectedDelegationResponses[4:],
		Pagination:          &query.PageResponse{NextKey: nil},
	})
	require.NoError(t, err)

	mockCalls := []abciQueryCall{
		{abciRequestQuery{heightStr, queryPath, page1Request, false}, abcitypes.ResponseQuery{Height: height, Value: page1Response}},
		{abciRequestQuery{heightStr, queryPath, page2Request, false}, abcitypes.ResponseQuery{Height: height, Value: page2Response}},
		{abciRequestQuery{heightStr, queryPath, page3Request, false}, abcitypes.ResponseQuery{Height: height, Value: page3Response}},
	}

	ts := rpcTestServer(t, newABCIQueryHandler(t, mockCalls))
	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	delegationResponses, err := client.Delegations(context.Background(), testAddr, height)
	require.NoError(t, err)
	require.Equal(t, expectedDelegationResponses, delegationResponses)
}

func TestHTTPClient_UnbondingDelegations(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	codec := encodingConfig.Marshaler

	completionTime := time.Now()

	height := int64(104)
	heightStr := strconv.FormatInt(height, 10)
	queryPath := "/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations"

	expectedUnbondingDelegations := stakingtypes.UnbondingDelegations{
		{
			DelegatorAddress: testAddr.String(),
			ValidatorAddress: "kavavaloper1ppj7c8tqt2e3rzqtmztsmd6ea6u3nz6qggcp5e",
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight:          50,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(1000000),
					Balance:                 sdk.NewInt(1000000),
					UnbondingId:             1,
					UnbondingOnHoldRefCount: 1,
				},
				{
					CreationHeight:          51,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(2000000),
					Balance:                 sdk.NewInt(2000000),
					UnbondingId:             2,
					UnbondingOnHoldRefCount: 2,
				},
			},
		},
		{
			DelegatorAddress: testAddr.String(),
			ValidatorAddress: "kavavaloper1zw8ce44kdqzfu0r2t9qwr75gqdcarclf9fj9lt",
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight:          52,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(1000000),
					Balance:                 sdk.NewInt(2000000),
					UnbondingId:             3,
					UnbondingOnHoldRefCount: 0,
				},
			},
		},
		{
			DelegatorAddress: testAddr.String(),
			ValidatorAddress: "kavavaloper1ffcujj05v6220ccxa6qdnpz3j48ng024ykh2df",
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight:          54,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(2000000),
					Balance:                 sdk.NewInt(3000000),
					UnbondingId:             4,
					UnbondingOnHoldRefCount: 0,
				},
				{
					CreationHeight:          55,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(1000000),
					Balance:                 sdk.NewInt(1000000),
					UnbondingId:             5,
					UnbondingOnHoldRefCount: 1,
				},
				{
					CreationHeight:          56,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(2000000),
					Balance:                 sdk.NewInt(2000000),
					UnbondingId:             6,
					UnbondingOnHoldRefCount: 2,
				},
			},
		},
		{
			DelegatorAddress: testAddr.String(),
			ValidatorAddress: "kavavaloper10m3hjapny44txmgr47rf277364htgqpr646cty",
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight:          57,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(1000000),
					Balance:                 sdk.NewInt(1000000),
					UnbondingId:             7,
					UnbondingOnHoldRefCount: 1,
				},
				{
					CreationHeight:          58,
					CompletionTime:          completionTime,
					InitialBalance:          sdk.NewInt(2000000),
					Balance:                 sdk.NewInt(2000000),
					UnbondingId:             8,
					UnbondingOnHoldRefCount: 2,
				},
			},
		},
		{
			DelegatorAddress: testAddr.String(),
			ValidatorAddress: "kavavaloper1ceun2qqw65qce5la33j8zv8ltyyaqqfctl35n4",
			Entries: []stakingtypes.UnbondingDelegationEntry{
				{
					CreationHeight:          59,
					CompletionTime:          time.Now().Add(24 * time.Hour),
					InitialBalance:          sdk.NewInt(1000000),
					Balance:                 sdk.NewInt(1000000),
					UnbondingId:             9,
					UnbondingOnHoldRefCount: 10,
				},
			},
		},
	}

	page1Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: nil, Limit: query.DefaultLimit},
	})
	require.NoError(t, err)
	page1Response, err := codec.Marshal(&stakingtypes.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingResponses: expectedUnbondingDelegations[:2],
		Pagination:         &query.PageResponse{NextKey: []byte("request-page-2")},
	})
	require.NoError(t, err)

	page2Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: []byte("request-page-2"), Limit: query.DefaultLimit},
	})
	page2Response, err := codec.Marshal(&stakingtypes.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingResponses: expectedUnbondingDelegations[2:4],
		Pagination:         &query.PageResponse{NextKey: []byte("request-page-3")},
	})
	require.NoError(t, err)

	page3Request, err := codec.Marshal(&stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: testAddr.String(),
		Pagination:    &query.PageRequest{Key: []byte("request-page-3"), Limit: query.DefaultLimit},
	})
	page3Response, err := codec.Marshal(&stakingtypes.QueryDelegatorUnbondingDelegationsResponse{
		UnbondingResponses: expectedUnbondingDelegations[4:],
		Pagination:         &query.PageResponse{NextKey: nil},
	})
	require.NoError(t, err)

	mockCalls := []abciQueryCall{
		{abciRequestQuery{heightStr, queryPath, page1Request, false}, abcitypes.ResponseQuery{Height: height, Value: page1Response}},
		{abciRequestQuery{heightStr, queryPath, page2Request, false}, abcitypes.ResponseQuery{Height: height, Value: page2Response}},
		{abciRequestQuery{heightStr, queryPath, page3Request, false}, abcitypes.ResponseQuery{Height: height, Value: page3Response}},
	}

	ts := rpcTestServer(t, newABCIQueryHandler(t, mockCalls))
	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	unbondingDelegations, err := client.UnbondingDelegations(context.Background(), testAddr, height)
	require.NoError(t, err)

	// use go-cmp due to marshal & unmarshal of timestamps
	require.True(t, cmp.Equal(expectedUnbondingDelegations, unbondingDelegations))
}

func TestHTTPClient_SimulateTx(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	testTx := encodingConfig.TxConfig.NewTxBuilder().GetTx()

	mockResponse := sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{
			GasWanted: 500000,
			GasUsed:   200000,
		},
	}

	var simulateResponse func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse
	ts := rpcTestServer(t, func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		assert.Equal(t, "abci_query", request.Method)
		return simulateResponse(request)
	})
	defer ts.Close()

	client, err := kava.NewHTTPClient(ts.URL)
	require.NoError(t, err)

	simulateResponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		var params struct {
			Path   string
			Data   bytes.HexBytes
			Height string
			Prove  bool
		}

		err := json.Unmarshal(request.Params, &params)
		require.NoError(t, err)

		assert.Equal(t, "0", params.Height)

		_, err = encodingConfig.TxConfig.TxDecoder()(params.Data)
		require.NoError(t, err)

		respValue, err := encodingConfig.Marshaler.MarshalJSON(&mockResponse)
		require.NoError(t, err)

		abciResult := ctypes.ResultABCIQuery{
			Response: abcitypes.ResponseQuery{
				Value: respValue,
			},
		}

		data, err := json.Marshal(&abciResult)
		require.NoError(t, err)

		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Result:  json.RawMessage(data),
		}
	}
	simResp, err := client.SimulateTx(context.Background(), testTx)
	assert.NoError(t, err)
	assert.Equal(t, mockResponse, *simResp)

	simulateResponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Error: &jsonrpctypes.RPCError{
				Code:    1,
				Message: "something went wrong",
			},
		}
	}
	simResp, err = client.SimulateTx(context.Background(), testTx)
	assert.Nil(t, simResp)
	assert.ErrorContains(t, err, "something went wrong")

	simulateResponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		abciResult := ctypes.ResultABCIQuery{
			Response: abcitypes.ResponseQuery{
				Value: []byte("invalid"),
			},
		}

		data, err := json.Marshal(&abciResult)
		require.NoError(t, err)

		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Result:  data,
		}
	}
	simResp, err = client.SimulateTx(context.Background(), testTx)
	assert.Nil(t, simResp)
	assert.Error(t, err)
}

func TestParseABCIResult(t *testing.T) {
	mockOKResponse := &ctypes.ResultABCIQuery{
		Response: abcitypes.ResponseQuery{
			Code:  uint32(0),
			Log:   "",
			Value: []byte("{}"),
		},
	}

	mockNotOKResponse := &ctypes.ResultABCIQuery{
		Response: abcitypes.ResponseQuery{
			Code:  uint32(1),
			Log:   "internal error",
			Value: []byte("{}"),
		},
	}

	mockNilByteResponse := &ctypes.ResultABCIQuery{
		Response: abcitypes.ResponseQuery{
			Code:  uint32(0),
			Log:   "",
			Value: []byte(nil),
		},
	}

	mockABCIError := errors.New("abci error")

	// if abci errors, we return error and empty bytes
	data, err := kava.ParseABCIResult(mockOKResponse, mockABCIError)
	assert.Equal(t, []byte{}, data)
	assert.Equal(t, mockABCIError, err)

	// if response is not OK, we return log error with empty bytes
	data, err = kava.ParseABCIResult(mockNotOKResponse, nil)
	assert.Equal(t, []byte{}, data)
	assert.Equal(t, errors.New(mockNotOKResponse.Response.Log), err)

	// if response is OK , we return nil error with Response value
	data, err = kava.ParseABCIResult(mockOKResponse, nil)
	assert.Equal(t, mockOKResponse.Response.Value, data)
	assert.Nil(t, err)

	// if response is len 0, we return
	data, err = kava.ParseABCIResult(mockNilByteResponse, nil)
	assert.Equal(t, []byte{}, data)
	assert.Nil(t, err)
}
