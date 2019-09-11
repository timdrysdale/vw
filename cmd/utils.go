package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func clean(in string) string {

	return strings.TrimRight(in, "\r\n")

}

func slashify(path string) string {

	//remove trailing slash (that's for directories)
	path = strings.TrimSuffix(path, "/")

	//ensure leading slash without needing it in config
	path = strings.TrimPrefix(path, "/")
	path = fmt.Sprintf("/%s", path)

	return path

}

func filterClients(clients []clientDetails, filter clientDetails) []clientDetails {
	filteredClients := clients[:0]
	for _, client := range clients {
		if client.name != filter.name {
			filteredClients = append(filteredClients, client)
		}
	}
	return filteredClients
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func sanitiseLevel(level string) logrus.Level {
	switch strings.ToLower(level) {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warning":
		return logrus.WarnLevel
	case "Info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}

func waitSignal(closed chan struct{}, channelSignal chan os.Signal, wg *sync.WaitGroup) {
	for _ = range channelSignal {
		close(closed)
		wg.Wait()
		os.Exit(1)
	}
}
