package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

func startWss(closed <-chan struct{}, wg *sync.WaitGroup, outputs Output, clientActionsChan chan clientAction, opts ClientOptions) {
	defer wg.Done()
	for _, stream := range outputs.Streams {
		wg.Add(1)
		name := "wssClient(" + uuid.New().String()[:3] + "):"
		go wssClient(closed, wg, stream, name, clientActionsChan, opts)
		log.Printf("%s spawned", name)
	}
}

//type ClientOptions struct {
//	BufferLength int `yaml:"bufferLength"`
//	TimeoutMS    in  `yaml:"timeoutMS"`
//}

func wssClient(closed <-chan struct{}, wg *sync.WaitGroup, stream Stream, name string, clientActionsChan chan clientAction, opts ClientOptions) {

	defer wg.Done()
	timeout := time.Duration(opts.TimeoutMS) * time.Millisecond //TODO make configurable
	//flipflop := true
	//subscribe this new client to the topic associated with each input name
	messagesForMe := make(chan message, opts.BufferLength)

	fmt.Printf("\nStream.InputNames %s", stream.InputNames)

	for i, input := range stream.InputNames {

		client := clientDetails{name: name, topic: input, messagesChan: messagesForMe}
		fmt.Printf("\n%d: %s subscribing to %s\n", i, name, input)
		clientActionsChan <- clientAction{action: clientAdd, client: client}

		defer func() {
			clientActionsChan <- clientAction{action: clientDelete, client: client}
			fmt.Printf("Disconnected %v, deleting from topics\n", client)
		}()
	}

	for {
		url := stream.Destination
		fmt.Printf("%s dialing %s\n", name, url) //TODO revert to log
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, _, _, err := ws.DefaultDialer.Dial(ctx, url)

		fmt.Printf("%s dialed %s getting %v\n", name, url, err)

		if err != nil {

			log.Printf("%s can not connect to %s: %v\n", name, url, err)
			select {
			case <-time.After(timeout):
			case <-closed:
				closeConn(conn, name)
				return
			}

		} else {

			log.Printf("%s connected to %s\n", name, url)

			for {
				select {
				case <-closed:
					closeConn(conn, name)
					return
				case msg := <-messagesForMe:
					//if flipflop == true {
					err := wsutil.WriteClientMessage(conn, ws.OpBinary, msg.data)
					//fmt.Printf("\n%s sent %d bytes\n", name, len(msg.data))
					if err != nil {
						log.Printf("%s send error: %v", name, err)
					}

					//}
					//flipflop = !flipflop

				}
			}
		}
	}
}

func closeConn(conn net.Conn, name string) {
	err := conn.Close()
	if err != nil {
		log.Printf("%s can not close: %v", name, err)
	} else {
		log.Printf("%s closed\n", name)
	}

}
