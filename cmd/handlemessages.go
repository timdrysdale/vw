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

	// assuming buffered messageChans, all writes should succeed immediately
	for _, destination := range distributionList {
		//don't send to sender
		if destination.name != msg.sender.name {
			//non-blocking write to chan - skips if can't write
			//go func() { destination.messagesChan <- message }()

			destination.messagesChan <- msg //we're dropping a lot of messages so try this for now

			//select {
			//case destination.messagesChan <- msg:
			//	//fmt.Printf("sent %v to %v", destination, msg)
			//default:
			//	fmt.Printf("Warn: not sending message to %v (%v)\n", destination, msg) //TODO log this "properly"
			//}
		}
	}

	topics.Unlock()
}
