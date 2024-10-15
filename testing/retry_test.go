//go:build integration
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
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
)

var lessOrEqualCurrentHeight = regexp.MustCompile(`height \d+ must be less than or equal to the current blockchain height`)

func TestBlockRetry(t *testing.T) {
	if config.Mode.String() == "offline" {
		t.Skip("offline: skipping block retry test")
	}

	if os.Getenv("SKIP_LIVE_NODE_TESTS") == "true" {
		t.Skip("skipping block retry test: it's designed to be run against a live (mainnet) node")
	}

	if os.Getenv("SKIP_RESOURCE_INTENSIVE_TESTS") == "true" {
		t.Skip("skipping block retry test: it's resource intensive and produces a lot of requests to the node")
	}

	numJobs := 10
	jobCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type serviceError struct {
		rerr *types.Error
		err  error
	}
	errChan := make(chan serviceError, numJobs)

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

				if err != nil {
					if lessOrEqualCurrentHeight.MatchString(err.Error()) {
						continue
					}
				}

				if rosettaErr != nil || err != nil {
					errChan <- serviceError{rosettaErr, err}
					return
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
