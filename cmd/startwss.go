package cmd

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/timdrysdale/hub"

	log "github.com/sirupsen/logrus"
)

func startWss(closed <-chan struct{}, wg *sync.WaitGroup, outputs Output, h *hub.Hub, opts ClientOptions) {
	defer wg.Done()
	for _, stream := range outputs.Streams {
		wg.Add(1)
		name := "wssClient(" + uuid.New().String()[:3] + "):"
		go wssClient(closed, wg, stream, name, h, opts)
		log.WithField("name", name).Info("Spawned WSSclient")
	}
}

func wssClient(closed <-chan struct{}, wg *sync.WaitGroup, stream Stream, name string, h *hub.Hub, opts ClientOptions) {

	defer wg.Done()

	log.WithField("To", stream.Destination).Info("Connecting")

	c, _, err := websocket.DefaultDialer.Dial(stream.Destination, nil)
	if err != nil {
		log.WithField("error", err).Error("Dialing")
		return
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

	messagesForMe := make(chan hub.Message, opts.BufferLength)

	log.WithField("InputNames", stream.InputNames).Debug("Setting up")

	for _, input := range stream.InputNames {

		client := &hub.Client{Hub: h,
			Name:  name,
			Send:  messagesForMe,
			Stats: hub.NewClientStats(),
			Topic: input}

		log.WithFields(log.Fields{"name": name, "input": input}).Info("Subscribing")
		h.Register <- client

		defer func() {
			h.Unregister <- client
			log.WithField("name", name).Fatal("Disconnected")
		}()
	}

	for {
		select {
		case <-done:
			return

		case msg := <-messagesForMe:

			err := c.WriteMessage(msg.Type, msg.Data)
			if err != nil {
				log.WithField("error", err).Fatal("Writing")
				return
			}
		case <-closed:

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection (TODO: where is the timeout?)
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.WithField("error", err).Error("Closing")
			} else {
				log.WithField("name", name).Info("Closed")
			}
			return
		}
	}
}
