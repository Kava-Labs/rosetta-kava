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

// Package configuration loads service configuration from the environment
package configuration

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kava-labs/rosetta-kava/kava"

	"github.com/coinbase/rosetta-sdk-go/types"
)

// MiddlewareVersion represents the kava rosetta service version
var MiddlewareVersion = "0.0.1"

// Mode identifies if the service is running in an 'online' or 'offline'
// capacity.  Ref: https://www.rosetta-api.org/docs/node_deployment.html#multiple-modes
type Mode string

// String converts mode to string
func (m Mode) String() string {
	return string(m)
}

const (
	// Online specifies that outbound connections are permitted.
	Online Mode = "online"

	// Offline specifies that outbound connections are not permitted.
	Offline Mode = "offline"

	// ModeEnv specifies the environment variable read to set the Mode
	ModeEnv = "MODE"

	// NetworkEnv specifies the environment variable to read Network/ChainId from
	NetworkEnv = "NETWORK"

	// PortEnv specifies the environment variable to read server port from
	PortEnv = "PORT"
)

// ModeFromString returns a Mode from a string value
func ModeFromString(val string) (m Mode, err error) {
	switch val {
	case Online.String():
		m = Online
	case Offline.String():
		m = Offline
	default:
		err = fmt.Errorf("invalid mode %s, must be one of [%s,%s]", val, Online, Offline)
	}

	return
}

// ConfigLoader provides an interface for loading values from a string key
type ConfigLoader interface {
	Get(key string) string
}

// EnvLoader loads keys from os environment
// and implements ConfigLoader
type EnvLoader struct {
}

var _ ConfigLoader = (*EnvLoader)(nil)

// Get retrieves key from os environment
func (l *EnvLoader) Get(key string) string {
	return os.Getenv(key)
}

// Configuration represents values to configure behavior of
// rosetta-kava and network to communicate with.
type Configuration struct {
	Mode              Mode
	NetworkIdentifier *types.NetworkIdentifier
	Port              int
}

// LoadConfig loads keys from a provided loader and returns a
// complete Configuration reference for rosetta-kava
func LoadConfig(loader ConfigLoader) (*Configuration, error) {
	modeEnv := loader.Get(ModeEnv)

	if modeEnv == "" {
		return nil, fmt.Errorf("%s must be set", ModeEnv)
	}

	mode, err := ModeFromString(modeEnv)
	if err != nil {
		return nil, err
	}

	networkEnv := loader.Get(NetworkEnv)

	if networkEnv == "" {
		return nil, fmt.Errorf("%s must be set", NetworkEnv)
	}

	networkIdentifier := &types.NetworkIdentifier{
		Blockchain: kava.Blockchain,
		Network:    networkEnv,
	}

	portEnv := loader.Get(PortEnv)

	if portEnv == "" {
		return nil, fmt.Errorf("%s must be set", PortEnv)
	}

	portNum, err := strconv.Atoi(portEnv)
	if err != nil || portNum <= 0 {
		return nil, fmt.Errorf("invalid port '%s'", portEnv)
	}

	return &Configuration{
		Mode:              mode,
		NetworkIdentifier: networkIdentifier,
		Port:              portNum,
	}, nil
}
