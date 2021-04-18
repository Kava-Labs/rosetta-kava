package kava

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	amino "github.com/tendermint/go-amino"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	jsonrpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestHTTPClient(t *testing.T) {
	cdc := amino.NewCodec()
	ctypes.RegisterAmino(cdc)

	ts := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		var request jsonrpctypes.RPCRequest
		err = json.Unmarshal(body, &request)
		require.NoError(t, err)

		var params struct {
			Hash string
		}
		err = json.Unmarshal(request.Params, &params)
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

		b, err := json.Marshal(&response)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	}))
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
