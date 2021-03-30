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
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAccountService_Offline(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	servicer := NewAccountAPIService(cfg, mockClient)
	ctx := context.Background()

	bal, err := servicer.AccountBalance(ctx, &types.AccountBalanceRequest{})
	assert.Nil(t, bal)
	assert.Equal(t, ErrUnavailableOffline.Code, err.Code)

	coins, err := servicer.AccountCoins(ctx, nil)
	assert.Nil(t, coins)
	assert.Equal(t, ErrUnimplemented.Code, err.Code)
	assert.Equal(t, ErrUnimplemented.Message, err.Message)

	mockClient.AssertExpectations(t)
}

func TestAccountService_Online(t *testing.T) {
	cfg := &configuration.Configuration{
		Mode: configuration.Online,
	}
	mockClient := &mocks.Client{}
	servicer := NewAccountAPIService(cfg, mockClient)

	ctx := context.Background()

	bal, err := servicer.AccountBalance(ctx, &types.AccountBalanceRequest{})
	assert.Nil(t, bal)
	assert.Equal(t, ErrUnimplemented.Code, err.Code)
	assert.Equal(t, ErrUnimplemented.Message, err.Message)

	coins, err := servicer.AccountCoins(ctx, nil)
	assert.Nil(t, coins)
	assert.Equal(t, ErrUnimplemented.Code, err.Code)
	assert.Equal(t, ErrUnimplemented.Message, err.Message)

	mockClient.AssertExpectations(t)
}
