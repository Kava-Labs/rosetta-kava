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

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestConstructionParse_Unsigned(t *testing.T) {
	signerAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	signerAccAddr, err := sdk.AccAddressFromBech32(signerAddr)
	require.NoError(t, err)
	toAccAddress, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	tx := auth.NewStdTx(
		[]sdk.Msg{bank.MsgSend{FromAddress: signerAccAddr, ToAddress: toAccAddress, Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))}},
		auth.NewStdFee(250001, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(62501)))),
		[]auth.StdSignature{},
		"some memo",
	)

	cdc := app.MakeCodec()
	txBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
	txPayload := hex.EncodeToString(txBytes)

	request := &types.ConstructionParseRequest{
		Signed:      false,
		Transaction: txPayload,
	}

	servicer, _ := setupConstructionAPIServicer()
	ctx := context.Background()
	response, rerr := servicer.ConstructionParse(ctx, request)
	require.Nil(t, rerr)

	require.Equal(t, 2, len(response.Operations))
	require.Equal(t, 0, len(response.AccountIdentifierSigners))

	op1 := response.Operations[0]
	assert.Equal(t, int64(0), op1.OperationIdentifier.Index)
	assert.Equal(t, signerAddr, op1.Account.Address)
	assert.Equal(t, kava.TransferOpType, op1.Type)
	assert.Equal(t, "KAVA", op1.Amount.Currency.Symbol)
	assert.Equal(t, int32(6), op1.Amount.Currency.Decimals)
	assert.Equal(t, "-1000000", op1.Amount.Value)

	op2 := response.Operations[1]
	assert.Equal(t, int64(1), op2.OperationIdentifier.Index)
	require.Equal(t, 1, len(op2.RelatedOperations))
	assert.Equal(t, int64(0), op2.RelatedOperations[0].Index)
	assert.Equal(t, toAccAddress.String(), op2.Account.Address)
	assert.Equal(t, kava.TransferOpType, op2.Type)
	assert.Equal(t, "KAVA", op2.Amount.Currency.Symbol)
	assert.Equal(t, int32(6), op2.Amount.Currency.Decimals)
	assert.Equal(t, "1000000", op2.Amount.Value)
}

func TestConstructionParse_Signed(t *testing.T) {
	signerAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	signerAccAddr, err := sdk.AccAddressFromBech32(signerAddr)
	require.NoError(t, err)
	signerPubKey, err := base64.StdEncoding.DecodeString("AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY")
	require.NoError(t, err)
	toAccAddress, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
	require.NoError(t, err)

	var pubkey secp256k1.PubKeySecp256k1
	copy(pubkey[:], signerPubKey) // compressed public key

	tx := auth.NewStdTx(
		[]sdk.Msg{bank.MsgSend{FromAddress: signerAccAddr, ToAddress: toAccAddress, Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))}},
		auth.NewStdFee(250001, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(62501)))),
		[]auth.StdSignature{
			{
				PubKey:    pubkey,
				Signature: []byte("some signature"),
			},
		},
		"some memo",
	)

	cdc := app.MakeCodec()
	txBytes, err := cdc.MarshalBinaryLengthPrefixed(tx)
	require.NoError(t, err)
	txPayload := hex.EncodeToString(txBytes)

	request := &types.ConstructionParseRequest{
		Signed:      true,
		Transaction: txPayload,
	}

	servicer, _ := setupConstructionAPIServicer()
	ctx := context.Background()
	response, rerr := servicer.ConstructionParse(ctx, request)
	require.Nil(t, rerr)

	require.Equal(t, 2, len(response.Operations))

	op1 := response.Operations[0]
	assert.Equal(t, int64(0), op1.OperationIdentifier.Index)
	assert.Equal(t, signerAddr, op1.Account.Address)
	assert.Equal(t, kava.TransferOpType, op1.Type)
	assert.Equal(t, "KAVA", op1.Amount.Currency.Symbol)
	assert.Equal(t, int32(6), op1.Amount.Currency.Decimals)
	assert.Equal(t, "-1000000", op1.Amount.Value)

	op2 := response.Operations[1]
	assert.Equal(t, int64(1), op2.OperationIdentifier.Index)
	require.Equal(t, 1, len(op2.RelatedOperations))
	assert.Equal(t, int64(0), op2.RelatedOperations[0].Index)
	assert.Equal(t, toAccAddress.String(), op2.Account.Address)
	assert.Equal(t, kava.TransferOpType, op2.Type)
	assert.Equal(t, "KAVA", op2.Amount.Currency.Symbol)
	assert.Equal(t, int32(6), op2.Amount.Currency.Decimals)
	assert.Equal(t, "1000000", op2.Amount.Value)

	require.Equal(t, 1, len(response.AccountIdentifierSigners))
	account := response.AccountIdentifierSigners[0]
	assert.Equal(t, signerAddr, account.Address)
}
