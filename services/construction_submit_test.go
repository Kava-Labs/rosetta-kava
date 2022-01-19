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
	"context"
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
	cdc := app.MakeEncodingConfig().Amino
	servicer := NewConstructionAPIService(cfg, mockClient, cdc)

	testCases := []struct {
		testFixtureFile string
		expectErr       bool
	}{
		{
			testFixtureFile: "msg-send.json",
			expectErr:       false,
		},
		{
			testFixtureFile: "msg-create-cdp.json",
			expectErr:       false,
		},
		{
			testFixtureFile: "msg-hard-deposit.json",
			expectErr:       false,
		},
		{
			testFixtureFile: "multiple-msgs.json",
			expectErr:       false,
		},
		{
			testFixtureFile: "unsigned-msg-send.json",
			expectErr:       true,
		},
	}

	for _, tc := range testCases {
		// Load signed transaction from file
		relPath, err := filepath.Rel(
			"services",
			fmt.Sprintf("kava/test-fixtures/txs/%s", tc.testFixtureFile),
		)
		require.NoError(t, err)
		bz, err := ioutil.ReadFile(relPath)
		require.NoError(t, err)

		cdc := app.MakeEncodingConfig().Amino
		var stdtx authtypes.StdTx
		err = cdc.UnmarshalJSON(bz, &stdtx)
		require.NoError(t, err)

		payload, err := cdc.MarshalBinaryLengthPrefixed(stdtx)
		require.NoError(t, err)

		// Expected response
		txIndentifier := &types.TransactionIdentifier{
			Hash: hex.EncodeToString(tmtypes.Tx(bz).Hash()),
		}
		err = stdtx.ValidateBasic()
		mockErr := err

		mockClient.On(
			"PostTx",
			payload,
		).Return(
			txIndentifier,
			mockErr,
		).Once()

		ctx := context.Background()
		request := &types.ConstructionSubmitRequest{
			SignedTransaction: hex.EncodeToString(payload),
		}
		res, rerr := servicer.ConstructionSubmit(ctx, request)
		if tc.expectErr {
			assert.Equal(t, wrapErr(ErrKava, mockErr), rerr)
		} else {
			require.Nil(t, err)
			assert.Equal(t, &types.TransactionIdentifierResponse{TransactionIdentifier: txIndentifier}, res)
		}
	}
}
