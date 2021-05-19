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
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coinbase/rosetta-sdk-go/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"
)

func TestConstructionSubmit(t *testing.T) {
	// Set up servicer with mock client
	cfg := &configuration.Configuration{
		Mode: configuration.Online,
	}
	mockClient := &mocks.Client{}
	cdc := app.MakeCodec()
	servicer := NewConstructionAPIService(cfg, mockClient, cdc)

	testCases := []struct {
		testFixtureFile string
		expectErr       bool
		expectedErrCode int32
	}{
		{
			testFixtureFile: "msg-send.json",
			expectErr:       false,
		},
	}

	for _, tc := range testCases {
		// Load signed transaction from file
		relPath, err := filepath.Rel(
			"services",
			fmt.Sprintf("kava/test-fixtures/signed-msgs/%s", tc.testFixtureFile),
		)
		require.NoError(t, err)
		bz, err := ioutil.ReadFile(relPath)
		require.NoError(t, err)

		cdc := app.MakeCodec()
		var stdtx authtypes.StdTx
		err = cdc.UnmarshalJSON(bz, &stdtx)
		require.NoError(t, err)

		payload, err := cdc.MarshalBinaryLengthPrefixed(stdtx)
		require.NoError(t, err)

		// Expected response
		txIndentifier := &types.TransactionIdentifier{
			Hash: hex.EncodeToString(tmtypes.Tx(bz).Hash()),
		}
		metadata := make(map[string]interface{})

		mockClient.On(
			"PostTx",
			payload,
		).Return(
			txIndentifier,
			metadata,
			nil,
		).Once()

		res, meta, err := servicer.client.PostTx(payload)
		require.Nil(t, err)
		assert.Equal(t, res, txIndentifier)
		assert.Equal(t, meta, metadata)
	}
}
