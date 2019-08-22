/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

type Stream struct {
	Destination string
	Feeds       []string
}
type Outputs struct {
	Streams []Stream `streams`
}

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vw",
	Short: "VW video websockets transporter",
	Long:  `VW initialises video and audio captures by syscall, receiving streams via http to avoid pipe latency issues, then forwards combinations of those streams to websocket servers`,
	Run: func(cmd *cobra.Command, args []string) {

		var outs Outputs
		err := viper.Unmarshal(&outs)
		if err != nil {
			fmt.Println("Didnt unpack streams config")
		} else {
			for _, stream := range outs.Streams {
				fmt.Printf("d:%v %v\n", stream.Destination, stream.Feeds)
			}
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vw.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		fmt.Println("using config from command line")
	} else {

		// Search config in home directory with name ".vw" (without extension).
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/vw/")
		//viper.AddConfigPath("$HOME/.vw")
		//viper.AddConfigPath("/home/tim/go/src/github.com/timdrysdale/vw") // optionally look for config in the working directory
		viper.AddConfigPath(".")
		viper.SetConfigName("vw")
		//viper.SetConfigFile("/home/tim/go/src/github.com/timdrysdale/vw/vw.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Printf("Error with config file %v\n", err)
	}
}
