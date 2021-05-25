// Copyright 2021 Kava Labs, Inc.
// Copyright 2020 Coinbase, Inc.
//
// Derived from github.com/coinbase/rosetta-ethereum@f81889b
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
	"encoding/hex"

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// ConstructionParse implements the /construction/parse endpoint.
func (s *ConstructionAPIService) ConstructionParse(
	ctx context.Context,
	request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	txBytes, err := hex.DecodeString(request.Transaction)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	var tx auth.StdTx
	err = s.cdc.UnmarshalBinaryLengthPrefixed(txBytes, &tx)
	if err != nil {
		return nil, wrapErr(ErrInvalidTx, err)
	}

	index := int64(0)
	ops := []*types.Operation{}

	for _, msg := range tx.Msgs {
		msgSend, ok := msg.(bank.MsgSend)
		if !ok {
			continue
		}

		for _, coin := range msgSend.Amount {
			currency, ok := kava.Currencies[coin.Denom]
			if !ok {
				continue
			}

			ops = append(ops, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: index},
				Type:                kava.TransferOpType,
				Account:             &types.AccountIdentifier{Address: msgSend.FromAddress.String()},
				Amount:              &types.Amount{Value: "-" + coin.Amount.String(), Currency: currency},
			})

			ops = append(ops, &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{Index: index + 1},
				RelatedOperations:   []*types.OperationIdentifier{&types.OperationIdentifier{Index: index}},
				Type:                kava.TransferOpType,
				Account:             &types.AccountIdentifier{Address: msgSend.ToAddress.String()},
				Amount:              &types.Amount{Value: coin.Amount.String(), Currency: currency},
			})

			index += 2
		}
	}

	signers := []*types.AccountIdentifier{}
	for _, signature := range tx.Signatures {
		addr := sdk.AccAddress(signature.PubKey.Address().Bytes())
		signers = append(signers, &types.AccountIdentifier{
			Address: addr.String(),
		})
	}

	return &types.ConstructionParseResponse{
		Operations:               ops,
		AccountIdentifierSigners: signers,
	}, nil
}
