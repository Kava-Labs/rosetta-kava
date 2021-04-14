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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	rclient "github.com/coinbase/rosetta-sdk-go/client"
	"github.com/kava-labs/rosetta-kava/configuration"
	router "github.com/kava-labs/rosetta-kava/server"
)

var config *configuration.Configuration
var server *httptest.Server
var client *rclient.APIClient

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

	client = rclient.NewAPIClient(clientConfig)

	os.Exit(m.Run())
}
