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
	"github.com/regel/cardano-p2p/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"

	"github.com/spf13/viper"
)

// populated via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)
var (
	logLevel string
	cfgFile  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cardano-p2p",
	Short: "A CLI application to update Cardano node topologies",
	Long: `Shelley has been launched without peer-to-peer (p2p) node discovery so that means we will need to manually add trusted nodes in order to configure our topology. This is a critical step as skipping this step will result in your minted blocks being orphaned by the rest of the network.

cardano-p2p is a CLI application used to send and receive node updates. It can be used as an alternative
to api.clio.one service.

Peer to peer node updates should not require external services. All Cardano node pool metadata information is written and signed in the blockchain.
The cardano-p2p application leverages and verifies pool information in the blockchain
and uses this data to produce valid topology files. Therefore, it does not have to communicate with
external services to receive and produce topology files.

The cardano-p2p application is run as a pod in cardano-charts for Kubernetes.`,
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cardano-p2p.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	logLevel := log.ParseLevel(viper.GetString("log-level"))

	// use zap logger instead of default
	var zapConfig zap.Config
	zapConfig = zap.NewProductionConfig()
	var zapLevel zapcore.Level
	switch logLevel {
	case log.Debug:
		zapLevel = zap.DebugLevel
	case log.Info:
		zapLevel = zap.InfoLevel
	case log.Warn:
		zapLevel = zap.WarnLevel
	case log.Error:
		zapLevel = zap.ErrorLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, _ := zapConfig.Build(zap.AddCallerSkip(2))
	defer func() {
		_ = logger.Sync()
	}()
	sugaredLogger := logger.Sugar()
	log.Log = &log.LevelWrapper{
		Logger:   sugaredLogger,
		LogLevel: logLevel,
	}

	log.Infof("cardano-p2p version %s %s [%s]\n", version, date, commit)

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cardano-p2p" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cardano-p2p")
	}

	viper.SetEnvPrefix("P2P")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file:", viper.ConfigFileUsed())
	}
}
