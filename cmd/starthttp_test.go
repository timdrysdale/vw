package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/phayes/freeport"
)

var msg1 = []byte{'f', 'o', 'o'}
var msg2 = []byte{'b', 'a', 'r'}
var packet1 = Packet{Data: msg1}
var packet2 = Packet{Data: msg2}

func TestChannelRouting(t *testing.T) {

	//make a feedmap
	ch1a := make(chan Packet, 6) //buffer so we can do all the writing first
	ch1b := make(chan Packet, 6)
	ch2 := make(chan Packet, 6)

	feedmap := make(FeedMap)
	feedmap["/ch1"] = make([]chan Packet, 2)
	feedmap["/ch2"] = make([]chan Packet, 1)

	feedmap["/ch1"][0] = ch1a
	feedmap["/ch1"][1] = ch1b
	feedmap["/ch2"][0] = ch2

	feedmap["/ch1"][0] <- packet1
	feedmap["/ch1"][1] <- packet1
	feedmap["/ch2"][0] <- packet2

	got1 := <-ch1a
	got2 := <-ch1b
	got3 := <-ch2

	if string(got1.Data) != string(msg1) {
		t.Errorf("channels in feedmap mismatch: got %v, wanted %v", got1, msg1)
	}

	if string(got2.Data) != string(msg1) {
		t.Errorf("channels in feedmap mismatch: got %v, wanted %v", got2, msg1)
	}

	if string(got3.Data) != string(msg2) {
		t.Errorf("channels in feedmap mismatch: got %v, wanted %v", got3, msg2)
	}

	//get a free port
	port, err := freeport.GetFreePort()
	if err != nil {
		fmt.Printf("Error getting free port %v", err)
	}

	addr := fmt.Sprintf("http://127.0.0.1:%d", port)
	var url1 = fmt.Sprintf("%s/ch1", addr)
	var url2 = fmt.Sprintf("%s/ch2", addr)

	var wg sync.WaitGroup
	wg.Add(1)

	srv := startHttpServer(&wg, port, feedmap)
	time.Sleep(time.Second)

	req1, err := http.NewRequest("POST", url1, bytes.NewBuffer(msg1))
	if err != nil {
		t.Errorf("\nError constructing request %v\n", err)
	}
	req2, err := http.NewRequest("POST", url2, bytes.NewBuffer(msg2))
	if err != nil {
		t.Errorf("\nError constructing request %v\n", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req1)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	resp, err = client.Do(req2)
	if err != nil {
		panic(err)
	}

	resp.Body.Close()

	select {
	case got4 := <-ch1a:
		if !bytes.Equal(got4.Data, msg1) {
			t.Errorf("muxed message does not match: got %v, wanted %v", got4, msg1)
		}
	case <-time.After(10 * time.Millisecond):
		t.Errorf("timeout on ch1a")
	}
	select {
	case got5 := <-ch1b:
		if !bytes.Equal(got5.Data, msg1) {
			t.Errorf("muxed message does not match: got %v, wanted %v", got5, msg1)
		}
	case <-time.After(10 * time.Millisecond):
		t.Errorf("timeout on ch1b")
	}
	select {
	case got6 := <-ch2:
		if !bytes.Equal(got6.Data, msg2) {
			t.Errorf("muxed message does not match: got %v, wanted %v", got6, msg2)
		}
	case <-time.After(10 * time.Millisecond):
		t.Errorf("timeout on ch1c")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO make configurable
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

}
