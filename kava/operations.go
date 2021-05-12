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
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	mint "github.com/cosmos/cosmos-sdk/x/mint"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	feeCollectorAddress = sdk.AccAddress(crypto.AddressHash([]byte(authtypes.FeeCollectorName)))
	mintModuleAddress   = sdk.AccAddress(crypto.AddressHash([]byte(mint.ModuleName)))
)

// EventsToOperations returns rosetta operations from abci block events
func EventsToOperations(events sdk.StringEvents, index int64) []*types.Operation {
	status := SuccessStatus
	operations := []*types.Operation{}

	for _, event := range events {
		eventOps := EventToOperations(event, &status, index)
		operations = appendOperationsAndUpdateIndex(operations, eventOps, &index)
	}

	return operations
}

// EventToOperations returns rosetta operations from a abci block event
func EventToOperations(event sdk.StringEvent, status *string, index int64) []*types.Operation {
	attributeMap := make(map[string]string)

	for _, attribute := range event.Attributes {
		attributeMap[attribute.Key] = attribute.Value
	}

	switch event.Type {
	case bank.EventTypeTransfer:
		return bankTransferEventToOperations(attributeMap, status, index)
	}

	return []*types.Operation{}
}

func bankTransferEventToOperations(attributes map[string]string, status *string, index int64) []*types.Operation {
	recipient := &types.AccountIdentifier{
		Address: attributes[bank.AttributeKeyRecipient],
	}

	amount, err := sdk.ParseCoins(attributes[sdk.AttributeKeyAmount])
	if err != nil {
		panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
	}

	if attributes[bank.AttributeKeySender] == mintModuleAddress.String() {
		return recipientBalanceOps(MintOpType, amount, recipient, status, index)
	}

	sender := &types.AccountIdentifier{
		Address: attributes[bank.AttributeKeySender],
	}

	return balanceTrackingOps(TransferOpType, sender, amount, recipient, status, index)
}

// TxToOperations returns rosetta operations from a transaction
func TxToOperations(tx *authtypes.StdTx, logs sdk.ABCIMessageLogs, status *string) []*types.Operation {
	operationIndex := int64(0)
	operations := []*types.Operation{}

	if !tx.Fee.Amount.Empty() {
		feeStatus := SuccessStatus
		feeOps := FeeToOperations(tx.FeePayer(), tx.Fee.Amount, &feeStatus, operationIndex)
		operations = appendOperationsAndUpdateIndex(operations, feeOps, &operationIndex)
	}

	for msgIndex, msg := range tx.GetMsgs() {
		var log sdk.ABCIMessageLog

		if msgIndex < len(logs) {
			log = logs[msgIndex]
		} else {
			log = sdk.ABCIMessageLog{
				MsgIndex: uint16(msgIndex),
			}
		}

		msgOps := MsgToOperations(msg, log, status, operationIndex)
		operations = appendOperationsAndUpdateIndex(operations, msgOps, &operationIndex)
	}

	return operations
}

// FeeToOperations returns rosetta operations from a transaction fee
func FeeToOperations(feePayer sdk.AccAddress, amount sdk.Coins, status *string, index int64) []*types.Operation {
	sender := newAccountID(feePayer)
	recipient := newAccountID(feeCollectorAddress)

	return balanceTrackingOps(FeeOpType, sender, amount, recipient, status, index)
}

// MsgToOperations returns rosetta operations for a cosmos sdk or kava message
func MsgToOperations(msg sdk.Msg, log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	ops := getTransferOpsFromMsg(log, status, index)
	return ops
}

func appendOperationsAndUpdateIndex(
	operations []*types.Operation,
	newOps []*types.Operation,
	index *int64,
) []*types.Operation {
	*index += int64(len(newOps))
	return append(operations, newOps...)
}

func newOpID(index int64) *types.OperationIdentifier {
	return &types.OperationIdentifier{
		Index: index,
	}
}

func newAccountID(addr sdk.AccAddress) *types.AccountIdentifier {
	return &types.AccountIdentifier{
		Address: addr.String(),
	}
}

func balanceTrackingOps(
	opType string,
	sender *types.AccountIdentifier,
	amount sdk.Coins,
	recipient *types.AccountIdentifier,
	status *string,
	index int64,
) []*types.Operation {
	operations := []*types.Operation{}

	for _, coin := range amount {
		currency, ok := Currencies[coin.Denom]
		if !ok {
			continue
		}

		operations = append(operations, &types.Operation{
			OperationIdentifier: newOpID(index),
			Type:                opType,
			Status:              status,
			Account:             sender,
			Amount: &types.Amount{
				Value:    "-" + coin.Amount.String(),
				Currency: currency,
			},
		})

		operations = append(operations, &types.Operation{
			OperationIdentifier: newOpID(index + 1),
			RelatedOperations:   []*types.OperationIdentifier{newOpID(index)},
			Type:                opType,
			Status:              status,
			Account:             recipient,
			Amount: &types.Amount{
				Value:    coin.Amount.String(),
				Currency: currency,
			},
		})

		index += 2
	}

	return operations
}

func recipientBalanceOps(
	opType string,
	amount sdk.Coins,
	recipient *types.AccountIdentifier,
	status *string,
	index int64,
) []*types.Operation {
	operations := []*types.Operation{}

	for _, coin := range amount {
		currency, ok := Currencies[coin.Denom]
		if !ok {
			continue
		}

		operations = append(operations, &types.Operation{
			OperationIdentifier: newOpID(index),
			Type:                opType,
			Status:              status,
			Account:             recipient,
			Amount: &types.Amount{
				Value:    coin.Amount.String(),
				Currency: currency,
			},
		})

		index++
	}

	return operations
}

func getTransferOpsFromMsg(log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	var ops []*types.Operation
	for _, ev := range log.Events {
		if ev.Type == "transfer" {
			unflattenedTransferEvents := unflattenTransferEvents(ev)
			for _, event := range unflattenedTransferEvents {
				transferOps := getTransferOpsFromEvent(event, status, index)
				ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
			}
		}
	}
	return ops
}

func unflattenTransferEvents(ev sdk.StringEvent) (events sdk.StringEvents) {
	if len(ev.Attributes)%3 != 0 {
		panic(fmt.Sprintf("unexpected number of attributes in transfer event %s", ev.Attributes))
	}
	numberOfTransferEvents := len(ev.Attributes) / 3
	for i := 0; i < numberOfTransferEvents; i++ {
		startingIndex := i * 3
		event := sdk.NewEvent(bank.EventTypeTransfer, ev.Attributes[startingIndex:startingIndex+3]...)
		events = append(events, sdk.StringifyEvent(abci.Event(event)))
	}
	return events
}

func getTransferOpsFromEvent(ev sdk.StringEvent, status *string, index int64) []*types.Operation {
	var sender sdk.AccAddress
	var recipient sdk.AccAddress
	var amount sdk.Coins
	for _, attr := range ev.Attributes {
		if attr.Key == "sender" {
			sender = mustAccAddressFromBech32(attr.Value)
		}
		if attr.Key == "recipient" {
			recipient = mustAccAddressFromBech32(attr.Value)
		}
		if attr.Key == "amount" {
			amount = mustParseCoins(attr.Value)
		}
	}
	return balanceTrackingOps(TransferOpType, newAccountID(sender), amount, newAccountID(recipient), status, index)
}

func mustAccAddressFromBech32(addr string) sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return acc
}

func mustParseCoins(coinsStr string) sdk.Coins {
	coins, err := sdk.ParseCoins(coinsStr)
	if err != nil {
		panic(err)
	}
	return coins
}
