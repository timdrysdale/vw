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

func TestHandleConnections(t *testing.T) {

	var wg sync.WaitGroup
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	messagesToDistribute := make(chan message)
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

	port, err := freeport.GetFreePort()
	if err != nil {
		t.Errorf("Error getting free port %v", err)
	}

	listen = fmt.Sprintf("ws://127.0.0.1:%v", port)

	host, err := url.Parse(listen)

	wg.Add(3)

	go HandleConnections(closed, &wg, clientActionsChan, messagesToDistribute, host)

	go HandleMessages(closed, &wg, &topics, messagesToDistribute)

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

	time.Sleep(10 * time.Millisecond)

	go clientSendJSON(t, topic1, i1)
	go clientSendJSON(t, topic2, i2)
	go clientSendJSON(t, topic1, i1)
	go clientSendJSON(t, topic2, i2)

	time.Sleep(10 * time.Millisecond)
}

func clientSendJSON(t *testing.T, url string, msgText string) {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), url)
	if err == nil {
		msg := []byte(msgText)
		wsutil.WriteClientMessage(conn, ws.OpText, msg)
		time.Sleep(time.Duration(2) * time.Second)
		conn.Close()
	} else {
		t.Errorf("ClientSendJSON %v", err)
	}
}

func clientReceiveJSON(t *testing.T, url string, msgText string) {
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), url)
	if err == nil {
		for i := 0; i <= 2; i++ {
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				fmt.Printf("%s can not receive: %v\n", msgText, err)
				return
			} else {
				if string(msg) == msgText {
				} else {
					t.Errorf("WRONG: %s != %s,", string(msg), msgText)
				}
			}
		}
		time.Sleep(time.Duration(1) * time.Second)

		err = conn.Close()
	} else {
		t.Errorf("ClientReceiveJSON %v", err)
	}
}
