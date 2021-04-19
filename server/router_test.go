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
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

var (
	networkIdentifier = &types.NetworkIdentifier{
		Blockchain: kava.Blockchain,
		Network:    "kava-testnet-1",
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
