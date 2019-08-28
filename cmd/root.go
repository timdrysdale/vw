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
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var listen string
var output Output
var inputChannels = make(map[string]chan Packet)
var inputAddresses = make(map[string]string)
var channelList []ChannelDetails
var channelBufferLength int
var cpuprofile string
var memprofile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vw",
	Short: "VW video websockets transporter",
	Long:  `VW initialises video and audio captures by syscall, receiving streams via http to avoid pipe latency issues, then forwards combinations of those streams to websocket servers`,
	Run: func(cmd *cobra.Command, args []string) {

		var wg sync.WaitGroup

		//channelBufferLength := 10 //TODO make configurable
		if viper.IsSet("mux.bufferLength") {
			channelBufferLength = viper.GetInt("mux.bufferLength")
		}
		//channelList := make([]ChannelDetails, 0)
		channelSignal := make(chan os.Signal, 1)
		//clientMap := make(ClientMap) //TODO - delete
		closed := make(chan struct{})
		//feedMap := make(FeedMap) //TODO - delete

		signal.Notify(channelSignal, os.Interrupt)

		go func() {
			if memprofile != "" {
				time.Sleep(10 * time.Second)
				f, err := os.Create(memprofile)
				if err != nil {
					log.Fatal("Could not create memory profile:", err)
				}
				defer f.Close()
				if err := pprof.WriteHeapProfile(f); err != nil {
					log.Fatal("could not write memory profile: ", err)
				}
				defer pprof.StopCPUProfile()
				close(closed)
			}

		}()

		go func() {
			for _ = range channelSignal {
				close(closed)
				wg.Wait()
				os.Exit(1)
			}
		}()

		if cpuprofile != "" {
			f, err := os.Create(cpuprofile)
			if err != nil {
				log.Fatal("Could not create CPU profile: ", err)
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Printf("Could not start CPU profile: %v\n", err)
			}
			defer pprof.StopCPUProfile()

		}

		var port int

		if viper.IsSet("http.port") {
			port = viper.GetInt("http.port")
		} else {
			var err error
			port, err = freeport.GetFreePort()
			if err != nil {
				panic(err)
			}
		}

		fmt.Println(port)

		var outputs Output
		err := viper.Unmarshal(&outputs)
		if err != nil {
			log.Fatalf("Viper unmarshal outputs failed: %v", err)
		}

		populateInputNames(&outputs)

		var writers ToFile
		err = viper.Unmarshal(&writers)
		if err != nil {
			log.Fatalf("Viper unmarshal writers failed: %v", err)
		}

		populateInputNamesForWriters(&writers)

		fmt.Printf("Writers:\n%v\n", writers)

		var variables Variables
		variables.Vars = viper.GetStringMapString("variables")

		var captureCommands Commands

		listen = fmt.Sprintf("http://127.0.0.1:%d", port)

		fmt.Printf("Vars: %v\n", variables)
		fmt.Printf("Vars via viper: %v\n", viper.Get("variables"))

		h := getHost()

		endpoints := mapEndpoints(outputs, h)

		err = viper.Unmarshal(&captureCommands)
		if err != nil {
			log.Fatalf("Viper unmarshal commands failed: %v", err)
		}

		expandCaptureCommands(&captureCommands, endpoints, variables)
		var topics topicDirectory
		topics.directory = make(map[string][]clientDetails)
		clientActionsChan := make(chan clientAction)
		messagesToDistribute := make(chan message, channelBufferLength)

		httpOpts := HTTPOptions{Port: 8080, WaitMS: 5000, FlushMS: 5, TimeoutMS: 1000}

		if viper.IsSet("http.port") {
			httpOpts.Port = viper.GetInt("http.port")
		}
		if viper.IsSet("http.waitMS") {
			httpOpts.WaitMS = viper.GetInt("http.waitMS")
		}
		if viper.IsSet("http.flushMS") {
			httpOpts.FlushMS = viper.GetInt("http.flushMS")
		}
		if viper.IsSet("http.timeoutMS") {
			httpOpts.TimeoutMS = viper.GetInt("http.timeoutMS")
		}

		fmt.Printf("http: %v\n", httpOpts)

		wg.Add(1)
		go startHttp(closed, &wg, *h, httpOpts, messagesToDistribute)

		expandDestinations(&outputs, variables)

		fmt.Printf("\nOutputs:\n%v\n", outputs)

		clientOpts := ClientOptions{BufferLength: 10, TimeoutMS: 1000}

		if viper.IsSet("client.BufferLength") {
			clientOpts.BufferLength = viper.GetInt("client.BufferLength")
		}
		if viper.IsSet("client.TimeoutMS") {
			clientOpts.TimeoutMS = viper.GetInt("client.TimeoutMS")
		}
		fmt.Printf("client: %v\n", clientOpts)

		wg.Add(3)
		go HandleMessages(closed, &wg, &topics, messagesToDistribute)
		go HandleClients(closed, &wg, &topics, clientActionsChan)
		go startWss(closed, &wg, outputs, clientActionsChan, clientOpts)
		go startWriters(closed, &wg, writers, clientActionsChan)

		// TODO wait until the http server is up - maybe send a test response? or have it signal on a channel?
		time.Sleep(1000 * time.Millisecond)

		wg.Add(1)
		go runCaptureCommands(closed, &wg, captureCommands)

		wg.Wait()
		time.Sleep(time.Second)

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
	rootCmd.PersistentFlags().StringVarP(&cpuprofile, "cpuprofile", "p", "", "write cpu profile to `file`")
	rootCmd.PersistentFlags().StringVarP(&memprofile, "memprofile", "m", "", "write memory profile to `file`")

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
