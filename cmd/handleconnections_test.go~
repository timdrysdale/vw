package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/phayes/freeport"
)

// func HandleConnections(closed <-chan struct{}, wg *sync.WaitGroup, clientActionsChan chan clientAction, messagesFromMe chan message)

func TestHandleConnections(t *testing.T) {

	var wg sync.WaitGroup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	messagesToDistribute := make(chan message) //TODO make buffer length configurable
	var topics topicDirectory
	topics.directory = make(map[string][]clientDetails)
	clientActionsChan := make(chan clientAction)
	closed := make(chan struct{})
	go func() {
		for _ = range c {

			close(closed)
			wg.Wait()
			os.Exit(1)

		}
	}()

	bufferSize = 32798

	port, err := freeport.GetFreePort()
	if err != nil {
		t.Errorf("Error getting free port %v", err)
	}
	fmt.Printf("port: %v\n", port)

	listen = fmt.Sprintf("ws://127.0.0.1:%v", port)

	host, err = url.Parse(listen)

	wg.Add(3)
	//func HandleConnections(closed <-chan struct{}, wg *sync.WaitGroup, clientActionsChan chan clientAction, messagesFromMe chan message)
	go HandleConnections(closed, &wg, clientActionsChan, messagesToDistribute, host)

	//func HandleMessages(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, messagesChan <-chan message)
	go HandleMessages(closed, &wg, &topics, messagesToDistribute)

	//func HandleClients(closed <-chan struct{}, wg *sync.WaitGroup, topics *topicDirectory, clientActionsChan chan clientAction)
	go HandleClients(closed, &wg, &topics, clientActionsChan)

	//wait for server to be up?
	time.Sleep(10 * time.Millisecond)

	topic1 := fmt.Sprintf("%v/in/stream01", listen)
	topic2 := fmt.Sprintf("%v/in/stream02", listen)
	i1 := "1"
	i2 := "2"

	go clientReceiveJSON(t, topic1, i1)
	go clientReceiveJSON(t, topic1, i1)
	go clientReceiveJSON(t, topic1, i1)
	go clientReceiveJSON(t, topic2, i2)
	go clientReceiveJSON(t, topic2, i2)
	go clientReceiveJSON(t, topic2, i2)

	//time.Sleep(10 * time.Millisecond)

	go clientSendJSON(t, topic1, i1)
	go clientSendJSON(t, topic2, i2)
	go clientSendJSON(t, topic1, i1)
	go clientSendJSON(t, topic2, i2)
	time.Sleep(1000 * time.Millisecond)
}

// This example dials a server, writes a single JSON message and then

func clientSendJSON(t *testing.T, url string, msgText string) {

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), url)

	if err != nil {
		fmt.Printf("%s can not connect: %v\n", msgText, err)
	} else {
		fmt.Printf("%s connected\n", msgText)
		msg := []byte(msgText)
		err = wsutil.WriteClientMessage(conn, ws.OpText, msg)
		if err != nil {
			fmt.Printf("%s can not send: %v\n", msgText, err)
			return
		} else {
			fmt.Printf("%s send: %s, type: %v\n", msgText, msg, ws.OpText)
		}

		//msg, op, err := wsutil.ReadServerData(conn)
		//if err != nil {
		//	fmt.Printf("%s can not receive: %v\n", i, err)
		//	return
		//} else {
		//	fmt.Printf("%s receive: %s，type: %v\n", i, msg, op)
		//}

		time.Sleep(time.Duration(2) * time.Second)

		err = conn.Close()
		if err != nil {
			fmt.Printf("%s can not close: %v\n", msgText, err)
		} else {
			fmt.Printf("%s closed\n", msgText)
		}
	}

}

func clientReceiveJSON(t *testing.T, url string, msgText string) {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), url)

	if err != nil {
		fmt.Printf("%s can not connect: %v\n", msgText, err)
	} else {
		fmt.Printf("%s connected\n", msgText)

		//msg := []byte(msgText)
		//err = wsutil.WriteClientMessage(conn, ws.OpText, msg)
		//if err != nil {
		//	fmt.Printf("%s can not send: %v\n", msgText, err)
		//	return
		//} else {
		//	fmt.Printf("%s send: %s, type: %v\n", msgText, msg, ws.OpText)
		//}
		for i := 0; i <= 2; i++ {
			msg, op, err := wsutil.ReadServerData(conn)
			if err != nil {
				fmt.Printf("%s can not receive: %v\n", msgText, err)
				return
			} else {
				if string(msg) == msgText {
					fmt.Printf("OK:")
				} else {
					fmt.Printf("WRONG:")
				}
				fmt.Printf("%s receive: %s，type: %v\n", msgText, msg, op)
			}
		}
		time.Sleep(time.Duration(1) * time.Second)

		err = conn.Close()
		if err != nil {
			fmt.Printf("%s can not close: %v\n", msgText, err)
		} else {
			fmt.Printf("%s closed\n", msgText)
		}
	}

}
