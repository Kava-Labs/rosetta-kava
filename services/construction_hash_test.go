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

	"github.com/stretchr/testify/require"

	"github.com/coinbase/rosetta-sdk-go/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kava-labs/kava/app"
)

func TestConstructionHash(t *testing.T) {
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		testFixtureFile string
		expectedTxHash  string
		expectErr       bool
		expectedErrCode int32
	}{
		{
			testFixtureFile: "msg-send.json",
			expectedTxHash:  "4E218DC828F45B7112F7CF6B328563045B5307B07D8602549389553F3B27D997",
			expectErr:       false,
		},
		{
			testFixtureFile: "msg-create-cdp.json",
			expectErr:       false,
			expectedTxHash:  "02C44611CD6898E89839F34830A089AD67A1FDA59D809EABA24B5A4B236849BB",
		},
		{
			testFixtureFile: "msg-hard-deposit.json",
			expectErr:       false,
			expectedTxHash:  "E47E8BB9FA3C90B925D46C75DA03BB316ABB9D04CB647854AC215CB7C743368C",
		},
		{
			testFixtureFile: "multiple-msgs.json",
			expectErr:       false,
			expectedTxHash:  "4F5EB96A9F29554F2BF0E01059268B1919D5702C29440B017E5C656547725F4C",
		},
		{
			testFixtureFile: "long-memo.json",
			expectErr:       false,
			expectedTxHash:  "C25EBDC1FB86BEE1F21FB1F0A97925A64ECF838B424D4E57758751806A100FBF",
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
		signedTx := hex.EncodeToString(payload)

		networkIdentifier := &types.NetworkIdentifier{
			Blockchain: "Kava",
			Network:    "testing",
		}

		request := &types.ConstructionHashRequest{
			NetworkIdentifier: networkIdentifier,
			SignedTransaction: signedTx,
		}

		// Check that response contains expected tx hash
		ctx := context.Background()
		response, rosettaErr := servicer.ConstructionHash(ctx, request)
		if tc.expectErr {
			require.Equal(t, tc.expectedErrCode, rosettaErr.Code)
		} else {
			require.Nil(t, rosettaErr)
			require.Equal(t, tc.expectedTxHash, response.TransactionIdentifier.Hash)
		}
	}
}
