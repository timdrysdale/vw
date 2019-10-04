package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/rwc"
)

func TestStreamUsingInternals(t *testing.T) {
	//
	//  This integration test is intended to show a file streaming
	//  to a websocket server, using elements of existing tests
	//  but joined together this time ...
	//
	//  +------+   +------+   +------+    +------+    +------+
	//  |      |   |      |   |      |    |      |    |      |
	//  |ffmpeg+--->handle+--->Agg   +---->rwc   +--->+ wss  |
	//  |      |   |Ts    |   |      |    |      |    |      |
	//  +-^----+   +------+   +-^----+    +-^----+    +-----++
	//    |                     |           |               |
	//    |                     |           |               |
	//    +                     +           +               v
	//  sample.ts             stream      destination    check
	//                        rule        rule           frame
	//                                                   sizes
	//

	// start up our streaming programme
	//go streamCmd.Run(streamCmd, nil) //streamCmd will populate the global app
	app := testApp(true)

	time.Sleep(2 * time.Millisecond)

	// server to action the handler under test
	r := mux.NewRouter()
	r.HandleFunc("/ts/{feed}", http.HandlerFunc(app.handleTs))

	s := httptest.NewServer(r)
	defer s.Close()

	time.Sleep(100 * time.Millisecond)

	// Set up our destination wss server and frame size check

	msgSize := make(chan int)

	serverExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reportSize(w, r, msgSize) }))
	defer serverExternal.Close()

	go func() {
		// did frame sizes come through correctly?
		frameSizes := []int{15980,
			20116,
			17296,
			16544,
			18988,
		}

		time.Sleep(100 * time.Millisecond) //give ffmpeg time to start before looking for frames

		for i := 0; i < len(frameSizes); i++ {
			select {
			case <-time.After(100 * time.Millisecond):
				t.Errorf("timed out on frame  %d", i)
			case frameSize, ok := <-msgSize:
				if ok {
					if frameSize != frameSizes[i] {
						t.Errorf("Frame size %d  wrong; got/wanted %v/%v\n", i, frameSize, frameSizes[i])
					}
				} else {
					t.Error("channel not ok")
				}
			}
		}
	}()

	time.Sleep(1 * time.Millisecond)

	// set up our rules (we've not got audio, but use stream for more thorough test
	streamRule := agg.Rule{Stream: "/stream/large", Feeds: []string{"video0", "audio"}}
	app.Hub.Add <- streamRule

	ue, _ := url.Parse(serverExternal.URL)
	wssUrl := fmt.Sprintf("ws://localhost:%s", ue.Port())
	destinationRule := rwc.Rule{Stream: "/stream/large", Destination: wssUrl, Id: "00"}
	app.Websocket.Add <- destinationRule

	time.Sleep(1 * time.Millisecond)

	uv, _ := url.Parse(s.URL)
	dest := fmt.Sprintf("http://localhost:%s/ts/video0", uv.Port())
	//dest := "http://localhost:8888/ts/video"
	args := fmt.Sprintf("-re -i sample.ts -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -r 24 -bf 0 %s", dest)
	argSlice := strings.Split(args, " ")
	cmd := exec.Command("ffmpeg", argSlice...)
	err := cmd.Run()
	if err != nil {
		t.Error("ffmpeg", err)
	}

	// hang on long enough for timeouts in the anonymous goroutine to trigger
	time.Sleep(300 * time.Millisecond)

	close(app.Closed)

	time.Sleep(time.Millisecond) //allow time for goroutines to end before starting a new http server

}

func TestStreamUsingStreamCmd(t *testing.T) {
	//
	//  This integration test is intended to show a file streaming
	//  to a websocket server, using elements of existing tests
	//  but joined together this time ...
	//
	//  +------+   +------+   +------+    +------+    +------+
	//  |      |   |      |   |      |    |      |    |      |
	//  |ffmpeg+--->handle+--->Agg   +---->rwc   +--->+ wss  |
	//  |      |   |Ts    |   |      |    |      |    |      |
	//  +-^----+   +------+   +-^----+    +-^----+    +-----++
	//    |                     |           |               |
	//    |                     |           |               |
	//    +                     +           +               v
	//  sample.ts             stream      destination    check
	//                        rule        rule           frame
	//                                                   sizes
	//

	// start up our streaming programme
	go streamCmd.Run(streamCmd, nil) //streamCmd will populate the global app

	// Set up our destination wss server and frame size check

	msgSize := make(chan int)
	msgSize2 := make(chan int)

	serverExternal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reportSize(w, r, msgSize) }))
	defer serverExternal.Close()

	serverExternal2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reportSize(w, r, msgSize2) }))
	defer serverExternal2.Close()

	fmt.Printf("Server: %s\n", serverExternal.URL)
	fmt.Printf("Server2: %s\n", serverExternal2.URL)

	time.Sleep(100 * time.Millisecond)

	changeChan := make(chan struct{})

	//check frames sent to first server
	go func() {
		// did frame sizes come through correctly?
		frameSizes := []int{15980,
			20116,
			17296,
			16544,
			18988, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		time.Sleep(100 * time.Millisecond) //give ffmpeg time to start before looking for frames

		for i := 0; i < len(frameSizes); i++ {
			select {
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Server 1 timed out on frame  %d", i)
			case frameSize, ok := <-msgSize:
				if ok {
					if frameSize != frameSizes[i] {
						t.Errorf("Server 1 Frame size %d  wrong; got/wanted %v/%v\n", i, frameSize, frameSizes[i])
					}
					if i == 10 {
						close(changeChan)
					}
					if i > 10 {
						t.Errorf("Server 1 received more frames than expected")
					}
				} else {
					t.Error("channel not ok")
				}
			}
		}
	}()

	//check frames sent to second external server
	go func() {
		// did frame sizes come through correctly?
		frameSizes := []int{
			17296,
			16544,
			18988,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		}

		<-changeChan //wait for destination change before starting timing

		for i := 0; i < len(frameSizes); i++ {
			select {
			case <-time.After(200 * time.Millisecond):
				t.Errorf("Server 2 timed out on frame  %d", i)
			case frameSize, ok := <-msgSize2:
				if ok {
					if frameSize != frameSizes[i] {
						t.Errorf("Server 2 frame %d wrong size; got/wanted %v/%v\n", i, frameSize, frameSizes[i])
					}
				} else {
					t.Error("Server 2 channel not ok")
				}
			}
		}
	}()
	time.Sleep(1 * time.Millisecond)

	// set up our rules (we've not got audio, but use stream for more thorough test
	streamRule := agg.Rule{Stream: "/stream/large", Feeds: []string{"video0", "audio"}}
	app.Hub.Add <- streamRule

	ue, _ := url.Parse(serverExternal.URL)
	wssUrl := fmt.Sprintf("ws://localhost:%s", ue.Port())
	Id := "00"
	destinationRule := rwc.Rule{Stream: "/stream/large", Destination: wssUrl, Id: Id}
	app.Websocket.Add <- destinationRule

	time.Sleep(1 * time.Millisecond)

	//default port for the code
	dest := "http://localhost:8888/ts/video0"
	args := fmt.Sprintf("-re -f concat -i list.txt -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -r 24 -bf 0 %s", dest)
	argSlice := strings.Split(args, " ")
	cmd := exec.Command("ffmpeg", argSlice...)
	go func() {
		err := cmd.Run()
		if err != nil {
			t.Error("ffmpeg", err)
		}
	}()

	select {
	case <-changeChan: //wait for two frames to be received
	case <-time.After(time.Second): //avoid hanging if test failed
		t.Errorf("Change of stream not triggered in time - were frames received?")
	}

	// Websocket.Delete <- Id //destinationRule

	time.Sleep(10 * time.Millisecond)

	//change
	ue2, _ := url.Parse(serverExternal2.URL)
	wssUrl = fmt.Sprintf("ws://localhost:%s", ue2.Port())
	destinationRule = rwc.Rule{Stream: "/stream/large", Destination: wssUrl, Id: "00"}
	//app.Websocket.Add <- destinationRule

	// hang on long enough for timeouts/panics in the anonymous goroutine to trigger
	time.Sleep(5 * time.Second)

	close(app.Closed)

	time.Sleep(10 * time.Millisecond)
}

func reportSize(w http.ResponseWriter, r *http.Request, msgSize chan int) {
	c, err := testUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		msgSize <- len(message)
	}
}
