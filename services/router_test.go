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

package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := mocks.Client{}
	server, err := asserter.NewServer(
		[]string{"Transfer"},
		true,
		[]*types.NetworkIdentifier{
			{
				Blockchain: "Kava",
				Network:    "kava-6",
			},
		},
		[]string{},
		false,
		"",
	)
	assert.NoError(t, err)

	handler := NewBlockchainRouter(cfg, &mockClient, server)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	onlineOnlyEndpoints := []string{
		"/account/balance",
		"/account/coins",
		"/block",
		"/block/transaction",
		"/mempool",
		"/mempool/transaction",
		"/network/status",
		"/call",
		"/construction/metadata",
		"/construction/submit",
	}

	for _, endpoint := range onlineOnlyEndpoints {
		t.Run(endpoint[1:], func(t *testing.T) {
			request := bytes.NewBuffer([]byte(`{"network_identifier":{"blockchain": "Kava","network": "kava-6"}}`))
			res, err := http.Post(ts.URL+"/network/status", "application/json", request)
			assert.NoError(t, err)

			data, err := ioutil.ReadAll(res.Body)
			assert.NoError(t, err)

			var errResp types.Error
			err = json.Unmarshal(data, &errResp)
			assert.NoError(t, err)

			assert.Equal(t, ErrUnavailableOffline.Code, errResp.Code)
			assert.Equal(t, ErrUnavailableOffline.Message, errResp.Message)
			assert.Equal(t, ErrUnavailableOffline.Retriable, false)
		})
	}
}
