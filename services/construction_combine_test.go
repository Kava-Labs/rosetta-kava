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
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestConstructionCombine(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "Kava",
		Network:    "kava-test-9000",
	}

	signerAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	signerAccAddr, err := sdk.AccAddressFromBech32(signerAddr)
	require.NoError(t, err)
	signerPubKey, err := base64.StdEncoding.DecodeString("AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY")
	require.NoError(t, err)
	toAccAddress, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	tx := auth.NewStdTx(
		[]sdk.Msg{bank.MsgSend{FromAddress: signerAccAddr, ToAddress: toAccAddress, Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))}},
		auth.NewStdFee(250001, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(62501)))),
		[]auth.StdSignature{},
		"some memo",
	)

	cdc := app.MakeEncodingConfig().Amino
	txBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
	require.NoError(t, err)
	txPayload := hex.EncodeToString(txBytes)

	signBytes := auth.StdSignBytes(
		networkIdentifier.Network, 10, 11, tx.Fee, tx.Msgs, tx.Memo,
	)

	mockSignBytes := []byte("some signature")

	request := &types.ConstructionCombineRequest{
		NetworkIdentifier:   networkIdentifier,
		UnsignedTransaction: txPayload,
		Signatures: []*types.Signature{
			{
				SigningPayload: &types.SigningPayload{
					AccountIdentifier: &types.AccountIdentifier{Address: signerAccAddr.String()},
					Bytes:             crypto.Sha256(signBytes),
					SignatureType:     types.Ecdsa,
				},
				PublicKey: &types.PublicKey{
					Bytes:     signerPubKey,
					CurveType: types.Secp256k1,
				},
				SignatureType: types.Ecdsa,
				Bytes:         mockSignBytes,
			},
		},
	}

	servicer, _ := setupConstructionAPIServicer()
	ctx := context.Background()
	response, rerr := servicer.ConstructionCombine(ctx, request)
	require.Nil(t, rerr)

	signedTxBytes, err := hex.DecodeString(response.SignedTransaction)
	require.NoError(t, err)

	var signedTx auth.StdTx
	err = cdc.UnmarshalBinaryLengthPrefixed(signedTxBytes, &signedTx)
	require.NoError(t, err)

	assert.Equal(t, tx.Msgs, signedTx.Msgs)
	assert.Equal(t, tx.Fee, signedTx.Fee)
	assert.Equal(t, tx.Memo, signedTx.Memo)

	require.Equal(t, 1, len(signedTx.Signatures))
	signature := signedTx.Signatures[0]
	var expectedPubKey secp256k1.PubKeySecp256k1
	copy(expectedPubKey[:], signerPubKey) // compressed public key

	assert.Equal(t, expectedPubKey, signature.PubKey)
	assert.Equal(t, mockSignBytes, signature.Signature)
}
