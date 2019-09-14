package cmd

import (
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/hub"
)

type Specification struct {
	Port               int    `default:"8888"`
	LogLevel           string `split_words:"true" default:"INFO"`
	MuxBufferLength    int    `default:"10"`
	ClientBufferLength int    `default:"5"`
	ClientTimeoutMs    int    `default:"1000"`
	HttpWaitMs         int    `default:"5000"`
	HttpFlushMs        int    `default:"5"`
	HttpTimeoutMs      int    `default:"1000"`
}

func init() {
	rootCmd.AddCommand(streamCmd)
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "stream video",
	Long:  `capture video incoming to http and stream out over websockets`,
	Run: func(cmd *cobra.Command, args []string) {

		var outputs Output
		var s Specification
		var wg sync.WaitGroup
		var writers ToFile

		// load & log configuration
		if err := envconfig.Process("vw", &s); err != nil {
			log.Fatal(err.Error())
		}

		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(sanitiseLevel(s.LogLevel))

		log.WithField("s", s).Info("Specification")

		// declare channels
		httpRunning := make(chan struct{})

		// trap SIGINT
		channelSignal := make(chan os.Signal, 1)
		closed := make(chan struct{})
		signal.Notify(channelSignal, os.Interrupt)
		go waitSignal(closed, channelSignal, &wg)

		// legacy configuration from yaml
		/*
			     if err := viper.Unmarshal(&outputs); err != nil {
					log.WithField("error", err).Fatal("Failed to read output configuration - malformed?")
				}
				populateInputNames(&outputs)

				if err := viper.Unmarshal(&writers); err != nil {
					log.WithField("error", err).Fatal("Failed to read writer configuration - malformed?")
				}

				populateInputNamesForWriters(&writers)
				variables.Vars = viper.GetStringMapString("variables")
				expandDestinations(&outputs, variables)
		*/
		// start our sub processess
		// TODO - setup the comms hub as a separate library

		httpOpts := HTTPOptions{Port: s.Port, WaitMS: s.HttpWaitMs, FlushMS: s.HttpFlushMs, TimeoutMS: s.HttpTimeoutMs}

		clientOpts := ClientOptions{BufferLength: s.ClientBufferLength, TimeoutMS: s.ClientTimeoutMs}

		h := hub.New()
		go h.RunWithStats(closed)

		wg.Add(1)
		go startHttp(closed, &wg, httpOpts, h, httpRunning)

		<-httpRunning //wait for http server to start

		wg.Add(2)
		go startWss(closed, &wg, outputs, h, clientOpts)
		go startWriters(closed, &wg, writers, h)

		wg.Wait()

	},
}
