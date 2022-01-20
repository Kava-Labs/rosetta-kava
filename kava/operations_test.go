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
	"io/ioutil"
	"math/big"
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kava-labs/kava/app"
)

const (
	mintAddress = "kava1m3h30wlvsf8llruxtpukdvsy0km2kum85yn938"
)

var (
	testAddresses = []string{
		"kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea",
		"kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w",
		"kava16g8lzm86f5wwf3x3t67qrpd46sjdpxpfazskwg",
		"kava1wn74shl496ktcfgqsc6yf0vvenhgq0hwuw6z2a",
	}
)

func getAccAddr(t *testing.T, addr string) sdk.AccAddress {
	accAddr, err := sdk.AccAddressFromBech32(addr)
	require.NoError(t, err)
	return accAddr
}

func accountEqual(a1 *types.AccountIdentifier, a2 *types.AccountIdentifier) bool {
	if a1.Address != a2.Address {
		return false
	}

	if a1.SubAccount == nil && a2.SubAccount == nil {
		return true
	}

	if a1.SubAccount != nil && a2.SubAccount == nil {
		return false
	}

	if a2.SubAccount != nil && a1.SubAccount == nil {
		return false
	}

	if a1.SubAccount.Address != a2.SubAccount.Address {
		return false
	}

	return true
}

func generateDefaultCoins() sdk.Coins {
	denoms := []string{
		// native
		"ukava", "hard", "usdx", "swp",
		// not native bep2 assets
		"bnb", "busd", "btcb",
		// ibc assets
		"ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2A",
	}

	return generateCoins(denoms)
}

func generateCoins(denoms []string) sdk.Coins {
	coins := sdk.Coins{}

	for _, denom := range denoms {
		coins = append(coins, sdk.Coin{
			Denom:  denom,
			Amount: sdk.NewInt(int64(rand.Intn(1000 * 1e6))),
		})
	}

	return coins.Sort()
}

type operationTestCase struct {
	// name of test case
	name string
	// coins to use when generating data
	coins sdk.Coins
	// status for operations
	status string
	// starting index for operations
	index int64
}

var operationTestCases = []operationTestCase{
	{
		name:   "success status",
		coins:  generateDefaultCoins(),
		status: SuccessStatus,
		index:  0,
	},
	{
		name:   "failure status",
		coins:  generateDefaultCoins(),
		status: FailureStatus,
		index:  0,
	},
	{
		name:   "non-zero starting index",
		coins:  generateDefaultCoins(),
		status: SuccessStatus,
		index:  10,
	},
	{
		name:   "single denom",
		coins:  generateCoins([]string{"ukava"}),
		status: SuccessStatus,
		index:  0,
	},
	{
		name:   "non-native single denom",
		coins:  generateCoins([]string{"busd"}),
		status: SuccessStatus,
		index:  0,
	},
}

func runAndAssertOperationInvariants(
	t *testing.T,
	opType string,
	testFn func(*operationTestCase) []*types.Operation,
) {
	for _, tc := range operationTestCases {
		t.Run(tc.name, func(t *testing.T) {
			ops := testFn(&tc)

			for index, op := range ops {
				// assert index, status, and op type all match
				expectedIndex := int64(index) + tc.index
				assert.Equal(t, expectedIndex, op.OperationIdentifier.Index)
				assert.Equal(t, &tc.status, op.Status)

				// TODO: improve op type tracking
				//assert.Equal(t, opType, op.Type)

				// assert related operations have a lower index
				if op.RelatedOperations != nil {
					for _, relatedOpID := range op.RelatedOperations {
						assert.Greater(t, op.OperationIdentifier.Index, relatedOpID.Index)
					}
				}

				// assert operation type is supported
				assert.Contains(t, OperationTypes, op.Type)

				// assert seen currencies are supported and correct
				symbol := op.Amount.Currency.Symbol
				denom, ok := Denoms[symbol]
				assert.Truef(t, ok, "currency %s not supported", symbol)
				if ok {
					assert.Equal(t, Currencies[denom], op.Amount.Currency)
				}
			}
		})
	}
}

func assertTrackedBalance(
	t *testing.T,
	name string,
	ops []*types.Operation,
	sender *types.AccountIdentifier,
	amount sdk.Coins,
	recipient *types.AccountIdentifier,
) {
	t.Run(name, func(t *testing.T) {
		supportedCurrenciesFound := false
		for _, coin := range amount {
			_, ok := Currencies[coin.Denom]
			if ok {
				supportedCurrenciesFound = true
			}
		}

		if supportedCurrenciesFound && len(ops) == 0 {
			t.Fatal("no operations found")
		}

		t.Run("coin operations sum to zero", func(t *testing.T) {
			if sender == nil || recipient == nil {
				t.Skip("no sender or recipient to match operations")
			}

			coinSums := make(map[string]*big.Int)

			for _, op := range ops {
				symbol := op.Amount.Currency.Symbol
				value, err := types.AmountValue(op.Amount)
				require.NoError(t, err)

				sum, ok := coinSums[symbol]
				if ok {
					sum.Add(sum, value)
				} else {
					coinSums[symbol] = value
				}
			}

			for _, sum := range coinSums {
				assert.True(t, big.NewInt(0).Cmp(sum) == 0)
			}
		})

		t.Run("coin operation amounts match for sender", func(t *testing.T) {
			if sender == nil {
				t.Skip("no sender")
			}

			for _, op := range ops {
				if !accountEqual(op.Account, sender) {
					continue
				}

				symbol := op.Amount.Currency.Symbol
				value, err := types.AmountValue(op.Amount)
				require.NoError(t, err)

				denom, ok := Denoms[symbol]
				require.True(t, ok)

				// sender operations are negative
				assert.Equal(t, amount.AmountOf(denom).String(), value.Neg(value).String())
			}
		})

		t.Run("coin operation amounts match for recipient", func(t *testing.T) {
			if recipient == nil {
				t.Skip("no recipient")
			}

			for _, op := range ops {
				if !accountEqual(op.Account, recipient) {
					continue
				}

				symbol := op.Amount.Currency.Symbol
				value, err := types.AmountValue(op.Amount)
				require.NoError(t, err)

				denom, ok := Denoms[symbol]
				require.True(t, ok)

				// recipient operations are negative
				assert.Equal(t, amount.AmountOf(denom).String(), value.String())
			}
		})

		t.Run("all operations are for sender or recipient", func(t *testing.T) {
			for _, op := range ops {
				assert.NotNil(t, op.Account)
				assert.Contains(t, []*types.AccountIdentifier{sender, recipient}, op.Account)
			}
		})

		t.Run("each sender op has no related ops", func(t *testing.T) {
			if sender == nil {
				t.Skip("no sender")
			}

			for _, op := range ops {
				if !accountEqual(op.Account, sender) {
					continue
				}

				assert.Equal(t, 0, len(op.RelatedOperations))
			}
		})

		t.Run("each recipient op is related to a sender op", func(t *testing.T) {
			if sender == nil {
				t.Skip("no sender")
			}

			for _, op := range ops {
				if !accountEqual(op.Account, recipient) {
					continue
				}

				require.Equal(t, 1, len(op.RelatedOperations))
				relatedOpIndex := op.RelatedOperations[0].Index
				relatedOp := ops[relatedOpIndex-ops[0].OperationIdentifier.Index]

				// index matches as expected
				assert.Equal(t, relatedOpIndex, relatedOp.OperationIdentifier.Index)

				// related op is sender address
				assert.Equal(t, relatedOp.Account, sender)

				// values match
				negatedRelValue, err := types.NegateValue(relatedOp.Amount.Value)
				require.NoError(t, err)
				assert.Equal(t, op.Amount.Value, negatedRelValue)

				// currencies match
				assert.Equal(t, op.Amount.Value, negatedRelValue)
			}
		})
	})
}

func TestEventsToOperations(t *testing.T) {
	testEvent1 := sdk.StringEvent{
		Type: banktypes.EventTypeTransfer,
		Attributes: []sdk.Attribute{
			{
				Key:   banktypes.AttributeKeyRecipient,
				Value: testAddresses[1],
			},
			{
				Key:   banktypes.AttributeKeySender,
				Value: testAddresses[0],
			},
			{
				Key:   sdk.AttributeKeyAmount,
				Value: generateDefaultCoins().String(),
			},
		},
	}

	testEvent2 := sdk.StringEvent{
		Type: banktypes.EventTypeTransfer,
		Attributes: []sdk.Attribute{
			{
				Key:   banktypes.AttributeKeyRecipient,
				Value: testAddresses[2],
			},
			{
				Key:   banktypes.AttributeKeySender,
				Value: testAddresses[1],
			},
			{
				Key:   sdk.AttributeKeyAmount,
				Value: generateDefaultCoins().String(),
			},
		},
	}

	index := int64(0)
	events := sdk.StringEvents{testEvent1, testEvent2}
	status := SuccessStatus
	ops := EventsToOperations(events, &status, index)

	assert.Greater(t, len(ops), 0)
	for opIndex, op := range ops {
		assert.Equal(t, int64(opIndex)+index, op.OperationIdentifier.Index)
		assert.Equal(t, SuccessStatus, *op.Status)
	}

	index = int64(10)
	events = sdk.StringEvents{testEvent1, testEvent2}
	ops = EventsToOperations(events, &status, index)

	assert.Greater(t, len(ops), 0)
	for opIndex, op := range ops {
		assert.Equal(t, int64(opIndex)+index, op.OperationIdentifier.Index)
		assert.Equal(t, SuccessStatus, *op.Status)
	}
}

func TestEventToOperations(t *testing.T) {
	tests := []struct {
		name      string
		createFn  func(coins sdk.Coins) sdk.StringEvent
		opType    string
		sender    *types.AccountIdentifier
		recipient *types.AccountIdentifier
	}{
		{
			name: "mint (transfer from mint module acct)",
			createFn: func(coins sdk.Coins) sdk.StringEvent {
				return sdk.StringEvent{
					Type: banktypes.EventTypeTransfer,
					Attributes: []sdk.Attribute{
						{
							Key:   banktypes.AttributeKeyRecipient,
							Value: testAddresses[0],
						},
						{
							Key:   banktypes.AttributeKeySender,
							Value: mintAddress,
						},
						{
							Key:   sdk.AttributeKeyAmount,
							Value: coins.String(),
						},
					},
				}
			},
			opType: MintOpType,
			sender: nil, // no sender for mint operations
			recipient: &types.AccountIdentifier{
				Address: testAddresses[0],
			},
		},
		{
			name: "trackable transfer (not mint or burn)",
			createFn: func(coins sdk.Coins) sdk.StringEvent {
				return sdk.StringEvent{
					Type: banktypes.EventTypeTransfer,
					Attributes: []sdk.Attribute{
						{
							Key:   banktypes.AttributeKeyRecipient,
							Value: testAddresses[1],
						},
						{
							Key:   banktypes.AttributeKeySender,
							Value: testAddresses[0],
						},
						{
							Key:   sdk.AttributeKeyAmount,
							Value: coins.String(),
						},
					},
				}
			},
			opType: TransferOpType,
			sender: &types.AccountIdentifier{
				Address: testAddresses[0],
			},
			recipient: &types.AccountIdentifier{
				Address: testAddresses[1],
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runAndAssertOperationInvariants(t, tc.opType, func(otc *operationTestCase) []*types.Operation {
				event := tc.createFn(otc.coins)
				ops := EventToOperations(event, &otc.status, otc.index)

				assertTrackedBalance(t, otc.name, ops, tc.sender, otc.coins, tc.recipient)

				return ops
			})
		})
	}
}

func TestTxToOperations(t *testing.T) {
	msg1 := banktypes.MsgSend{
		FromAddress: getAccAddr(t, testAddresses[0]).String(),
		ToAddress:   getAccAddr(t, testAddresses[1]).String(),
		Amount:      generateDefaultCoins(),
	}

	msg2 := banktypes.MsgSend{
		FromAddress: getAccAddr(t, testAddresses[0]).String(),
		ToAddress:   getAccAddr(t, testAddresses[1]).String(),
		Amount:      generateDefaultCoins(),
	}

	// one less than message length
	logs := sdk.ABCIMessageLogs{
		sdk.ABCIMessageLog{},
	}

	success := SuccessStatus
	failure := FailureStatus

	t.Run("no fee", func(t *testing.T) {
		tx := legacytx.StdTx{
			Msgs: []sdk.Msg{&msg1, &msg2},
			Fee:  legacytx.StdFee{Gas: 500000},
		}

		// all ops succesful and indexed correctly
		ops := TxToOperations(&tx, logs, &success, &success)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, success, *op.Status)
		}

		// all ops failed and indexed correctly
		ops = TxToOperations(&tx, logs, &failure, &failure)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, failure, *op.Status)
		}

		// there are no fee operations
		ops = TxToOperations(&tx, logs, &success, &success)
		for _, op := range ops {
			assert.NotEqual(t, FeeOpType, op.Type)
		}
	})

	t.Run("with fee", func(t *testing.T) {
		tx := legacytx.StdTx{
			Msgs: []sdk.Msg{&msg1, &msg2},
			Fee:  legacytx.StdFee{Amount: generateCoins([]string{"ukava"}), Gas: 500000},
		}

		// all ops succesful and indexed correctly
		ops := TxToOperations(&tx, logs, &success, &success)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, success, *op.Status)
		}

		// all ops failed and indexed correctly
		ops = TxToOperations(&tx, logs, &failure, &failure)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, failure, *op.Status)
		}

		// there are fee operations
		feeOpTypeFound := false
		ops = TxToOperations(&tx, logs, &success, &success)
		for _, op := range ops {
			if op.Type == FeeOpType {
				feeOpTypeFound = true
			}
		}

		assert.True(t, feeOpTypeFound)
	})
}

func TestFeeToOperations(t *testing.T) {
	// fee payer of the transaction, passed explicity
	payer := testAddresses[0]
	payerAddr, err := sdk.AccAddressFromBech32(payer)
	require.NoError(t, err)

	// fixed for all fee operations
	feeCollector := "kava17xpfvakm2amg962yls6f84z3kell8c5lvvhaa6"

	// assert balance tracking and operation invarians are correct
	runAndAssertOperationInvariants(t, FeeOpType, func(tc *operationTestCase) []*types.Operation {
		ops := FeeToOperations(payerAddr, tc.coins, &tc.status, tc.index)

		sender := &types.AccountIdentifier{
			Address: payer,
		}
		recipient := &types.AccountIdentifier{
			Address: feeCollector,
		}

		assertTrackedBalance(t, tc.name, ops, sender, tc.coins, recipient)

		return ops
	})
}

func TestMsgToOperations_BalanceTracking(t *testing.T) {
	tests := []struct {
		name string
		log  sdk.ABCIMessageLog
		msg  sdk.Msg
	}{
		{
			name: "hard.MsgDeposit",
			log:  readABCILogFromFile(t, "hard-deposit-tx-response.json"),
		},
		{
			name: "hard.MsgWithdraw",
			log:  readABCILogFromFile(t, "hard-withdraw-tx-response.json"),
		},
		{
			name: "hard.MsgBorrow",
			log:  readABCILogFromFile(t, "hard-borrow-tx-response.json"),
		},
		{
			name: "hard.MsgRepay",
			log:  readABCILogFromFile(t, "hard-repay-tx-response.json"),
		},
		{
			name: "hard.MsgLiquidate",
			log:  readABCILogFromFile(t, "hard-liquidate-tx-response.json"),
		},
		{
			name: "auction.MsgPlaceBid",
			log:  readABCILogFromFile(t, "auction-bid-tx-response.json"),
		},
		{
			name: "bep3.MsgCreateAtomicSwap",
			log:  readABCILogFromFile(t, "bep3-create-tx-response.json"),
		},
		{
			name: "bep3.MsgRefundAtomicSwap",
			log:  readABCILogFromFile(t, "bep3-refund-tx-response.json"),
		},
		{
			name: "bep3.MsgClaimAtomicSwap",
			log:  readABCILogFromFile(t, "bep3-claim-tx-response.json"),
		},
		{
			name: "cdp.MsgCreateCDP",
			log:  readABCILogFromFile(t, "cdp-create-tx-response.json"),
		},
		{
			name: "cdp.MsgDeposit",
			log:  readABCILogFromFile(t, "cdp-deposit-tx-response.json"),
		},
		{
			name: "cdp.MsgWithdraw",
			log:  readABCILogFromFile(t, "cdp-withdraw-tx-response.json"),
		},
		{
			name: "cdp.MsgDrawDebt",
			log:  readABCILogFromFile(t, "cdp-draw-tx-response.json"),
		},
		{
			name: "cdp.MsgRepayDebt",
			log:  readABCILogFromFile(t, "cdp-repay-tx-response.json"),
		},
		{
			name: "cdp.MsgLiquidate",
			log:  readABCILogFromFile(t, "cdp-liquidate-tx-response.json"),
		},
		{
			name: "kava.SubmitProposal",
			log:  readABCILogFromFile(t, "committee-submit-tx-response.json"),
		},
		{
			name: "kava.MsgVote",
			log:  readABCILogFromFile(t, "committee-vote-tx-response.json"),
		},
		{
			name: "incentive.MsgClaimUSDXMintingReward",
			log:  readABCILogFromFile(t, "incentive-claim-usdx-tx-response.json"),
		},
		{
			name: "incentive.MsgClaimHardReward",
			log:  readABCILogFromFile(t, "incentive-claim-hard-tx-response.json"),
		},
		{
			name: "pricefeed.MsgPostPrice",
			log:  readABCILogFromFile(t, "pricefeed-post-tx-response.json"),
		},
		{
			name: "cosmos-sdk.MsgSend",
			log:  readABCILogFromFile(t, "msg-send-tx-response.json"),
		},
		{
			name: "cosmos-sdk.MsgMultiSend",
			log:  readABCILogFromFile(t, "msg-multisend-tx-response.json"),
			msg:  readMsgFromFile(t, "msg-multisend-tx-response.json"),
		},
		{
			name: "cosmos-sdk.MsgDelegate",
			log:  readABCILogFromFile(t, "msg-delegate-tx-response.json"),
			msg:  readMsgFromFile(t, "msg-delegate-tx-response.json"),
		},
		{
			name: "cosmos-sdk.MsgCreateValidator",
			log:  readABCILogFromFile(t, "msg-create-validator-tx-response.json"),
			msg:  readMsgFromFile(t, "msg-create-validator-tx-response.json"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runAndAssertOperationInvariants(t, TransferOpType,
				func(otc *operationTestCase) []*types.Operation {
					ops := MsgToOperations(tc.msg, tc.log, &otc.status, otc.index)
					senders, receivers := calculateSendersReceivers(tc.msg, tc.log)
					coins := calculateCoins(tc.log)
					assertTransferOpsBalanceTrack(t, otc.name, ops, senders, receivers, coins)
					return ops
				})
		})
	}
}

func assertTransferOpsBalanceTrack(
	t *testing.T,
	name string,
	ops []*types.Operation,
	senderTracking []accountBalance,
	receiverTracking []accountBalance,
	transferCoins sdk.Coins,
) {
	t.Run(name, func(t *testing.T) {
		supportedCurrenciesFound := false
		for _, coin := range transferCoins {
			_, ok := Currencies[coin.Denom]
			if ok {
				supportedCurrenciesFound = true
			}
		}

		if supportedCurrenciesFound && len(ops) == 0 {
			t.Fatal("no operations found")
		}
	})

	t.Run("coin transfer operations sum to zero", func(t *testing.T) {
		if len(senderTracking) == 0 || len(receiverTracking) == 0 {
			t.Skip("no sender or recipient to match operations")
		}

		coinSums := make(map[string]*big.Int)

		for _, op := range ops {
			if op.Type != TransferOpType {
				continue
			}

			symbol := op.Amount.Currency.Symbol
			value, err := types.AmountValue(op.Amount)
			require.NoError(t, err)

			sum, ok := coinSums[symbol]
			if ok {
				sum.Add(sum, value)
			} else {
				coinSums[symbol] = value
			}
		}

		for _, sum := range coinSums {
			assert.True(t, big.NewInt(0).Cmp(sum) == 0)
		}
	})

	t.Run("coin transfer operation amounts match for sender", func(t *testing.T) {
		for _, st := range senderTracking {
			if st.Account == nil {
				t.Skip("no sender")
			}
			opCoins := sdk.NewCoins()
			for _, op := range ops {
				if op.Type != TransferOpType {
					continue
				}

				if !mustAccAddressFromBech32(op.Account.Address).Equals(st.Account) {
					continue
				}

				symbol := op.Amount.Currency.Symbol
				value, err := types.AmountValue(op.Amount)
				if value.Sign() != -1 {
					continue // exit if value is non-negative, as this is not a send
				}
				require.NoError(t, err)

				denom, ok := Denoms[symbol]
				require.True(t, ok)

				opCoins = opCoins.Add(sdk.NewCoin(denom, sdk.NewIntFromBigInt(value.Neg(value))))
			}
			assert.True(t, opCoins.IsEqual(st.Balance))
		}
	})

	t.Run("coin transfer operation amounts match for recipient", func(t *testing.T) {
		for _, rt := range receiverTracking {
			if rt.Account == nil {
				t.Skip("no recipient")
			}
			opCoins := sdk.NewCoins()
			for _, op := range ops {
				if op.Type != TransferOpType {
					continue
				}

				if !mustAccAddressFromBech32(op.Account.Address).Equals(rt.Account) {
					continue
				}
				symbol := op.Amount.Currency.Symbol
				value, err := types.AmountValue(op.Amount)
				if value.Sign() != 1 {
					continue
				}
				require.NoError(t, err)

				denom, ok := Denoms[symbol]
				require.True(t, ok)

				opCoins = opCoins.Add(sdk.NewCoin(denom, sdk.NewIntFromBigInt(value)))
			}
			assert.True(t, opCoins.IsEqual(rt.Balance))
		}

	})

	t.Run("each sender op has no related ops", func(t *testing.T) {

		for _, op := range ops {
			value, err := types.AmountValue(op.Amount)
			require.NoError(t, err)
			if value.Sign() == -1 {
				assert.Equal(t, 0, len(op.RelatedOperations))
			}
		}
	})

	t.Run("each transfer recipient op is related to a sender op", func(t *testing.T) {

		for _, op := range ops {
			if op.Type != TransferOpType {
				continue
			}

			value, err := types.AmountValue(op.Amount)
			require.NoError(t, err)
			if value.Sign() == 1 {
				if len(op.RelatedOperations) != 1 {
					continue
				}
				relatedOpIndex := op.RelatedOperations[0].Index
				relatedOp := ops[relatedOpIndex-ops[0].OperationIdentifier.Index]

				// index matches as expected
				assert.Equal(t, relatedOpIndex, relatedOp.OperationIdentifier.Index)
				// values match
				negatedRelValue, err := types.NegateValue(relatedOp.Amount.Value)
				require.NoError(t, err)
				assert.Equal(t, op.Amount.Value, negatedRelValue)

				// currencies match
				assert.Equal(t, op.Amount.Value, negatedRelValue)
			}
		}
	})
}

func calculateCoins(log sdk.ABCIMessageLog) sdk.Coins {
	coins := sdk.NewCoins()
	for _, ev := range log.Events {
		if ev.Type == banktypes.EventTypeTransfer {
			var amount sdk.Coins
			for _, attr := range ev.Attributes {
				if attr.Key == "amount" {
					amount = mustParseCoinsNormalized(attr.Value)
				}
				coins = coins.Add(amount...)
			}
		}
		if ev.Type == "delegate" {
			for _, attr := range ev.Attributes {
				if attr.Key == "amount" {
					coins = coins.Add(sdk.NewCoin("ukava", mustNewIntFromStr(attr.Value)))
				}
			}
		}
	}
	return coins
}

func readABCILogFromFile(t *testing.T, file string) sdk.ABCIMessageLog {
	txResponse := sdk.TxResponse{}
	bz, err := ioutil.ReadFile(filepath.Join("test-fixtures", file))
	if err != nil {
		t.Fatalf("could not read %s: %v", file, err)
	}
	cdc := app.MakeEncodingConfig().Amino
	cdc.MustUnmarshalJSON(bz, &txResponse)
	if len(txResponse.Logs) != 1 {
		t.Fatalf("each transaction should have one log, found %d for %s", len(txResponse.Logs), file)
	}
	return txResponse.Logs[0]
}

// TODO: fix to return real message
func readMsgFromFile(t *testing.T, file string) sdk.Msg {
	txResponse := sdk.TxResponse{}
	bz, err := ioutil.ReadFile(filepath.Join("test-fixtures", file))
	if err != nil {
		t.Fatalf("could not read %s: %v", file, err)
	}
	cdc := app.MakeEncodingConfig().Amino
	cdc.MustUnmarshalJSON(bz, &txResponse)
	//if len(txResponse.Tx.GetMsgs()) != 1 {
	//t.Fatalf("each transaction should have one msg, found %d for %s", len(txResponse.Tx.GetMsgs()), file)
	//}
	//return txResponse.Tx.GetMsgs()[0]
	return &banktypes.MsgSend{}
}

type accountBalance struct {
	Account sdk.AccAddress
	Balance sdk.Coins
}

func (ab accountBalance) String() string {
	return fmt.Sprintf(`
	Account: %s
	Balance %s
	`, ab.Account, ab.Balance)
}

func calculateSendersReceivers(msg sdk.Msg, log sdk.ABCIMessageLog) (senders, receivers []accountBalance) {
	senderMap := make(map[string]sdk.Coins)
	receiverMap := make(map[string]sdk.Coins)

	var sender sdk.AccAddress
	numTransferAttributes := 3

	if _, ok := msg.(*banktypes.MsgMultiSend); ok {
		numTransferAttributes = 2
		for _, ev := range log.Events {
			if ev.Type == "message" {
				for _, attr := range ev.Attributes {
					if attr.Key == "sender" {
						sender = mustAccAddressFromBech32(attr.Value)
					}
				}
			}
		}
	}

	for _, ev := range log.Events {
		if ev.Type == banktypes.EventTypeTransfer {
			unflattenedTransferEvents := unflattenEvents(ev, banktypes.EventTypeTransfer, numTransferAttributes)
			for _, event := range unflattenedTransferEvents {
				var recipient sdk.AccAddress
				var amount sdk.Coins
				for _, attr := range event.Attributes {

					if attr.Key == "sender" {
						sender = mustAccAddressFromBech32(attr.Value)
					}
					if attr.Key == "recipient" {
						recipient = mustAccAddressFromBech32(attr.Value)
					}
					if attr.Key == "amount" {
						amount = mustParseCoinsNormalized(attr.Value)
					}
				}
				filteredCoins := filterCoins(amount)
				if filteredCoins.Empty() {
					continue
				}
				senderCoins, seenSender := senderMap[sender.String()]
				if !seenSender {
					senderCoins = amount
				} else {
					senderCoins = senderCoins.Add(amount...)
				}
				senderMap[sender.String()] = senderCoins
				receiverCoins, seenReceiver := receiverMap[recipient.String()]
				if !seenReceiver {
					receiverCoins = amount
				} else {
					receiverCoins = receiverCoins.Add(amount...)
				}
				receiverMap[recipient.String()] = receiverCoins
			}
		}
	}

	for sender, balance := range senderMap {
		senders = append(senders, accountBalance{Account: mustAccAddressFromBech32(sender), Balance: balance})
	}
	for receiver, balance := range receiverMap {
		receivers = append(receivers, accountBalance{Account: mustAccAddressFromBech32(receiver), Balance: balance})
	}
	switch msg.(type) {
	case *stakingtypes.MsgDelegate:
		senders, receivers = calcDelegationSendersReceivers(senders, receivers, log)
	case *stakingtypes.MsgCreateValidator:
		senders, receivers = calcCreateValdiatorSendersReceivers(senders, receivers, log)
	}
	return senders, receivers
}

func calcDelegationSendersReceivers(senders, receivers []accountBalance, log sdk.ABCIMessageLog) (delegationSenders, delegationReceivers []accountBalance) {
	var amount sdk.Coin
	var sender sdk.AccAddress
	recipient := stakingModuleAddress
	for _, ev := range log.Events {
		if ev.Type == "delegate" {
			for _, attr := range ev.Attributes {
				if attr.Key == "amount" {
					amount = sdk.NewCoin("ukava", mustNewIntFromStr(attr.Value))
				}
			}
		} else if ev.Type == "message" {
			for _, attr := range ev.Attributes {
				if attr.Key == "sender" && !mustAccAddressFromBech32(attr.Value).Equals(recipient) {
					sender = mustAccAddressFromBech32(attr.Value)
				}
			}
		}
	}
	delegationSenders = append(senders, accountBalance{Account: sender, Balance: sdk.NewCoins(amount)})
	delegationReceivers = append(receivers, accountBalance{Account: recipient, Balance: sdk.NewCoins(amount)})
	return delegationSenders, delegationReceivers
}

func calcCreateValdiatorSendersReceivers(senders, receivers []accountBalance, log sdk.ABCIMessageLog) (delegationSenders, delegationReceivers []accountBalance) {
	var amount sdk.Coin
	var sender sdk.AccAddress
	recipient := stakingModuleAddress
	for _, ev := range log.Events {
		if ev.Type == "create_validator" {
			for _, attr := range ev.Attributes {
				if attr.Key == "amount" {
					amount = sdk.NewCoin("ukava", mustNewIntFromStr(attr.Value))
				}
			}
		} else if ev.Type == "message" {
			for _, attr := range ev.Attributes {
				if attr.Key == "sender" && !mustAccAddressFromBech32(attr.Value).Equals(recipient) {
					sender = mustAccAddressFromBech32(attr.Value)
				}
			}
		}
	}
	delegationSenders = append(senders, accountBalance{Account: sender, Balance: sdk.NewCoins(amount)})
	delegationReceivers = append(receivers, accountBalance{Account: recipient, Balance: sdk.NewCoins(amount)})
	return delegationSenders, delegationReceivers
}

func filterCoins(amount sdk.Coins) sdk.Coins {
	filtered := sdk.NewCoins()
	for _, c := range amount {
		_, ok := Currencies[c.Denom]
		if ok {
			filtered = filtered.Add(c)
		}
	}
	return filtered
}
