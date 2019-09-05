package cmd

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	log "github.com/sirupsen/logrus"
)

func startWss(closed <-chan struct{}, wg *sync.WaitGroup, outputs Output, clientActionsChan chan clientAction, opts ClientOptions) {
	defer wg.Done()
	for _, stream := range outputs.Streams {
		wg.Add(1)
		name := "wssClient(" + uuid.New().String()[:3] + "):"
		go wssClient(closed, wg, stream, name, clientActionsChan, opts)
		log.WithField("name", name).Info("Spawned WSSclient")
	}
}

func wssClient(closed <-chan struct{}, wg *sync.WaitGroup, stream Stream, name string, clientActionsChan chan clientAction, opts ClientOptions) {

	log.WithField("To", stream.Destination).Info("Connecting")

	c, _, err := websocket.DefaultDialer.Dial(stream.Destination, nil)
	if err != nil {
		log.WithField("error", err).Error("Dialing")
		return //TODO is this fatal?
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			//silently drop messages
			_, _, err := c.ReadMessage()
			if err != nil {
				log.WithField("error", err).Error("Reading")
				return
			}

		}
	}()

	messagesForMe := make(chan message, opts.BufferLength)

	log.WithField("InputNames", stream.InputNames).Debug("Setting up")

	for _, input := range stream.InputNames {

		client := clientDetails{name: name, topic: input, messagesChan: messagesForMe}
		log.WithFields(log.Fields{"name": name, "input": input}).Info("Subscribing")
		clientActionsChan <- clientAction{action: clientAdd, client: client}

		defer func() {
			clientActionsChan <- clientAction{action: clientDelete, client: client}
			log.WithField("name", name).Warn("Disconnected")
		}()
	}

	for {
		select {
		case <-done:
			return

		case msg := <-messagesForMe:

			err := c.WriteMessage(websocket.BinaryMessage, msg.data)
			if err != nil {
				log.WithField("error", err).Error("Writing")
				return
			}
		case <-closed:
			log.Info("Closed")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.WithField("error", err).Warn("Closing")
				return
			}
			select {
			case <-done:
			}
			return
		}
	}
}
