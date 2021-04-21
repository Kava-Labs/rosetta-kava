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
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	amino "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/bytes"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func rpcTestServer(
	t *testing.T,
	rpcHandler func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse,
) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var request jsonrpctypes.RPCRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err)

		response := rpcHandler(request)

		b, err := json.Marshal(&response)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	}))
}

func TestHTTPClient_BlockByHash(t *testing.T) {
	cdc := amino.NewCodec()
	ctypes.RegisterAmino(cdc)

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

	client, err := NewHTTPClient(ts.URL)
	require.NoError(t, err)

	testHash := []byte("D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75")
	block, err := client.BlockByHash(testHash)
	assert.NoError(t, err)
	assert.Equal(t, testHash, []byte(block.BlockID.Hash))

	testHash = []byte{}
	block, err = client.BlockByHash(testHash)
	assert.Error(t, err)
	assert.Nil(t, block)
	assert.EqualError(t, err, "BlockByHash: RPC error 1 - invalid hash")
}

func TestHTTPClient_Account(t *testing.T) {
	cdc := amino.NewCodec()
	ctypes.RegisterAmino(cdc)
	height := int64(100)

	testAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	mockAccountPath := filepath.Join("test-fixtures", "vesting-account.json")
	mockAccount, err := ioutil.ReadFile(mockAccountPath)
	require.NoError(t, err)

	var accountRPCReponse func(jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse

	ts := rpcTestServer(t, func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		assert.Equal(t, "abci_query", request.Method)

		var params struct {
			Path   string
			Data   bytes.HexBytes
			Height string
			Prove  bool
		}

		err := json.Unmarshal(request.Params, &params)
		require.NoError(t, err)

		assert.Equal(t, strconv.FormatInt(height, 10), params.Height)

		return accountRPCReponse(request)
	})
	defer ts.Close()

	client, err := NewHTTPClient(ts.URL)
	require.NoError(t, err)

	addr, err := sdk.AccAddressFromBech32(testAddr)
	require.NoError(t, err)

	accountRPCReponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		abciResult := ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				Value: mockAccount,
			},
		}

		data, err := cdc.MarshalJSON(&abciResult)
		require.NoError(t, err)

		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Result:  json.RawMessage(data),
		}
	}
	acc, err := client.Account(addr, height)
	assert.NoError(t, err)
	assert.NotNil(t, acc)

	accountRPCReponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Error: &jsonrpctypes.RPCError{
				Code:    1,
				Message: "invalid account",
			},
		}
	}
	acc, err = client.Account(addr, height)
	assert.Nil(t, acc)
	assert.EqualError(t, err, "ABCIQuery: RPC error 1 - invalid account")

	accountRPCReponse = func(request jsonrpctypes.RPCRequest) jsonrpctypes.RPCResponse {
		return jsonrpctypes.RPCResponse{
			JSONRPC: request.JSONRPC,
			ID:      request.ID,
			Result:  json.RawMessage("{}"),
		}
	}
	acc, err = client.Account(addr, height)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "UnmarshalJSON")
}

func TestParseABCIResult(t *testing.T) {
	mockOKResponse := &ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Code:  uint32(0),
			Log:   "",
			Value: []byte("{}"),
		},
	}

	mockNotOKResponse := &ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Code:  uint32(1),
			Log:   "internal error",
			Value: []byte("{}"),
		},
	}

	mockNilByteResponse := &ctypes.ResultABCIQuery{
		Response: abci.ResponseQuery{
			Code:  uint32(0),
			Log:   "",
			Value: []byte(nil),
		},
	}

	mockABCIError := errors.New("abci error")

	// if abci errors, we return error and empty bytes
	data, err := parseABCIResult(mockOKResponse, mockABCIError)
	assert.Equal(t, []byte{}, data)
	assert.Equal(t, mockABCIError, err)

	// if response is not OK, we return log error with empty bytes
	data, err = parseABCIResult(mockNotOKResponse, nil)
	assert.Equal(t, []byte{}, data)
	assert.Equal(t, errors.New(mockNotOKResponse.Response.Log), err)

	// if response is OK , we return nil error with Reponse value
	data, err = parseABCIResult(mockOKResponse, nil)
	assert.Equal(t, mockOKResponse.Response.Value, data)
	assert.Nil(t, err)

	// if response is len 0, we return
	data, err = parseABCIResult(mockNilByteResponse, nil)
	assert.Equal(t, []byte{}, data)
	assert.Nil(t, err)
}
