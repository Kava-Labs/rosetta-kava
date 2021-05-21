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
	"encoding/json"
	"testing"

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/kava-labs/kava/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestConstructionPayloads(t *testing.T) {
	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: "Kava",
		Network:    "kava-test-9000",
	}

	signerAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
	signerPubKey, err := base64.StdEncoding.DecodeString("AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY")
	require.NoError(t, err)

	ops := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: signerAddr},
			Amount:              &types.Amount{Value: "-5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			RelatedOperations:   []*types.OperationIdentifier{{Index: 0}},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"},
			Amount:              &types.Amount{Value: "5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
	}

	signers := []signerInfo{
		{
			AccountNumber:   10,
			AccountSequence: 11,
		},
	}
	encodedSigners, err := json.Marshal(signers)
	require.NoError(t, err)

	metadata := map[string]interface{}{
		"signers":    string(encodedSigners),
		"gas_wanted": float64(250001),
		"gas_price":  float64(0.25),
		"memo":       "some memo",
	}

	pubkeys := []*types.PublicKey{
		{
			CurveType: types.Secp256k1,
			Bytes:     signerPubKey,
		},
	}

	request := &types.ConstructionPayloadsRequest{
		NetworkIdentifier: networkIdentifier,
		Operations:        ops,
		Metadata:          metadata,
		PublicKeys:        pubkeys,
	}

	servicer, _ := setupConstructionAPIServicer()
	ctx := context.Background()
	response, rerr := servicer.ConstructionPayloads(ctx, request)
	require.Nil(t, rerr)

	cdc := app.MakeCodec()
	txBytes, err := hex.DecodeString(response.UnsignedTransaction)
	require.NoError(t, err)
	var tx auth.StdTx
	err = cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	require.NoError(t, err)

	msgs := tx.GetMsgs()
	require.Equal(t, 1, len(msgs))
	msgSend, ok := msgs[0].(bank.MsgSend)
	require.True(t, ok)
	assert.Equal(t, signerAddr, msgSend.FromAddress.String())
	assert.Equal(t, auth.NewStdFee(250001, sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(62501)))), tx.Fee)
	assert.Equal(t, 0, len(tx.Signatures))
	assert.Equal(t, "some memo", tx.Memo)

	require.Equal(t, 1, len(response.Payloads))
	payload := response.Payloads[0]
	assert.Equal(t, signerAddr, payload.AccountIdentifier.Address)

	signBytes := auth.StdSignBytes(
		networkIdentifier.Network, 10, 11, tx.Fee, tx.Msgs, tx.Memo,
	)
	assert.Equal(t, crypto.Sha256(signBytes), payload.Bytes)
}
