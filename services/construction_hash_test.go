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
	"testing"

	"github.com/kava-labs/go-sdk/keys"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/cdp"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
	TestMnemonic = "conduct elegant layer interest unknown warrior deliver fringe fall door link start then potato stand concert bacon east glare boat scene ring idea cruel"
	TestAddress  = "kava10an7smc50d6xlgxul3sh8uhz2887t7ruk93a5a"
)

func validConstructionHashRequest() (*types.ConstructionHashRequest, error) {

	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "Kava",
		Network:    "kava-7",
	}

	cdc := app.MakeCodec()

	testAddr, err := sdk.AccAddressFromBech32(TestAddress)
	if err != nil {
		return nil, err
	}

	keyManager, err := keys.NewMnemonicKeyManager(TestMnemonic, app.Bip44CoinType)
	if err != nil {
		return nil, err
	}

	collateral := sdk.NewCoin("ukava", sdk.NewInt(2))
	principal := sdk.NewCoin("usdx", sdk.NewInt(100))
	msg := cdp.NewMsgCreateCDP(testAddr, collateral, principal, "ukava-a")
	unsignedMsgs := []sdk.Msg{msg}

	signMsg := authtypes.StdSignMsg{
		ChainID:       networkIdentifier.Network,
		AccountNumber: 0,
		Sequence:      0,
		Fee:           authtypes.NewStdFee(250000, nil),
		Msgs:          unsignedMsgs,
		Memo:          "",
	}

	for _, m := range signMsg.Msgs {
		if err := m.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	bz, err := keyManager.Sign(signMsg, cdc)
	if err != nil {
		return nil, err
	}
	tx := tmtypes.Tx(bz)

	return &types.ConstructionHashRequest{
		NetworkIdentifier: networkIdentifier,
		SignedTransaction: tx.String(),
	}, nil
}

func TestConstructionHash(t *testing.T) {
	servicer, _ := setupContructionAPIServicer()

	hashReq, err := validConstructionHashRequest()
	require.Nil(t, err)

	ctx := context.Background()
	response, rosettaErr := servicer.ConstructionHash(ctx, hashReq)
	require.Nil(t, rosettaErr)
	assert.NotNil(t, response)
}
