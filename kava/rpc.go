package kava

import (
	"github.com/pkg/errors"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// HTTPClient extends the tendermint http client to enable finding blocks by hash
type HTTPClient struct {
	*tmhttp.HTTP
	caller *tmclient.Client
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
	cdc := rpc.Codec()
	ctypes.RegisterAmino(cdc)
	rpc.SetCodec(cdc)

	return &HTTPClient{
		HTTP:   http,
		caller: rpc,
	}, nil
}

// BlockByHash fetches a block by it's hash value and return the resulting block
func (c *HTTPClient) BlockByHash(hash []byte) (*ctypes.ResultBlock, error) {
	result := new(ctypes.ResultBlock)
	_, err := c.caller.Call("block_by_hash", map[string]interface{}{"hash": hash}, result)
	if err != nil {
		return nil, errors.Wrap(err, "BlockByHash")
	}
	return result, nil
}
