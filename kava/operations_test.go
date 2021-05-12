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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"

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
		"ukava", "hard", "usdx",
		// not native
		"bnb", "busd", "btcb",
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
				assert.Equal(t, opType, op.Type)

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
		Type: bank.EventTypeTransfer,
		Attributes: []sdk.Attribute{
			{
				Key:   bank.AttributeKeyRecipient,
				Value: testAddresses[1],
			},
			{
				Key:   bank.AttributeKeySender,
				Value: testAddresses[0],
			},
			{
				Key:   sdk.AttributeKeyAmount,
				Value: generateDefaultCoins().String(),
			},
		},
	}

	testEvent2 := sdk.StringEvent{
		Type: bank.EventTypeTransfer,
		Attributes: []sdk.Attribute{
			{
				Key:   bank.AttributeKeyRecipient,
				Value: testAddresses[2],
			},
			{
				Key:   bank.AttributeKeySender,
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
	ops := EventsToOperations(events, index)

	assert.Greater(t, len(ops), 0)
	for opIndex, op := range ops {
		assert.Equal(t, int64(opIndex)+index, op.OperationIdentifier.Index)
		assert.Equal(t, SuccessStatus, *op.Status)
	}

	index = int64(10)
	events = sdk.StringEvents{testEvent1, testEvent2}
	ops = EventsToOperations(events, index)

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
					Type: bank.EventTypeTransfer,
					Attributes: []sdk.Attribute{
						{
							Key:   bank.AttributeKeyRecipient,
							Value: testAddresses[0],
						},
						{
							Key:   bank.AttributeKeySender,
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
					Type: bank.EventTypeTransfer,
					Attributes: []sdk.Attribute{
						{
							Key:   bank.AttributeKeyRecipient,
							Value: testAddresses[1],
						},
						{
							Key:   bank.AttributeKeySender,
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
	msg1 := bank.MsgSend{
		FromAddress: getAccAddr(t, testAddresses[0]),
		ToAddress:   getAccAddr(t, testAddresses[1]),
		Amount:      generateDefaultCoins(),
	}

	msg2 := bank.MsgSend{
		FromAddress: getAccAddr(t, testAddresses[0]),
		ToAddress:   getAccAddr(t, testAddresses[1]),
		Amount:      generateDefaultCoins(),
	}

	// one less than message length
	logs := sdk.ABCIMessageLogs{
		sdk.ABCIMessageLog{},
	}

	success := SuccessStatus
	failure := FailureStatus

	t.Run("no fee", func(t *testing.T) {
		tx := authtypes.StdTx{
			Msgs: []sdk.Msg{msg1, msg2},
			Fee:  authtypes.StdFee{Gas: 500000},
		}

		// all ops succesful and indexed correctly
		ops := TxToOperations(&tx, logs, &success)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, success, *op.Status)
		}

		// all ops failed and indexed correctly
		ops = TxToOperations(&tx, logs, &failure)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, failure, *op.Status)
		}

		// there are no fee operations
		ops = TxToOperations(&tx, logs, &success)
		for _, op := range ops {
			assert.NotEqual(t, FeeOpType, op.Type)
		}
	})

	t.Run("with fee", func(t *testing.T) {
		tx := authtypes.StdTx{
			Msgs: []sdk.Msg{msg1, msg2},
			Fee:  authtypes.StdFee{Amount: generateCoins([]string{"ukava"}), Gas: 500000},
		}

		// all ops succesful and indexed correctly
		ops := TxToOperations(&tx, logs, &success)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)
			assert.Equal(t, success, *op.Status)
		}

		// all ops failed and indexed correctly
		ops = TxToOperations(&tx, logs, &failure)
		for index, op := range ops {
			assert.Equal(t, int64(index), op.OperationIdentifier.Index)

			// fee operations are always successful
			if op.Type == FeeOpType {
				assert.Equal(t, success, *op.Status)
			} else {
				assert.Equal(t, failure, *op.Status)
			}
		}

		// there are fee operations
		feeOpTypeFound := false
		ops = TxToOperations(&tx, logs, &success)
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
		name   string
		logs   []sdk.ABCIMessageLog
		status string
	}{
		{
			name:   "hard.MsgDeposit",
			logs:   readABCILogFromFile("hard-deposit-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "hard.MsgWithdraw",
			logs:   readABCILogFromFile("hard-withdraw-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "hard.MsgBorrow",
			logs:   readABCILogFromFile("hard-borrow-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "hard.MsgRepay",
			logs:   readABCILogFromFile("hard-repay-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "hard.MsgLiquidate",
			logs:   readABCILogFromFile("hard-liquidate-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "auction.MsgPlaceBid",
			logs:   readABCILogFromFile("auction-bid-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "bep3.MsgCreateAtomicSwap",
			logs:   readABCILogFromFile("bep3-create-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "bep3.MsgRefundAtomicSwap",
			logs:   readABCILogFromFile("bep3-refund-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "bep3.MsgClaimAtomicSwap",
			logs:   readABCILogFromFile("bep3-claim-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgCreateCDP",
			logs:   readABCILogFromFile("cdp-create-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgDeposit",
			logs:   readABCILogFromFile("cdp-deposit-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgWithdraw",
			logs:   readABCILogFromFile("cdp-withdraw-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgDrawDebt",
			logs:   readABCILogFromFile("cdp-draw-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgRepayDebt",
			logs:   readABCILogFromFile("cdp-repay-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cdp.MsgLiquidate",
			logs:   readABCILogFromFile("cdp-liquidate-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "kava.SubmitProposal",
			logs:   readABCILogFromFile("committee-submit-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "kava.MsgVote",
			logs:   readABCILogFromFile("committee-vote-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "incentive.MsgClaimUSDXMintingReward",
			logs:   readABCILogFromFile("incentive-claim-usdx-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "incentive.MsgClaimHardReward",
			logs:   readABCILogFromFile("incentive-claim-hard-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "pricefeed.MsgPostPrice",
			logs:   readABCILogFromFile("pricefeed-post-tx-response.json"),
			status: SuccessStatus,
		},
		{
			name:   "cosmos-sdk.MsgSend",
			logs:   readABCILogFromFile("msg-send-tx-response.json"),
			status: SuccessStatus,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for _, log := range tc.logs {
				ops := getTransferOpsFromMsg(log, &tc.status, 0)
				assertTransferOpsBalanceTrack(t, tc.name, ops)
			}
		})
	}
}

func assertTransferOpsBalanceTrack(
	t *testing.T,
	name string,
	ops []*types.Operation,
) {
	amount := calculateCoins(ops)
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
	})

	senderTracking, receiverTracking := calculateSendersReceivers(ops)

	t.Run("coin operations sum to zero", func(t *testing.T) {
		if len(senderTracking) == 0 || len(receiverTracking) == 0 {
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
		for _, st := range senderTracking {
			if st.Account == nil {
				t.Skip("no sender")
			}
			opCoins := sdk.NewCoins()
			for _, op := range ops {
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

	t.Run("coin operation amounts match for recipient", func(t *testing.T) {
		for _, rt := range receiverTracking {
			if rt.Account == nil {
				t.Skip("no recipient")
			}
			opCoins := sdk.NewCoins()
			for _, op := range ops {
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

	t.Run("each recipient op is related to a sender op", func(t *testing.T) {

		for _, op := range ops {
			value, err := types.AmountValue(op.Amount)
			require.NoError(t, err)
			if value.Sign() == 1 {
				assert.Equal(t, 1, len(op.RelatedOperations))
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

func calculateCoins(ops []*types.Operation) sdk.Coins {
	coins := sdk.NewCoins()
	for _, op := range ops {
		coinAmount, ok := sdk.NewIntFromString(op.Amount.Value)
		if !ok {
			panic(fmt.Sprintf("invalid input amount: %s\n", op.Amount.Value))
		}
		if coinAmount.IsPositive() {
			coins = coins.Add(sdk.NewCoin(Denoms[op.Amount.Currency.Symbol], coinAmount))
		}
	}
	return coins
}

func readABCILogFromFile(file string) sdk.ABCIMessageLogs {
	txResponse := sdk.TxResponse{}
	bz, err := ioutil.ReadFile(filepath.Join("mocks", file))
	if err != nil {
		panic(err)
	}
	cdc := app.MakeCodec()
	cdc.MustUnmarshalJSON(bz, &txResponse)
	return txResponse.Logs

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

func calculateSendersReceivers(ops []*types.Operation) (senders, receivers []accountBalance) {
	senderMap := make(map[string]sdk.Coins)
	receiverMap := make(map[string]sdk.Coins)
	for _, op := range ops {
		// if len(op.RelatedOperations) > 0 {
		// 	fmt.Printf("receiver operation: {Identifier: %d, Type: %s, Status: %s, Address: %s, Amount: %s, Currency: %s, Realated Ops: [%d]}\n", op.OperationIdentifier.Index, op.Type, *op.Status, op.Account.Address, op.Amount.Value, op.Amount.Currency.Symbol, op.RelatedOperations[0].Index)
		// } else {
		// 	fmt.Printf("sender operation: {Identifier: %d, Type: %s, Status: %s, Address: %s, Amount: %s, Currency: %s}\n", op.OperationIdentifier.Index, op.Type, *op.Status, op.Account.Address, op.Amount.Value, op.Amount.Currency.Symbol)
		// }
		coinAmount, ok := sdk.NewIntFromString(op.Amount.Value)
		if !ok {
			panic(fmt.Sprintf("invalid input amount: %s\n", op.Amount.Value))
		}
		if coinAmount.IsNegative() {
			sender := mustAccAddressFromBech32(op.Account.Address)
			coins, seen := senderMap[sender.String()]
			if !seen {
				coins = sdk.NewCoins(sdk.NewCoin(Denoms[op.Amount.Currency.Symbol], coinAmount.Neg()))
			} else {
				coins = coins.Add(sdk.NewCoin(Denoms[op.Amount.Currency.Symbol], coinAmount.Neg()))
			}
			senderMap[sender.String()] = coins
		} else {
			receiver := mustAccAddressFromBech32(op.Account.Address)
			coins, seen := receiverMap[receiver.String()]
			if !seen {
				coins = sdk.NewCoins(sdk.NewCoin(Denoms[op.Amount.Currency.Symbol], coinAmount))
			} else {
				coins = coins.Add(sdk.NewCoin(Denoms[op.Amount.Currency.Symbol], coinAmount))
			}
			receiverMap[receiver.String()] = coins
		}
	}
	for sender, balance := range senderMap {
		senders = append(senders, accountBalance{Account: mustAccAddressFromBech32(sender), Balance: balance})
	}
	for receiver, balance := range receiverMap {
		receivers = append(receivers, accountBalance{Account: mustAccAddressFromBech32(receiver), Balance: balance})
	}
	return senders, receivers
}
