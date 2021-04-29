// Copyright 2021 Kava Labs, Inc.
// Copyright 2016 All in Bits, Inc
//
// Derived from github.com/cosmos/cosmos-sdk/server/rosetta@5ea817a6fb,
//    github.com/tendermint/cosmos-rosetta-gateway@fc0ec20 (v3.0.0-rc2),
//    github.com/tendermint/cosmos-rosetta-gateway@b81c4e9 (v0.1.1)
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
)

const (
	BeginBlockHashStart = 0x0
	EndBlockHashStart   = 0x1
)

// BeginBlockTxHash caluclates the begin blocker transaction hash
func BeginBlockTxHash(blockHash []byte) string {
	prefixedHash := append([]byte{BeginBlockHashStart}, blockHash...)
	return strings.ToUpper(hex.EncodeToString(prefixedHash))
}

// EndBlockTxHash caluclates the end blocker transaction hash
func EndBlockTxHash(blockHash []byte) string {
	prefixedHash := append([]byte{EndBlockHashStart}, blockHash...)
	return strings.ToUpper(hex.EncodeToString(prefixedHash))
}
