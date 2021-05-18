package kava_test

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
