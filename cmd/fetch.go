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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc"
	"github.com/go-redis/redis/v8"
	"github.com/regel/cardano-p2p/log"
	"github.com/regel/cardano-p2p/pkg"
	"github.com/regel/cardano-p2p/server"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"strings"
)

const (
	defaultFetchMax   = 10
	defaultIpVersion  = 4
	defaultRedisTopic = "p2p"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Connects to api.clio.one or similar service to fetch a list of cardano nodes.",
	Long: `
The fetch command connects to api.clio.one or another equivalent API to find a list of active
cardano nodes.`,
	Run: fetch,
}

func newFetchCmd() *cobra.Command {
	cmd := fetchCmd
	flags := cmd.Flags()
	addFetchFlags(flags)
	return cmd
}

func addFetchFlags(flags *flag.FlagSet) {
	flags.String("endpoint-url", defaultEndpoint, heredoc.Doc(`
The http(s) address used to get a list of Cardano nodes`))
	flags.Int64("network", testnetMagic, heredoc.Doc(`
Unique network magic of the Cardano blockchain, eg. 1097911063 for testnet`))
	flags.Int64("max", defaultFetchMax, heredoc.Doc(`
The maximum number of expected Cardano node addresses`))
	flags.Int64("ipv", defaultIpVersion, heredoc.Doc(`
The IP protocol version of expected Cardano nodes addresses`))
	flags.String("publish-addr", "", heredoc.Doc(`
The address of a Redis node to publish topology.json output`))
	flags.String("topic", defaultRedisTopic, heredoc.Doc(`
The Redis topic where topology.json output will be published`))
	flags.String("output", "", heredoc.Doc(`
Write topology.json output to a file`))
	flags.String("custom-peers", "", heredoc.Doc(`
*Additional* custom peers to (IP,port[,valency]) to add to your target topology.json
eg: "10.0.0.1,3001|10.0.0.2,3002|relays.mydomain.com,3003,3"
`))
}

func init() {
	rootCmd.AddCommand(newFetchCmd())
}

func decodeProducers(arg string) []pkg.Producer {
	out := make([]pkg.Producer, 0)
	producers := strings.Split(arg, "|")
	for _, producer := range producers {
		var port int
		var val int
		s := strings.SplitN(producer, ",", 3)
		port = 0
		if len(s) > 1 {
			port, _ = strconv.Atoi(s[1])
		}
		val = 1
		if len(s) > 2 {
			val, _ = strconv.Atoi(s[2])
		}
		out = append(out, pkg.Producer{
			Addr:    s[0],
			Port:    port,
			Valency: val,
		})
	}
	return out
}

func printResult(out string, filename string) {
	var f *os.File
	var err error
	outputToFile := filename != ""
	if !outputToFile {
		f = os.Stdout
	} else {
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Error opening file: '%s'", filename)
			panic(err)
		}
		defer f.Close()
	}
	writer := bufio.NewWriter(f)
	_, err = writer.WriteString(out)
	if err != nil {
		fmt.Println("Error writing to buffer")
		panic(err)
	}
	writer.Flush()
}

func fetch(cmd *cobra.Command, args []string) {
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
	max, _ := cmd.Flags().GetInt64("max")
	ipv, _ := cmd.Flags().GetInt64("ipv")
	customPeers, _ := cmd.Flags().GetString("custom-peers")

	src, err := pkg.Fetch(context, endpointUrl, magic, max, ipv)
	if err != nil {
		log.Errorf("Unable to get data: %v", err)
		os.Exit(1)
	}
	var payload pkg.PullPayload
	err = json.Unmarshal(src, &payload)
	if err != nil {
		log.Errorf("Unmarshal error: %v\n", err)
		os.Exit(1)
	}
	if customPeers != "" {
		peers := decodeProducers(customPeers)
		payload.Producers = append(payload.Producers, peers...)
	}
	dst := &bytes.Buffer{}
	data, _ := json.Marshal(payload)
	if err := json.Indent(dst, data, "", "  "); err != nil {
		panic(err)
	}

	fname, _ := cmd.Flags().GetString("output")
	printResult(dst.String(), fname)

	publishAddr, _ := cmd.Flags().GetString("publish-addr")
	topic, _ := cmd.Flags().GetString("topic")
	if publishAddr != "" && topic != "" {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     publishAddr,
			Username: os.Getenv("REDIS_USER"),
			Password: os.Getenv("REDISCLI_AUTH"),
		})
		if err := redisClient.Publish(context, topic, dst.String()).Err(); err != nil {
			panic(err)
		}
	}
}
