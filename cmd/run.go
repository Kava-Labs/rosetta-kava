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

package cmd

import (
	"fmt"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/server"

	"github.com/spf13/cobra"
)

var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run rosetta-kava",
		RunE:  runRunCmd,
	}
)

func runRunCmd(cmd *cobra.Command, args []string) error {
	configLoader := &configuration.EnvLoader{}

	config, err := configuration.LoadConfig(configLoader)
	if err != nil {
		return fmt.Errorf("%w: unable to load configuration", err)
	}

	handler, err := server.NewRouter(config)
	if err != nil {
		return fmt.Errorf("%w: unable to initialize router", err)
	}

	return server.Run(config, handler)
}
