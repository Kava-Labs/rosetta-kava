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
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestCosmosSDKConfig(t *testing.T) {
	config := sdk.GetConfig()

	coinType := config.GetCoinType()
	assert.Equal(t, uint32(459), coinType)

	prefix := config.GetBech32AccountAddrPrefix()
	assert.Equal(t, "kava", prefix)

	prefix = config.GetBech32ValidatorAddrPrefix()
	assert.Equal(t, "kavavaloper", prefix)

	prefix = config.GetBech32ConsensusAddrPrefix()
	assert.Equal(t, "kavavalcons", prefix)

	prefix = config.GetBech32AccountPubPrefix()
	assert.Equal(t, "kavapub", prefix)

	prefix = config.GetBech32ConsensusPubPrefix()
	assert.Equal(t, "kavavalconspub", prefix)

	assert.PanicsWithValue(t, "Config is sealed", func() { config.SetCoinType(459) })
}

func TestClient(t *testing.T) {
	_, err := NewClient()
	assert.NoError(t, err)
}
