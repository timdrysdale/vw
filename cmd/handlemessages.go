package cmd

import (
	"sync"
)

func HandleMessages(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, messagesChan <-chan message) {
	defer wg.Done()

	for {
		select {
		case <-closed:
			return
		case msg := <-messagesChan:
			go distributeMessage(topics, msg)
		}
	}
}

func distributeMessage(topics *topicDirectory, msg message) {

	// unsubscribing client would close channel so lock throughout
	topics.Lock()

	distributionList := topics.directory[msg.sender.topic]

	for _, destination := range distributionList {

		//don't send to sender
		if destination.name != msg.sender.name {
			select {
			case destination.messagesChan <- msg:
			default:
				//TODO alert monitor to this issue?
				close(destination.messagesChan)
			}
		}
	}
	topics.Unlock()
}
