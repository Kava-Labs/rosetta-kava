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
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cdptypes "github.com/kava-labs/kava/x/cdp/types"
	kavadisttypes "github.com/kava-labs/kava/x/kavadist/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	feeCollectorAddress    = sdk.AccAddress(crypto.AddressHash([]byte(authtypes.FeeCollectorName)))
	mintModuleAddress      = sdk.AccAddress(crypto.AddressHash([]byte(minttypes.ModuleName)))
	kavaDistModuleAddress  = sdk.AccAddress(crypto.AddressHash([]byte(kavadisttypes.ModuleName)))
	stakingModuleAddress   = sdk.AccAddress(crypto.AddressHash([]byte(stakingtypes.BondedPoolName)))
	unbondingModuleAddress = sdk.AccAddress(crypto.AddressHash([]byte(stakingtypes.NotBondedPoolName)))
	cdpModuleAddress       = sdk.AccAddress(crypto.AddressHash([]byte(cdptypes.ModuleName)))
)

// EventsToOperations returns rosetta operations from abci block events
func EventsToOperations(events sdk.StringEvents, status *string, index int64) []*types.Operation {
	operations := []*types.Operation{}

	for _, event := range events {
		eventOps := EventToOperations(event, status, index)
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
	case banktypes.EventTypeTransfer:
		return bankTransferEventToOperations(attributeMap, status, index)
	case banktypes.EventTypeCoinMint:
		return bankMintEventToOperations(attributeMap, status, index)
	case banktypes.EventTypeCoinBurn:
		return bankBurnEventToOperations(attributeMap, status, index)
	// This is a block event only; we dot not yet track coin spent/received events implemented in v44
	case stakingtypes.EventTypeCompleteUnbonding:
		return completeUnbondingEventToOperations(attributeMap, status, index)
	}

	return []*types.Operation{}
}

func bankTransferEventToOperations(attributes map[string]string, status *string, index int64) []*types.Operation {
	recipient := &types.AccountIdentifier{
		Address: attributes[banktypes.AttributeKeyRecipient],
	}

	amount, err := sdk.ParseCoinsNormalized(attributes[sdk.AttributeKeyAmount])
	if err != nil {
		panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
	}

	sender := &types.AccountIdentifier{
		Address: attributes[banktypes.AttributeKeySender],
	}

	return balanceTrackingOps(TransferOpType, sender, amount, recipient, status, index)
}

func bankMintEventToOperations(attributes map[string]string, status *string, index int64) []*types.Operation {
	minter := &types.AccountIdentifier{
		Address: attributes[banktypes.AttributeKeyMinter],
	}

	amount, err := sdk.ParseCoinsNormalized(attributes[sdk.AttributeKeyAmount])
	if err != nil {
		panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
	}

	return accountBalanceOps(MintOpType, amount, false, minter, status, index)
}

func bankBurnEventToOperations(attributes map[string]string, status *string, index int64) []*types.Operation {
	burner := &types.AccountIdentifier{
		Address: attributes[banktypes.AttributeKeyBurner],
	}

	amount, err := sdk.ParseCoinsNormalized(attributes[sdk.AttributeKeyAmount])
	if err != nil {
		panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
	}

	return accountBalanceOps(BurnOpType, amount, true, burner, status, index)
}

// an unbonding event does not emit a transfer event -- it only emits a coin spent and coin received event
func completeUnbondingEventToOperations(attributes map[string]string, status *string, index int64) []*types.Operation {
	recipient := &types.AccountIdentifier{
		Address: attributes[stakingtypes.AttributeKeyDelegator],
	}

	amount, err := sdk.ParseCoinsNormalized(attributes[sdk.AttributeKeyAmount])
	if err != nil {
		panic(fmt.Sprintf("could not parse coins: %s", attributes[sdk.AttributeKeyAmount]))
	}

	sender := &types.AccountIdentifier{
		Address: unbondingModuleAddress.String(),
	}

	return balanceTrackingOps(TransferOpType, sender, amount, recipient, status, index)
}

// TxToOperations returns rosetta operations from a transaction
func TxToOperations(tx authsigning.Tx, logs sdk.ABCIMessageLogs, feeStatus *string, opStatus *string) []*types.Operation {
	operationIndex := int64(0)
	operations := []*types.Operation{}

	if !tx.GetFee().Empty() {
		feeOps := FeeToOperations(tx.FeePayer(), tx.GetFee(), feeStatus, operationIndex)
		operations = appendOperationsAndUpdateIndex(operations, feeOps, &operationIndex)
	}

	for msgIndex, msg := range tx.GetMsgs() {
		var log sdk.ABCIMessageLog

		if msgIndex < len(logs) {
			log = logs[msgIndex]
		} else {
			log = sdk.ABCIMessageLog{
				MsgIndex: uint32(msgIndex),
			}
		}

		msgOps := MsgToOperations(msg, log, opStatus, operationIndex)
		operations = appendOperationsAndUpdateIndex(operations, msgOps, &operationIndex)
	}

	return operations
}

// FeeToOperations returns rosetta operations from a transaction fee
func FeeToOperations(feePayer sdk.AccAddress, amount sdk.Coins, status *string, index int64) []*types.Operation {
	sender := newAccountID(feePayer.String())
	recipient := newAccountID(feeCollectorAddress.String())

	return balanceTrackingOps(FeeOpType, sender, amount, recipient, status, index)
}

// MsgToOperations returns rosetta operations for a cosmos sdk or kava message
func MsgToOperations(msg sdk.Msg, log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	ops := getOpsFromMsg(msg, log, status, index)

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

func newAccountID(addr string) *types.AccountIdentifier {
	return &types.AccountIdentifier{
		Address: addr,
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

func accountBalanceOps(
	opType string,
	amount sdk.Coins,
	negative bool,
	account *types.AccountIdentifier,
	status *string,
	index int64,
) []*types.Operation {
	operations := []*types.Operation{}

	for _, coin := range amount {
		currency, ok := Currencies[coin.Denom]
		if !ok {
			continue
		}

		value := coin.Amount.String()
		if negative {
			value = "-" + value
		}

		operations = append(operations, &types.Operation{
			OperationIdentifier: newOpID(index),
			Type:                opType,
			Status:              status,
			Account:             account,
			Amount: &types.Amount{
				Value:    value,
				Currency: currency,
			},
		})

		index++
	}

	return operations
}

func getOpsFromMsg(msg sdk.Msg, log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	var ops []*types.Operation

	if m, ok := msg.(*banktypes.MsgMultiSend); ok {
		transferOps := msgMultiSendToTransferOperations(m, status, index)
		ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
		return ops
	}

	for _, ev := range log.Events {
		if ev.Type == banktypes.EventTypeTransfer {
			events := unflattenEvents(ev, banktypes.EventTypeTransfer, 3)
			transferOps := EventsToOperations(events, status, index)
			ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
		}

		if ev.Type == banktypes.EventTypeCoinMint {
			events := unflattenEvents(ev, banktypes.EventTypeCoinMint, 2)
			transferOps := EventsToOperations(events, status, index)
			ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
		}

		if ev.Type == banktypes.EventTypeCoinBurn {
			events := unflattenEvents(ev, banktypes.EventTypeCoinBurn, 2)
			transferOps := EventsToOperations(events, status, index)
			ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
		}
	}

	// messages that do not give great events
	// TODO: re-check this in v45
	switch msg.(type) {
	case *stakingtypes.MsgDelegate:
		return msgDelegateToOperations(ops, log, status, index)
	case *stakingtypes.MsgCreateValidator:
		return msgCreateValidatorToOperations(ops, log, status, index)
	}

	// Gives contstruction support for msg send -- required for proper construction?
	if *status != SuccessStatus {
		switch m := msg.(type) {
		case *banktypes.MsgSend:
			transferOps := msgSendToTransferOperations(m, status, index)
			ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
		}
	}
	return ops
}

func msgSendToTransferOperations(msg *banktypes.MsgSend, status *string, index int64) []*types.Operation {
	sender := newAccountID(msg.FromAddress)
	recipient := newAccountID(msg.ToAddress)
	amount := msg.Amount

	return balanceTrackingOps(TransferOpType, sender, amount, recipient, status, index)
}

func msgMultiSendToTransferOperations(msg *banktypes.MsgMultiSend, status *string, index int64) []*types.Operation {
	ops := []*types.Operation{}

	for _, input := range msg.Inputs {
		sender := newAccountID(input.Address)
		transferOps := accountBalanceOps(TransferOpType, input.Coins, true, sender, status, index)
		ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
	}

	for _, output := range msg.Outputs {
		recipient := newAccountID(output.Address)
		transferOps := accountBalanceOps(TransferOpType, output.Coins, false, recipient, status, index)
		ops = appendOperationsAndUpdateIndex(ops, transferOps, &index)
	}

	return ops
}

func unflattenEvents(ev sdk.StringEvent, eventType string, numAttributes int) (events sdk.StringEvents) {
	if len(ev.Attributes)%numAttributes != 0 {
		panic(fmt.Sprintf("unexpected number of attributes in transfer event %s", ev.Attributes))
	}
	numberOfEvents := len(ev.Attributes) / numAttributes
	for i := 0; i < numberOfEvents; i++ {
		startingIndex := i * numAttributes
		event := sdk.NewEvent(eventType, ev.Attributes[startingIndex:startingIndex+numAttributes]...)
		events = append(events, sdk.StringifyEvent(abci.Event(event)))
	}
	return events
}

// Because there's no transfer event, and worse the `delegate` event doesn't contain the delegators' address, just the validator.
// A delegate message moves coins from an account to the staking module account - ideally there would be a transfer event in there, but there's not. Instead, I had to resort to parsing the delegate  and the message events to recreate the transfer
func msgDelegateToOperations(ops []*types.Operation, log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	var delegationOps []*types.Operation

	if len(log.Events) == 0 {
		return delegationOps
	}

	var (
		sender    string
		recipient string
		amount    sdk.Coins
	)
	for _, ev := range log.Events {
		if ev.Type == banktypes.EventTypeCoinSpent {
			for _, attr := range ev.Attributes {
				if attr.Key == banktypes.AttributeKeySpender {
					sender = attr.Value
				}

				if attr.Key == sdk.AttributeKeyAmount {
					amount = mustParseCoinsNormalized(attr.Value)
				}
			}
		}

		if ev.Type == banktypes.EventTypeCoinReceived {
			for _, attr := range ev.Attributes {
				if attr.Key == banktypes.AttributeKeyReceiver {
					recipient = attr.Value
				}
			}
		}
	}
	delegationOps = balanceTrackingOps(TransferOpType, newAccountID(sender), amount, newAccountID(recipient), status, index)
	return appendOperationsAndUpdateIndex(ops, delegationOps, &index)
}

// Because there's no transfer event, and worse the `create_validator` event doesn't contain the delegators' address, just the validator.
// A delegate message moves coins from an account to the staking module account - ideally there would be a transfer event in there, but there's not. Instead, I had to resort to parsing the delegate  and the message events to recreate the transfer
func msgCreateValidatorToOperations(ops []*types.Operation, log sdk.ABCIMessageLog, status *string, index int64) []*types.Operation {
	var createValidatorOps []*types.Operation

	if len(log.Events) == 0 {
		return createValidatorOps
	}

	var (
		sender    string
		recipient string
		amount    sdk.Coins
	)
	for _, ev := range log.Events {
		if ev.Type == banktypes.EventTypeCoinSpent {
			for _, attr := range ev.Attributes {
				if attr.Key == banktypes.AttributeKeySpender {
					sender = attr.Value
				}

				if attr.Key == sdk.AttributeKeyAmount {
					amount = mustParseCoinsNormalized(attr.Value)
				}
			}
		}

		if ev.Type == banktypes.EventTypeCoinReceived {
			for _, attr := range ev.Attributes {
				if attr.Key == banktypes.AttributeKeyReceiver {
					recipient = attr.Value
				}
			}
		}
	}
	createValidatorOps = balanceTrackingOps(TransferOpType, newAccountID(sender), amount, newAccountID(recipient), status, index)
	return appendOperationsAndUpdateIndex(ops, createValidatorOps, &index)
}

func mustAccAddressFromBech32(addr string) sdk.AccAddress {
	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		panic(err)
	}
	return acc
}

func mustParseCoinsNormalized(coinsStr string) sdk.Coins {
	coins, err := sdk.ParseCoinsNormalized(coinsStr)
	if err != nil {
		panic(err)
	}
	return coins
}
