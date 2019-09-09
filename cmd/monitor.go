package cmd

import (
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func runMonitor(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, feeds Monitor, clientActionsChan chan clientAction) {

	defer wg.Done()

	log.WithField("feeds", feeds.Monitor).Info("Starting monitors")

	for _, feed := range feeds.Monitor { //.([]interface{}) {

		wg.Add(1)

		go monitorFeed(closed, wg, topics, clientActionsChan, slashify(feed)) //.(string))

		log.WithField("feed", slashify(feed)).Info("Spawned feed monitor")

	}

}

func monitorFeed(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, clientActionsChan chan clientAction, feed string) {

	defer wg.Done()

	messagesForMe := make(chan message, 1)

	name := fmt.Sprintf("monitor:%s", feed)

	client := clientDetails{name: name, topic: feed, messagesChan: messagesForMe}

	log.WithFields(log.Fields{"name": name, "feed": feed}).Info("Subscribing")

	clientActionsChan <- clientAction{action: clientAdd, client: client}

	defer func() {
		clientActionsChan <- clientAction{action: clientDelete, client: client}
		log.WithField("name", name).Fatal("Monitor disconnected")
	}()

	for {
		select {
		case <-time.After(2 * time.Second): //TODO make configurable
			// log a fatal error so as to halt vw - we're missing a monitored feed
			log.WithFields(log.Fields{"name": name, "feed": feed}).Fatal("Timeout receiving feed")
			return
		case <-messagesForMe:
			continue
		case <-closed:
			return
		}
	}
}
