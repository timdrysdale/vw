package cmd

import (
	"sync"
)

func HandleClients(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, clientActionsChan chan clientAction) {
	defer wg.Done()

	for {
		select {
		case <-closed:
			return
		case request := <-clientActionsChan:
			if request.action == clientAdd {

				addClientToTopic(topics, request.client)

			} else if request.action == clientDelete {
				deleteClientFromTopic(topics, request.client)

			}
		}
	}
}

func addClientToTopic(topics *topicDirectory, client clientDetails) {

	_, exists := topics.directory[client.topic]

	if !exists {
		topics.Lock()
		topics.directory[client.topic] = []clientDetails{client}
		topics.Unlock()
	} else {
		topics.Lock()
		topics.directory[client.topic] = append(topics.directory[client.topic], client)
		topics.Unlock()
	}

}

func deleteClientFromTopic(topics *topicDirectory, client clientDetails) {

	_, exists := topics.directory[client.topic]
	if exists {
		topics.Lock()
		existingClients := topics.directory[client.topic]
		topics.directory[client.topic] = filterClients(existingClients, client)
		topics.Unlock()
	}

}
