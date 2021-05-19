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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func validConstructionHashRequest(txBytes []byte, blockchain, network string) *types.ConstructionHashRequest {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: blockchain,
		Network:    network,
	}

	return &types.ConstructionHashRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: hex.EncodeToString(txBytes),
	}
}

func TestConstructionHash(t *testing.T) {
	servicer := setupContructionAPIServicer()

	testCases := []struct {
		testFixtureFile string
		expectErr       bool
		expectedErrCode int32
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
			testFixtureFile: "long-memo.json", // memo length = maxABCIDataLength
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

		// Calculated expected tx hash
		hash := sha256.Sum256(bz)
		bzHash := hash[:]
		expectedTxHash := strings.ToUpper(hex.EncodeToString(bzHash))

		// Check that response contains expected tx hash
		ctx := context.Background()
		request := validConstructionHashRequest(bz, "Kava", "testing")
		response, rosettaErr := servicer.ConstructionHash(ctx, request)
		if tc.expectErr {
			require.Equal(t, tc.expectedErrCode, rosettaErr.Code)
		} else {
			require.Nil(t, rosettaErr)
			require.Equal(t, expectedTxHash, response.TransactionIdentifier.Hash)
		}
	}
}
