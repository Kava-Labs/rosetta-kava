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

package configuration

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
)

type testEnvLoader struct {
	Env map[string]string
}

func (l *testEnvLoader) Get(key string) string {
	value, ok := l.Env[key]

	if !ok {
		return ""
	}

	return value
}

func TestLoadConfig_Mode(t *testing.T) {
	blockchain := "Kava"
	testChainID := "kava-testnet-9999"
	testPort := "8001"
	testPortNum, err := strconv.Atoi(testPort)
	testKavaRpcUrl := "https://rpc.testnet.kava.io:443"
	assert.NoError(t, err)

	tests := map[string]struct {
		Env            map[string]string
		ExpectedConfig *Configuration
		ExpectedErr    error
	}{
		"no env vars set": {
			Env:         make(map[string]string),
			ExpectedErr: fmt.Errorf("%s must be set", ModeEnv),
		},
		"network not set": {
			Env: map[string]string{
				ModeEnv: Online.String(),
			},
			ExpectedErr: fmt.Errorf("%s must be set", NetworkEnv),
		},
		"port not set": {
			Env: map[string]string{
				ModeEnv:    Online.String(),
				NetworkEnv: testChainID,
			},
			ExpectedErr: fmt.Errorf("%s must be set", PortEnv),
		},
		"invalid mode set": {
			Env: map[string]string{
				ModeEnv: "sync",
			},
			ExpectedErr: fmt.Errorf("invalid mode sync, must be one of [%s,%s]", Online, Offline),
		},
		"invalid port set - not a number": {
			Env: map[string]string{
				ModeEnv:    Offline.String(),
				NetworkEnv: testChainID,
				PortEnv:    "invalid number",
			},
			ExpectedErr: fmt.Errorf("invalid port 'invalid number'"),
		},
		"invalid port set - negative": {
			Env: map[string]string{
				ModeEnv:    Offline.String(),
				NetworkEnv: testChainID,
				PortEnv:    "-8000",
			},
			ExpectedErr: fmt.Errorf("invalid port '-8000'"),
		},
		"kavaRpcUrl not set": {
			Env: map[string]string{
				ModeEnv:    Offline.String(),
				NetworkEnv: testChainID,
				PortEnv:    "8000",
			},
			ExpectedErr: fmt.Errorf("%s must be set", KavaRpcUrlEnv),
		},
		"env set with online mode": {
			Env: map[string]string{
				ModeEnv:       Online.String(),
				NetworkEnv:    testChainID,
				PortEnv:       testPort,
				KavaRpcUrlEnv: testKavaRpcUrl,
			},
			ExpectedConfig: &Configuration{
				Mode: Online,
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: blockchain,
					Network:    testChainID,
				},
				Port:       testPortNum,
				KavaRpcUrl: testKavaRpcUrl,
			},
		},
		"env set with offline mode": {
			Env: map[string]string{
				ModeEnv:       Offline.String(),
				NetworkEnv:    testChainID,
				PortEnv:       testPort,
				KavaRpcUrlEnv: testKavaRpcUrl,
			},
			ExpectedConfig: &Configuration{
				Mode: Offline,
				NetworkIdentifier: &types.NetworkIdentifier{
					Blockchain: blockchain,
					Network:    testChainID,
				},
				Port:       testPortNum,
				KavaRpcUrl: testKavaRpcUrl,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			loader := &testEnvLoader{Env: tc.Env}
			cfg, err := LoadConfig(loader)

			if tc.ExpectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedConfig, cfg)
			} else {
				assert.Nil(t, cfg)
				assert.Equal(t, tc.ExpectedErr, err)
			}
		})
	}
}

func TestConfigLoader(t *testing.T) {
	testVarName := "ROSETTA_KAVA_TEST_VAR"
	testVarVal := "a test value"

	err := os.Setenv(testVarName, testVarVal)
	assert.NoError(t, err)
	loader := &EnvLoader{}
	assert.Equal(t, loader.Get(testVarName), testVarVal)
}
