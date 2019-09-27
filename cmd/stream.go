package cmd

import (
	"os"
	"os/signal"
	"sync"

	"github.com/spf13/cobra"
	"github.com/timdrysdale/agg"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
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

		var s Specification
		var wg sync.WaitGroup

		// load configuration from environment variables VW_<var>
		if err := envconfig.Process("vw", &s); err != nil {
			log.Fatal(err.Error())
		}

		//set log format
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(sanitiseLevel(s.LogLevel))

		//log configuration
		log.WithField("s", s).Info("Specification")

		// declare channels
		httpRunning := make(chan struct{})

		// trap SIGINT
		channelSignal := make(chan os.Signal, 1)
		closed := make(chan struct{})
		signal.Notify(channelSignal, os.Interrupt)
		go func() {
			for _ = range channelSignal {
				close(closed)
				wg.Wait()
				os.Exit(1)
			}
		}()

		httpOpts := HTTPOptions{Port: s.Port, WaitMS: s.HttpWaitMs, FlushMS: s.HttpFlushMs, TimeoutMS: s.HttpTimeoutMs}

		//clientOpts := ClientOptions{BufferLength: s.ClientBufferLength, TimeoutMS: s.ClientTimeoutMs}

		h := agg.New()
		go h.RunWithStats(closed)

		wg.Add(1)
		go startHttp(closed, &wg, httpOpts, h, httpRunning)

		<-httpRunning //wait for http server to start

		//wg.Add(2)

		//go startWriters(closed, &wg, writers, h)

		wg.Wait()

	},
}
