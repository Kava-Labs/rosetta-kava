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
	"context"
	"errors"

	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	kava "github.com/kava-labs/kava/app"
)

func init() {
	//
	// bootstrap kava chain config
	//
	// sets a global cosmos sdk for bech32 prefix
	//
	// required before loading config
	//
	kavaConfig := sdk.GetConfig()
	kava.SetBech32AddressPrefixes(kavaConfig)
	kava.SetBip44CoinType(kavaConfig)
	kavaConfig.Seal()
}

// Client implements services.Client interface for communicating with the kava chain
type Client struct {
}

// NewClient initialized a new Kava Client
func NewClient() (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Status(ctx context.Context) (
	*types.BlockIdentifier,
	int64,
	*types.BlockIdentifier,
	*types.SyncStatus,
	[]*types.Peer,
	error,
) {
	return nil, int64(-1), nil, nil, nil, errors.New("not implemented")
}
