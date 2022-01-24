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

//func TestConstructionCombine(t *testing.T) {
//	encodingConfig := app.MakeEncodingConfig()
//	networkIdentifier := &types.NetworkIdentifier{
//		Blockchain: "Kava",
//		Network:    "kava-test-9000",
//	}
//
//	txBuilder := encodingConfig.TxConfig.NewTxBuilder()
//
//	signerAddr := "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq"
//	signerAccAddr, err := sdk.AccAddressFromBech32(signerAddr)
//	require.NoError(t, err)
//	signerPubKey, err := base64.StdEncoding.DecodeString("AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY")
//	require.NoError(t, err)
//	toAccAddress, err := sdk.AccAddressFromBech32("kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w")
//	require.NoError(t, err)
//
//	txBuilder.SetMsgs(&banktypes.MsgSend{FromAddress: signerAccAddr.String(), ToAddress: toAccAddress.String(), Amount: sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(1000000)))})
//	txBuilder.SetGasLimit(250001)
//	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("ukava", sdk.NewInt(62501))))
//	txBuilder.SetMemo("some memo")
//
//	tx := txBuilder.GetTx()
//
//	txBytes, err := encodingConfig.TxConfig.TxEncoder()(tx)
//	require.NoError(t, err)
//	txPayload := hex.EncodeToString(txBytes)
//
//	// TODO: add real payload
//	//signBytes := auth.StdSignBytes(
//	//networkIdentifier.Network, 10, 11, tx.Fee, tx.Msgs, tx.Memo,
//	//)
//	signBytes := []byte("some payload")
//	mockSignatureBytes := []byte("some signature")
//
//	request := &types.ConstructionCombineRequest{
//		NetworkIdentifier:   networkIdentifier,
//		UnsignedTransaction: txPayload,
//		Signatures: []*types.Signature{
//			{
//				SigningPayload: &types.SigningPayload{
//					AccountIdentifier: &types.AccountIdentifier{Address: signerAccAddr.String()},
//					Bytes:             crypto.Sha256(signBytes),
//					SignatureType:     types.Ecdsa,
//				},
//				PublicKey: &types.PublicKey{
//					Bytes:     signerPubKey,
//					CurveType: types.Secp256k1,
//				},
//				SignatureType: types.Ecdsa,
//				Bytes:         mockSignatureBytes,
//			},
//		},
//	}
//
//	servicer, _ := setupConstructionAPIServicer()
//	ctx := context.Background()
//	response, rerr := servicer.ConstructionCombine(ctx, request)
//	require.Nil(t, rerr)
//
//	_ = response
//	//signedTxBytes, err := hex.DecodeString(response.SignedTransaction)
//	//require.NoError(t, err)
//
//	sdkTx, err := encodingConfig.TxConfig.TxDecoder()(txBytes)
//	require.NoError(t, err)
//
//	signedTx, ok := sdkTx.(authsigning.Tx)
//	require.True(t, ok)
//
//	assert.Equal(t, tx.GetMsgs(), signedTx.GetMsgs())
//	assert.Equal(t, tx.GetGas(), signedTx.GetGas())
//	assert.Equal(t, tx.GetFee(), signedTx.GetFee())
//	assert.Equal(t, tx.GetMemo(), signedTx.GetMemo())
//
//	// TODO: test signature
//	//require.Equal(t, 1, len(signedTx.Signatures))
//	//signature := signedTx.Signatures[0]
//	//var expectedPubKey secp256k1.PubKeySecp256k1
//	//copy(expectedPubKey[:], signerPubKey) // compressed public key
//
//	//assert.Equal(t, expectedPubKey, signature.PubKey)
//	//assert.Equal(t, mockSignBytes, signature.Signature)
//}
