package cmd

import (
	"sync"

	"github.com/timdrysdale/vw/config"
)

func HandleClients(closed <-chan struct{}, wg *sync.WaitGroup, topics *config.TopicDirectory, clientActionsChan chan config.ClientAction) {
	defer wg.Done()

	for {
		select {
		case <-closed:
			return
		case request := <-clientActionsChan:
			if request.Action == config.ClientAdd {

				addClientToTopic(topics, request.Client)

			} else if request.Action == config.ClientDelete {
				deleteClientFromTopic(topics, request.Client)

			}
		}
	}
}

func addClientToTopic(topics *config.TopicDirectory, client config.ClientDetails) {

	_, exists := topics.Directory[client.Topic]

	if !exists {
		topics.Lock()
		topics.Directory[client.Topic] = []config.ClientDetails{client}
		topics.Unlock()
	} else {
		topics.Lock()
		topics.Directory[client.Topic] = append(topics.Directory[client.Topic], client)
		topics.Unlock()
	}

}

func deleteClientFromTopic(topics *config.TopicDirectory, client config.ClientDetails) {

	_, exists := topics.Directory[client.Topic]
	if exists {
		topics.Lock()
		existingClients := topics.Directory[client.Topic]
		topics.Directory[client.Topic] = filterClients(existingClients, client)
		topics.Unlock()
	}

}
