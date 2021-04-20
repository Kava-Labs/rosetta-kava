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
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank"
)

// Questions
// 1. Harmony implements transfers as subtraction first, then addition, with the addition operation having a related operation. Is that correct? Cosmos-sdk doesn't seem to relate the operations at all.
// 2. In operation, Status seems to be a string in cosmos-sdk implementation, but *string in this. Was there a change?
// 3. In operation, cosmos-sdk doesn't seem to use an index - do we need to?

func msgSendToOperations(msg banktypes.MsgSend, status *string, startingOpIndex *int64) []*types.Operation {
	return transferToRosettaOperations(msg.FromAddress, msg.ToAddress, msg.Amount, status, startingOpIndex)
}

// transferToRosettaOperations converts a transfer from cosmos-sdk types to rosetta operations
// only accounts for ukava, hard, and usdx
// creates two operations per input coin, so if 2 coins are being transferred, 4 operations will be created
func transferToRosettaOperations(from, to sdk.AccAddress, amount sdk.Coins, status *string, startingOpIndex *int64) []*types.Operation {
	var opIndex int64
	if startingOpIndex != nil {
		opIndex = *startingOpIndex
	} else {
		opIndex = 0
	}

	transferAmount := getRosettaCoins(amount)

	operations := []*types.Operation{}

	for _, coin := range transferAmount {
		subOperationID := &types.OperationIdentifier{
			Index: opIndex,
		}
		subOp := &types.Operation{
			Type:    banktypes.EventTypeTransfer,
			Status:  status,
			Account: &types.AccountIdentifier{Address: to.String()},
			Amount: &types.Amount{
				Value: "-" + coin.Amount.String(), // use negative amount for sub-op
				Currency: &types.Currency{
					Symbol:   coin.Denom,
					Decimals: 0,
				},
			},
			OperationIdentifier: subOperationID,
		}
		addOperationID := &types.OperationIdentifier{
			Index: opIndex + 1,
		}
		addOp := &types.Operation{
			Type:    banktypes.EventTypeTransfer,
			Status:  status,
			Account: &types.AccountIdentifier{Address: to.String()},
			Amount: &types.Amount{
				Value: coin.Amount.String(),
				Currency: &types.Currency{
					Symbol:   coin.Denom,
					Decimals: 0,
				},
			},
			OperationIdentifier: addOperationID,
			RelatedOperations: []*types.OperationIdentifier{
				subOperationID,
			},
		}
		operations = append(operations, subOp, addOp)
	}
	return operations

}

// getRosettaCoins filters input coins for native assets (ukava, hard, usdx)
func getRosettaCoins(input sdk.Coins) sdk.Coins {
	outputCoins := sdk.NewCoins()
	for _, c := range input {
		if c.Denom == "ukava" || c.Denom == "hard" || c.Denom == "usdx" {
			outputCoins = outputCoins.Add(c)
		}
	}
	return outputCoins
}
