/*
Copyright © 2021 Sebastien Leger

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"github.com/pkg/errors"
	"github.com/regel/cardano-p2p/log"
	"github.com/spf13/viper"
	"time"
)

var (
	defaultListenAddr      = ":8080"
	defaultTestnetEndpoint = "localhost:6001"
	defaultMainnetEndpoint = "localhost:6002"
	testnetMagic           = uint64(1097911063)
	mainnetMagic           = uint64(764824073)
	defaultMaximumPeers    = 10
	defaultPeriodSeconds   = 60 * time.Second
	defaultFetchMaximum    = 2000
	defaultReadTimeout     = 1 * time.Second
	defaultTestnetPeer     = "relays-new.cardano-testnet.iohkdev.io:3001"
	defaultMainnetPeer     = "relays-new.cardano-mainnet.iohk.io:3001"
)

type EndpointConfig struct {
	Endpoint      string        `mapstructure:"endpoint,omitempty"`
	PeriodSeconds time.Duration `mapstructure:"period-seconds,omitempty"`
	FetchMaximum  int           `mapstructure:"fetch-maximum,omitempty"`
	Enabled       bool          `mapstructure:"enabled,omitempty"`
	MaxPeers      int           `mapstructure:"max-peers,omitempty"`
	NetworkMagic  uint64        `mapstructure:"magic,omitempty"`
	DefaultPeer   string        `mapstructure:"default-peer,omitempty"`
}

type Config struct {
	Debug         bool           `mapstructure:"debug,omitempty"`
	ListenAddress string         `mapstructure:"listen-addr,omitempty"`
	ReadTimeout   time.Duration  `mapstructure:"read-timeout,omitempty"`
	Testnet       EndpointConfig `mapstructure:"testnet,omitempty"`
	Mainnet       EndpointConfig `mapstructure:"mainnet,omitempty"`
}

// DefaultConfig returns a config with defaults set
func DefaultConfig() *Config {
	return &Config{
		Debug:         false,
		ListenAddress: defaultListenAddr,
		ReadTimeout:   defaultReadTimeout,
		Testnet: EndpointConfig{
			Enabled:       true,
			PeriodSeconds: defaultPeriodSeconds,
			FetchMaximum:  defaultFetchMaximum,
			NetworkMagic:  testnetMagic,
			Endpoint:      defaultTestnetEndpoint,
			MaxPeers:      defaultMaximumPeers,
			DefaultPeer:   defaultTestnetPeer,
		},
		Mainnet: EndpointConfig{
			Enabled:       true,
			PeriodSeconds: defaultPeriodSeconds,
			FetchMaximum:  defaultFetchMaximum,
			NetworkMagic:  mainnetMagic,
			Endpoint:      defaultMainnetEndpoint,
			MaxPeers:      defaultMaximumPeers,
			DefaultPeer:   defaultMainnetPeer,
		},
	}
}

// Load config
func (c *Config) Load(configFile string) error {
	log.Debugf("reading config from: %s", configFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			return errors.Wrap(err, "config file not found")
		} else {
			// Config file was found but another error was produced
			return errors.Wrap(err, "could not read config file")
		}
	}
	if err := viper.Unmarshal(c); err != nil {
		return errors.Wrap(err, "bad config file format")
	}
	return c.Validate()
}

// Validate validates the config
func (c *Config) Validate() error {
	return nil
}
