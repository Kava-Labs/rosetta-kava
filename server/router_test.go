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

package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kava-labs/rosetta-kava/clientid"
	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	networkIdentifier = &types.NetworkIdentifier{
		Blockchain: kava.Blockchain,
		Network:    "kava-testnet",
	}
)

func TestRouter(t *testing.T) {
	config := &configuration.Configuration{
		Mode:              configuration.Offline,
		NetworkIdentifier: networkIdentifier,
		Port:              8000,
		KavaRPCURL:        "https://rpc.testnet.kava.io:443",
	}

	_, err := NewRouter(config)
	assert.NoError(t, err)
}

// TestRouter_ClientID sends a Rosetta request through the full router and
// checks that the node receives the request's client ID header.
func TestRouter_ClientID(t *testing.T) {
	received := make(chan http.Header, 16)
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer node.Close()

	config := &configuration.Configuration{
		Mode:              configuration.Online,
		NetworkIdentifier: networkIdentifier,
		Port:              8000,
		KavaRPCURL:        node.URL,
		ClientID:          clientid.Config{Header: "x-client-id", Default: "example-rosetta"},
	}

	handler, err := NewRouter(config)
	require.NoError(t, err)
	rosetta := httptest.NewServer(handler)
	defer rosetta.Close()

	body, err := json.Marshal(types.NetworkRequest{NetworkIdentifier: networkIdentifier})
	require.NoError(t, err)

	send := func(clientID string) http.Header {
		req, err := http.NewRequest(http.MethodPost, rosetta.URL+"/network/status", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if clientID != "" {
			req.Header.Set("X-Client-ID", clientID)
		}
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		_ = res.Body.Close()

		select {
		case h := <-received:
			// Drain any further node calls from this request.
			for len(received) > 0 {
				<-received
			}
			return h
		default:
			t.Fatal("node received no call")
			return nil
		}
	}

	assert.Equal(t, "caller-a", send("caller-a").Get("x-client-id"))

	_, present := send("")["X-Client-Id"]
	assert.False(t, present, "a request without the header sends none")
}
