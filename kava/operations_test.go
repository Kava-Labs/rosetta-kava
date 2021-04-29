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
	"math/big"
	"math/rand"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			for _, op := range ops {
				if op.Account != sender {
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
			for _, op := range ops {
				if op.Account != recipient {
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
				assert.Contains(t, []*types.AccountIdentifier{sender, recipient}, op.Account)
			}
		})

		t.Run("each sender op has no related ops", func(t *testing.T) {
			for _, op := range ops {
				if op.Account != sender {
					continue
				}

				assert.Equal(t, 0, len(op.RelatedOperations))
			}
		})

		t.Run("each recipient op is related to a sender op", func(t *testing.T) {
			for _, op := range ops {
				if op.Account != recipient {
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
	assert.Equal(t, []*types.Operation{}, EventsToOperations(sdk.StringEvents{}, 0))
}

func TestTxToOperations(t *testing.T) {
	msg1 := bank.MsgSend{
		ToAddress:   getAccAddr(t, testAddresses[0]),
		FromAddress: getAccAddr(t, testAddresses[1]),
		Amount:      generateDefaultCoins(),
	}

	msg2 := bank.MsgSend{
		ToAddress:   getAccAddr(t, testAddresses[0]),
		FromAddress: getAccAddr(t, testAddresses[1]),
		Amount:      generateDefaultCoins(),
	}

	logs := sdk.ABCIMessageLogs{
		sdk.ABCIMessageLog{},
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
		name      string
		createFn  func(coins sdk.Coins) (sdk.Msg, sdk.ABCIMessageLog)
		opType    string
		sender    *types.AccountIdentifier
		recipient *types.AccountIdentifier
	}{
		{
			name: "bank.MsgSend",
			createFn: func(coins sdk.Coins) (sdk.Msg, sdk.ABCIMessageLog) {
				return bank.MsgSend{
					FromAddress: getAccAddr(t, testAddresses[0]),
					ToAddress:   getAccAddr(t, testAddresses[1]),
					Amount:      coins,
				}, sdk.ABCIMessageLog{}
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
				msg, log := tc.createFn(otc.coins)
				ops := MsgToOperations(msg, log, &otc.status, otc.index)

				assertTrackedBalance(t, otc.name, ops, tc.sender, otc.coins, tc.recipient)

				return ops
			})
		})
	}
}