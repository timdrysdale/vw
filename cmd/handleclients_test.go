package cmd

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timdrysdale/vw/config"
)

func TestAddDeleteClients(t *testing.T) {

	var topics config.TopicDirectory
	topics.Directory = make(map[string][]config.ClientDetails)

	client1 := randomClient()
	client2 := randomClient()
	client3 := randomClientForTopic(client2.Topic)
	client4 := randomClientForTopic(client2.Topic)

	addClientToTopic(&topics, client1)
	addClientToTopic(&topics, client2)
	addClientToTopic(&topics, client3)
	addClientToTopic(&topics, client4)

	clientList := []config.ClientDetails{client1, client2, client3, client4}
	clientShouldExist := []bool{true, true, true, true}

	for i := range clientList {
		if clientExists(&topics, clientList[i]) != clientShouldExist[i] {
			t.Errorf("bare/addClientToTopic: client %v has WRONG existence status, should be %v\n", i, clientShouldExist[i])
		}
	}

	deleteClientFromTopic(&topics, client1)
	deleteClientFromTopic(&topics, client2)
	deleteClientFromTopic(&topics, client4)

	clientShouldExist = []bool{false, false, true, false}

	for i := range clientList {
		if clientExists(&topics, clientList[i]) != clientShouldExist[i] {
			t.Errorf("bare/deleteClientFromTopic(): client %v has WRONG existence status, should be %v\n", i, clientShouldExist[i])
		}
	}

}

func TestHandler(t *testing.T) {

	//This test fails if you don't give the handler a chance to action the commands, hence the time.Sleep
	var wg sync.WaitGroup
	closed := make(chan struct{})
	clientActionsChan := make(chan config.ClientAction)

	var topics config.TopicDirectory
	topics.Directory = make(map[string][]config.ClientDetails)

	go HandleClients(closed, &wg, &topics, clientActionsChan)

	client1 := config.ClientDetails{"client1", "topic1", make(chan config.Message, 2)}
	client2 := config.ClientDetails{"client2", "topic2", make(chan config.Message, 2)}
	client3 := config.ClientDetails{"client3", "topic2", make(chan config.Message, 2)}
	client4 := config.ClientDetails{"client4", "topic2", make(chan config.Message, 2)}

	clientActionsChan <- config.ClientAction{config.ClientAdd, client1}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client2}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client3}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client4}

	time.Sleep(1 * time.Millisecond)

	clientList := []config.ClientDetails{client1, client2, client3, client4}
	clientShouldExist := []bool{true, true, true, true}

	for i := range clientList {
		if clientExists(&topics, clientList[i]) != clientShouldExist[i] {
			t.Errorf("handler/addClientToTopic: client %v has WRONG existence status, should be %v\n", i, clientShouldExist[i])
		}
	}

	clientActionsChan <- config.ClientAction{config.ClientDelete, client1}
	clientActionsChan <- config.ClientAction{config.ClientDelete, client2}
	clientActionsChan <- config.ClientAction{config.ClientDelete, client4}

	time.Sleep(1 * time.Millisecond)

	clientShouldExist = []bool{false, false, true, false}

	for i := range clientList {
		if clientExists(&topics, clientList[i]) != clientShouldExist[i] {
			t.Errorf("handler/deleteClientFromTopic(): client %v has WRONG existence status, should be %v\n", i, clientShouldExist[i])
			t.Errorf("%v", topics.Directory)
		}
	}

}

func randomClient() config.ClientDetails {
	return config.ClientDetails{uuid.New().String(), uuid.New().String(), make(chan config.Message)}
}

func randomClientForTopic(topic string) config.ClientDetails {
	return config.ClientDetails{uuid.New().String(), topic, make(chan config.Message)}
}
