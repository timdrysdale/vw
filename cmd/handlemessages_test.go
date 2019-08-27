package cmd

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/timdrysdale/vw/config"
)

func TestHandleMessages(t *testing.T) {

	var wg sync.WaitGroup
	closed := make(chan struct{})
	messagesChan := make(chan config.Message, 10)
	clientActionsChan := make(chan config.ClientAction)

	//setup client handling
	var topics config.TopicDirectory
	topics.Directory = make(map[string][]config.ClientDetails)

	go HandleClients(closed, &wg, &topics, clientActionsChan)
	go HandleMessages(closed, &wg, &topics, messagesChan)

	//make some clients
	client1Chan := make(chan config.Message, 2)
	client2Chan := make(chan config.Message, 2)
	client3Chan := make(chan config.Message, 2)
	client4Chan := make(chan config.Message, 2)

	client1 := config.ClientDetails{"client1", "/topic1", client1Chan}
	client2 := config.ClientDetails{"client2", "/topic2", client2Chan}
	client3 := config.ClientDetails{"client3", "/topic2", client3Chan}
	client4 := config.ClientDetails{"client4", "/topic2", client4Chan}

	clientActionsChan <- config.ClientAction{config.ClientAdd, client1}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client2}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client3}
	clientActionsChan <- config.ClientAction{config.ClientAdd, client4}

	//let the clients be processed
	time.Sleep(1 * time.Millisecond)

	//check clients exist before proceeding
	clientList := []config.ClientDetails{client1, client2, client3, client4}
	clientShouldExist := []bool{true, true, true, true}

	for i := range clientList {
		if clientExists(&topics, clientList[i]) != clientShouldExist[i] {
			t.Errorf("HandleMessages/addClientToTopic: client %v has WRONG existence status, should be %v\n", i, clientShouldExist[i])
		}
	}

	//make test messages
	b1 := []byte{'c', 'r', 'o', 's', 's'}
	var testMessage1 = config.Message{Sender: client1, Op: ws.OpBinary, Data: b1}
	b2 := []byte{'b', 'a', 'r'}
	var testMessage2 = config.Message{Sender: client3, Op: ws.OpBinary, Data: b2}

	// send some messages on behalf the clients
	messagesChan <- testMessage1
	messagesChan <- testMessage2

	//let the messages be processed
	time.Sleep(2 * time.Millisecond)

	//check who got what ...
	msg, err := read(client1Chan, 1*time.Millisecond)
	if err == nil {
		t.Errorf("Client 1 should have got an timeout but got %v,%v", msg, err)
	}

	msg, err = read(client2Chan, 1*time.Millisecond)
	if err != nil && !bytes.Equal(msg.Data, b2) {
		t.Errorf("Client 2 should have got msg but got %v,%v", msg, err)
	}
	msg, err = read(client3Chan, 1*time.Millisecond)
	if err == nil {
		t.Errorf("Client 3 should have got an timeout but got %v,%v", msg, err)
	}
	msg, err = read(client4Chan, 1*time.Millisecond)
	if err != nil && !bytes.Equal(msg.Data, b2) {
		t.Errorf("Client 4 should have got msg but got %v, %v", msg, err)
	}

	// delete a client and see what happens
	clientActionsChan <- config.ClientAction{config.ClientDelete, client4}

	time.Sleep(1 * time.Millisecond)

	//send a msg but this time from client 2 - should only go to client 3
	//swapping clients like this detected that sending clients were being unsubscribed
	//so keep the test like this (sending to a previous sender)
	b3 := []byte{'f', 'o', 'o'}
	var testMessage3 = config.Message{Sender: client2, Op: ws.OpBinary, Data: b2}

	messagesChan <- testMessage3

	time.Sleep(1 * time.Millisecond)

	msg, err = read(client2Chan, 1*time.Millisecond)
	if err == nil {
		t.Errorf("Client 2 should have got an timeout but got %v,%v", msg, err)
	}

	msg, err = read(client3Chan, 1*time.Millisecond)
	if err != nil && !bytes.Equal(msg.Data, b3) {
		t.Errorf("Client 3 should have got msg but got %v,%v,%v", msg, err, topics.Directory)
	}

	msg, err = read(client4Chan, 1*time.Millisecond)
	if err == nil {
		t.Errorf("Client 4 should have got an timeout but got %v,%v", msg, err)
	}
}

func read(messageChannel chan config.Message, timeout time.Duration) (config.Message, error) {

	select {
	case msg := <-messageChannel:
		return msg, nil
	case <-time.After(timeout):
		return config.Message{}, errors.New("timeout reading from channel")
	}

}
