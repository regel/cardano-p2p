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
package cmd

import (
	"os"

	"encoding/json"
	"github.com/regel/cardano-p2p/log"
	"github.com/regel/cardano-p2p/pkg"
	"github.com/regel/cardano-p2p/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	testnetChan chan pkg.Producer
	mainnetChan chan pkg.Producer
)

var p2pCmd = &cobra.Command{
	Use:   "p2p",
	Short: "Run p2p service",
	Long: `Runs an http service to share secure and up to date node information with peers.

The p2p command runs an http service. The API is similar to api.clio.one to ease transitions
to the p2p service`,
	Run: p2p,
}

func init() {
	rootCmd.AddCommand(p2pCmd)
}

func p2p(cmd *cobra.Command, args []string) {
	configFile := viper.ConfigFileUsed()
	config := server.DefaultConfig()
	if err := config.Load(configFile); err != nil {
		log.Errorf("Unable to load config: %s:\n%v", configFile, err)
		os.Exit(1)
	}
	b, _ := json.Marshal(config)
	log.Debugf("Config: \n%v", string(b))
	testnetChan = make(chan pkg.Producer, config.Testnet.FetchMaximum)
	mainnetChan = make(chan pkg.Producer, config.Mainnet.FetchMaximum)
	if config.Testnet.Enabled {
		go pkg.Push(&config.Testnet, testnetChan)
	}
	if config.Mainnet.Enabled {
		go pkg.Push(&config.Mainnet, mainnetChan)
	}
	pkg.Serve(config, testnetChan, mainnetChan)
	select {} // infinite loop
}
