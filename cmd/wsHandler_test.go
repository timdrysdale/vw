package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/reconws"
)

func TestSendMessageViaClient(t *testing.T) {

	closed := make(chan struct{})

	h := agg.New()
	go h.Run(closed)

	ctx := &hub.Client{Hub: h.Hub, Name: "tx", Topic: "/greetings", Send: make(chan hub.Message), Stats: hub.NewClientStats()}
	crx := &hub.Client{Hub: h.Hub, Name: "rx", Topic: "/greetings", Send: make(chan hub.Message), Stats: hub.NewClientStats()}

	h.Register <- crx

	greeting := []byte("hello")

	m := &hub.Message{Sender: *ctx, Sent: time.Now(), Data: greeting, Type: websocket.TextMessage}
	h.Broadcast <- *m

	time.Sleep(time.Millisecond)

	fmt.Println(h.Hub.Clients)
	fmt.Println("Waiting for message ...")
	select {
	case <-time.After(10 * time.Millisecond):
	case msg, ok := <-crx.Send:
		if ok {
			fmt.Println("Got message...")
			if bytes.Compare(msg.Data, greeting) != 0 {
				t.Errorf("Greeting content unexpected; got/wanted %v/%v\n", string(msg.Data), string(greeting))
			}
		}
	}
}

func TestSendMessageViaWs(t *testing.T) {

	closed := make(chan struct{})

	h := agg.New()
	go h.Run(closed)

	//ctx := &hub.Client{Hub: h.Hub, Name: "tx", Topic: "/greetings", Send: make(chan hub.Message), Stats: hub.NewClientStats()}
	crx := &hub.Client{Hub: h.Hub, Name: "rx", Topic: "/greetings", Send: make(chan hub.Message), Stats: hub.NewClientStats()}

	h.Register <- crx

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { wsHandler(closed, w, r, h) }))

	defer s.Close()

	r := reconws.New()

	r.Url = "ws" + strings.TrimPrefix(s.URL, "http") //+ "/ws/greetings"

	fmt.Println(r.Url)

	go r.Reconnect()

	greeting := []byte("hello")

	go func() {
		m := &reconws.WsMessage{Data: greeting, Type: websocket.TextMessage}
		r.In <- *m
	}()

	time.Sleep(time.Millisecond)

	fmt.Println(h.Hub.Clients)
	fmt.Println("Waiting for message ...")
	select {
	case <-time.After(10 * time.Millisecond):
		t.Error("timed out")
	case msg := <-crx.Send:

		fmt.Println("Got message...")
		if bytes.Compare(msg.Data, greeting) != 0 {
			t.Errorf("Greeting content unexpected; got/wanted %v/%v\n", string(msg.Data), string(greeting))
		}

	}

	//select {
	//case msg := <-crx.Send:
	//	if bytes.Compare(msg.Data, greeting) != 0 {
	//		t.Errorf("Greeting content unexpected; got/wanted %v/%v\n", string(msg.Data), string(greeting))
	//	}
	//default:
	//	t.Error("Failed to receive message")
	//}

	close(closed)
	close(r.Stop)
}

// this test only shows that the httptest server is working ok
func TestWsEcho(t *testing.T) {

	r := reconws.New()

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(echo))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	r.Url = "ws" + strings.TrimPrefix(s.URL, "http")

	go r.Reconnect()

	payload := []byte("Hello")
	mtype := int(websocket.TextMessage)

	r.Out <- reconws.WsMessage{Data: payload, Type: mtype}

	reply := <-r.In

	if bytes.Compare(reply.Data, payload) != 0 {
		t.Errorf("Got unexpected response: %s, wanted %s\n", reply.Data, payload)
	}

}

var testUpgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := testUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, message)
		if err != nil {
			break
		}
	}
}
