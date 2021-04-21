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

package kava

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth"
	kava "github.com/kava-labs/kava/app"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func parseTx(rawTx tmtypes.Tx, txResult *abci.ResponseDeliverTx) (*types.Transaction, error) {
	cdc := kava.MakeCodec()

	var decodedTx authtypes.StdTx
	err := cdc.UnmarshalBinaryLengthPrefixed(rawTx, &decodedTx)
	if err != nil {
		return &types.Transaction{}, err
	}

	status := SuccessStatus
	switch txResult {
	case nil:
		status = ""
	default:
		if txResult.Code != abci.CodeTypeOK {
			status = FailureStatus
		}
	}

	var totalOps []*types.Operation

	feeOps, err := getFeeOps(decodedTx, &status, nil)
	if err != nil {
		return nil, err
	}
	totalOps = append(totalOps, feeOps...)

	msgs := decodedTx.GetMsgs()

	var msgOps []*types.Operation
	for i, msg := range msgs {
		index := int64(i)
		ops, err := getMsgOps(msg, &status, &index)
		if err != nil {
			return nil, err
		}
		msgOps = append(msgOps, ops...)
	}

	totalOps = append(totalOps, msgOps...)
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: fmt.Sprintf("%X", rawTx.Hash())},
		Operations:            totalOps,
	}, nil
}
