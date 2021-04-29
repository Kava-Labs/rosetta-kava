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

package kava

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockTxHash(t *testing.T) {
	mockBlockHash := "D92BDF0B5EDB04434B398A59B2FD4ED3D52B4820A18DAC7311EBDF5D37467E75"
	mockBlockHashBytes, err := hex.DecodeString(mockBlockHash)
	require.NoError(t, err)

	beginBlockHash := BeginBlockTxHash(mockBlockHashBytes)
	endBlockHash := EndBlockTxHash(mockBlockHashBytes)

	// not equal to block hash
	assert.NotEqual(t, mockBlockHash, beginBlockHash)
	assert.NotEqual(t, mockBlockHash, endBlockHash)

	// not equal to each other
	assert.NotEqual(t, beginBlockHash, endBlockHash)

	// all uppercase
	assert.Equal(t, strings.ToUpper(beginBlockHash), beginBlockHash)
	assert.Equal(t, strings.ToUpper(endBlockHash), endBlockHash)

	// get byte representation
	beginBlockHashBytes, err := hex.DecodeString(beginBlockHash)
	require.NoError(t, err)
	endBlockHashBytes, err := hex.DecodeString(endBlockHash)
	require.NoError(t, err)

	// one byte longer than block hash
	assert.Equal(t, len(mockBlockHashBytes)+1, len(beginBlockHashBytes))
	assert.Equal(t, len(mockBlockHashBytes)+1, len(endBlockHashBytes))

	// first byte of begin block hash is 0x0
	assert.Equal(t, uint8(0), beginBlockHashBytes[0])
	// first byte of end block hash is 0x1
	assert.Equal(t, uint8(1), endBlockHashBytes[0])
}
