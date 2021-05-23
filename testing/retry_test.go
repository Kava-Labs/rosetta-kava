// +build integration
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

package testing

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
)

func TestBlockRetry(t *testing.T) {
	if config.Mode.String() == "offline" {
		t.Skip("offline: skipping block retry test")
	}

	numJobs := 10
	jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type serviceError struct {
		rerr *types.Error
		err  error
	}
	errChan := make(chan serviceError)

	for i := 0; i < numJobs; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			for {
				select {
				case <-jobCtx.Done():
					return
				default:
				}

				ctx := context.Background()
				networkStatus, rosettaErr, err := client.NetworkAPI.NetworkStatus(
					ctx,
					&types.NetworkRequest{
						NetworkIdentifier: config.NetworkIdentifier,
					},
				)

				currentBlock := networkStatus.CurrentBlockIdentifier
				request := &types.BlockRequest{
					NetworkIdentifier: config.NetworkIdentifier,
					BlockIdentifier: &types.PartialBlockIdentifier{
						Index: &currentBlock.Index,
					},
				}

				_, rosettaErr, err = client.BlockAPI.Block(ctx, request)
				if rosettaErr != nil || err != nil {
					errChan <- serviceError{rosettaErr, err}
				}
			}
		}()
	}

	select {
	case blockErr := <-errChan:
		if blockErr.err != nil {
			t.Fatalf("received error fetching block %s", blockErr.err)
		}

		if blockErr.rerr != nil {
			t.Fatalf("received rosetta error fetching block %s", blockErr.rerr.Message)
		}
	case <-jobCtx.Done():
	}
}
