package cmd

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/rwc"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Specification struct {
	Port               int    `default:"8888"`
	LogLevel           string `split_words:"true" default:"ERROR"`
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

var app App

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "stream video",
	Long:  `capture video incoming to http and stream out over websockets`,
	Run: func(cmd *cobra.Command, args []string) {
		//Websocket has to be instantiated AFTER the Hub
		app = App{Hub: agg.New(), Closed: make(chan struct{})}
		app.Websocket = rwc.New(app.Hub)

		// load configuration from environment variables VW_<var>
		if err := envconfig.Process("vw", &app.Opts); err != nil {
			log.Fatal("Configuration Failed", err.Error())
		}

		//set log format
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(sanitiseLevel(app.Opts.LogLevel))

		//log configuration
		log.WithField("s", app.Opts).Info("Specification")

		// trap SIGINT
		channelSignal := make(chan os.Signal, 1)
		closed := make(chan struct{})
		signal.Notify(channelSignal, os.Interrupt)
		go func() {
			for _ = range channelSignal {
				close(closed)
				app.WaitGroup.Wait()
				os.Exit(1)
			}
		}()

		//TODO add waitgroup into agg/hub and rwc

		go app.Hub.RunWithStats(closed)

		go app.Websocket.Run(closed)

		app.WaitGroup.Add(1)
		go app.startHttp()

		// take it easy, pal
		app.WaitGroup.Wait()

	},
}
