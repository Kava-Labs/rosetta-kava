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
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"
	"github.com/kava-labs/rosetta-kava/services"

	"github.com/coinbase/rosetta-sdk-go/asserter"
	sdkserver "github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	// readTimeout is the maximum duration for reading the entire
	// request, including the body.
	readTimeout = 5 * time.Second

	// writeTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read.
	writeTimeout = 120 * time.Second

	// idleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled.
	idleTimeout = 30 * time.Second
)

// NewRouter returns an rossetta server handler with assertion, logging and cors support
func NewRouter(config *configuration.Configuration) (http.Handler, error) {
	http, err := kava.NewHTTPClient(config.KavaRPCURL)
	if err != nil {
		return nil, fmt.Errorf("%w: could not initialize http client", err)
	}

	accountBalanceFactory := kava.NewRPCBalanceFactory(http)

	client, err := kava.NewClient(http, accountBalanceFactory)
	if err != nil {
		return nil, fmt.Errorf("%w: could not initialize kava client", err)
	}

	// The asserter automatically rejects incorrectly formatted requests.
	asserter, err := asserter.NewServer(
		kava.OperationTypes,
		kava.HistoricalBalanceSupported,
		[]*types.NetworkIdentifier{config.NetworkIdentifier},
		kava.CallMethods,
		kava.IncludeMempoolCoins,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: could not initialize server asserter", err)
	}

	router := services.NewBlockchainRouter(config, client, asserter)

	loggedRouter := sdkserver.LoggerMiddleware(router)
	corsRouter := sdkserver.CorsMiddleware(loggedRouter)

	return corsRouter, nil
}

// Run starts a http server using the provided handler with read, write, and idle timeouts
func Run(config *configuration.Configuration, handler http.Handler) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Printf("server listening on port %d", config.Port)

	return server.ListenAndServe()
}
