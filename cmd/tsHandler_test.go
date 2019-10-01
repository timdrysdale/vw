package cmd

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
)

func TestMessageBoundaries(t *testing.T) {
	//
	// +----------+       +----------+        +---------+             +----------+
	// |          |       |          |        |         |             |          |
	// | ffmpeg   |       | tsHandler|        |  Agg    | crx.Send    |  crx     |
	// |          +------>+  (under  +------->+         +-------------> (hub     |
	// |          |       |   test)  |        |         |             |  test    |
	// |          |       |          |        |         |             |  client) |
	// +----------+       +----------+        +---------+             +----------+
	//
	// TEST HARNESS--><--ITEM UNDER TEST--><---TEST HARNESS---------------------->
	//
	//                                       --diagram created using asciiflow.com
	//
	// Explanation:
	// What's under test is the tsHandler's ability to detect frame boundaries due to the
	// pause between frames that are being posted from (presumably) ffmpeg
	// Obviously this makes the assumption that there is slack time between frames
	// Which _should_ hold for even single board computers and consumer bitrates/framerates
	//
	// Test harness comprises a syscall to ffmpeg to stream some frames and hub client
	// which checks for those messages to have the appropriate size

	closed := make(chan struct{})

	// Test harness, receiving side (agg, and hub.Client)
	h := agg.New()
	go h.Run(closed)

	time.Sleep(2 * time.Millisecond)

	crx := &hub.Client{Hub: h.Hub, Name: "rx", Topic: "/video", Send: make(chan hub.Message), Stats: hub.NewClientStats()}
	h.Register <- crx

	time.Sleep(2 * time.Millisecond)

	// check hubstats
	if len(h.Hub.Clients) != 1 {
		t.Errorf("Wrong number of clients registered to hub wanted/got %d/%d", 1, len(h.Hub.Clients))
	}

	// server to action the handler under test
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { tsHandler(closed, w, r, h) }))

	defer s.Close()

	time.Sleep(2 * time.Millisecond)

	data := make([]byte, 188*1024)
	rand.Read(data)

	// taken from a known good version of the code when running on the sample video
	frameSizes := []int{15980,
		20116,
		17296,
		16544,
		18988,
		15792,
	}

	go func() {
		// did frame sizes come through correctly?
		for i := 0; i < len(frameSizes); i++ {
			select {
			case <-time.After(100 * time.Millisecond):
				t.Errorf("timed out on frame  %d", i)
			case msg, ok := <-crx.Send:
				if ok {
					if len(msg.Data) != frameSizes[i] {
						t.Errorf("Frame %d content size wrong; got/wanted %v/%v\n", i, len(msg.Data), frameSizes[i])
					}
				} else {
					t.Error("channel not ok") //this test seems sensitive to timing off the sleeps, registration delay?
				}
			}
		}
	}()

	time.Sleep(1 * time.Millisecond)

	u, _ := url.Parse(s.URL)

	dest := fmt.Sprintf("http://localhost:%s/ts/video", u.Port())

	args := fmt.Sprintf("-re -i sample.ts -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -r 24 -bf 0 %s", dest)

	argSlice := strings.Split(args, " ")

	cmd := exec.Command("ffmpeg", argSlice...)

	err := cmd.Run()

	if err != nil {
		t.Error("ffmpeg", err)
	}

	// hang on long enough for all the timeouts in the anonymous goroutine to trigger
	time.Sleep(300 * time.Millisecond)

	close(closed)

	time.Sleep(time.Millisecond) //allow time for goroutines to end before starting a new http server
}
