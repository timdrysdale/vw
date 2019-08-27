package cmd

import (
	"sync"

	"github.com/timdrysdale/vw/config"
)

func HandleMessages(closed <-chan struct{}, wg *sync.WaitGroup, topics *config.TopicDirectory, messagesChan <-chan config.Message) {
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

func distributeMessage(topics *config.TopicDirectory, msg config.Message) {

	// unsubscribing client would close channel so lock throughout
	topics.Lock()

	distributionList := topics.Directory[msg.Sender.Topic]
	// assuming buffered messageChans, all writes should succeed immediately
	for _, destination := range distributionList {
		//don't send to sender
		if destination.Name != msg.Sender.Name {
			//non-blocking write to chan - skips if can't write
			//go func() { destination.messagesChan <- message }()

			destination.MessagesChan <- msg //we're dropping a lot of messages so try this for now

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
