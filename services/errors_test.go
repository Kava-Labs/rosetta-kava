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

package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapErr(t *testing.T) {
	for _, rosettaError := range Errors {
		t.Run(rosettaError.Message, func(t *testing.T) {
			originalError := errors.New("some internal error")
			wrappedErr := wrapErr(rosettaError, originalError)

			assert.Equal(t, rosettaError.Code, rosettaError.Code)
			assert.Equal(t, rosettaError.Message, rosettaError.Message)
			assert.Equal(t, rosettaError.Retriable, rosettaError.Retriable)

			errContext := wrappedErr.Details["context"]
			assert.Equal(t, originalError.Error(), errContext)
		})
	}
}
