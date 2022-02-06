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
	"bytes"
	"context"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"os"

	"encoding/json"
	"github.com/regel/cardano-p2p/log"
	"github.com/regel/cardano-p2p/pkg"
	"github.com/regel/cardano-p2p/server"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Connects to api.clio.one or similar service to push our Cardano ledger tip..",
	Long: `
This command is provided for backward compatibility service with other Cardano nodes.
It ensures the CLIO service receives up to date blockNo and adds our node to a list
of active Cardano nodes.`,
	Run: push,
}

func newPushCmd() *cobra.Command {
	cmd := pushCmd
	flags := cmd.Flags()
	addPushFlags(flags)
	return cmd
}

func addPushFlags(flags *flag.FlagSet) {
	flags.String("endpoint-url", defaultEndpoint, heredoc.Doc(`
The http(s) address used to get a list of Cardano nodes`))
	flags.Int64("network", testnetMagic, heredoc.Doc(`
Unique network magic of the Cardano blockchain, eg. 1097911063 for testnet`))
	flags.Int64("port", defaultNodePort, heredoc.Doc(`
Public port number of the Cardano node`))
}

func init() {
	rootCmd.AddCommand(newPushCmd())
}

func push(cmd *cobra.Command, args []string) {
	configFile := viper.ConfigFileUsed()
	config := server.DefaultConfig()
	if err := config.Load(configFile); err != nil {
		log.Errorf("Unable to load config: %s:\n%v", configFile, err)
		os.Exit(1)
	}
	b, _ := json.Marshal(config)
	log.Debugf("Config: \n%v", string(b))
	context := context.Background()

	endpointUrl, _ := cmd.Flags().GetString("endpoint-url")
	magic, _ := cmd.Flags().GetInt64("network")
	port, _ := cmd.Flags().GetInt64("port")

	blockNo, err := pkg.GetBlockHeight(config.Client.Endpoint)
	if err != nil {
		log.Errorf("Cannot get blockNo: %v", err)
		os.Exit(1)
	}
	src, err := pkg.PushBlockNo(context, endpointUrl, magic, port, *blockNo)
	if err != nil {
		log.Errorf("Cannot push ledger data: %v", err)
		os.Exit(1)
	}
	dst := &bytes.Buffer{}
	if err := json.Indent(dst, src, "", "  "); err != nil {
		panic(err)
	}
	fmt.Println(dst.String())
}
