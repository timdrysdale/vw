package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/viper"
)

var twoFeedExample = []byte(`--- 
commands: 
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontMedium} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontSmall}"
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideMedium} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideSmall}"
  - "ffmpeg -f alsa -ar 44100 -i hw:0 -f mpegts -codec:a mp2 -b:a 128k -muxdelay 0.001 -ac 1 -filter:a ''volume=50'' ${audio}"

config: 
  control: 
    path: control
    scheme: http
  host: "wss://video.practable.io:443"
  log: ./vw.log
  retry_wait: 1000
  strict: false
  tuning: 
    bufferSize: 1024000
  uuid: 49270598-9da2-4209-98da-e559f0c587b4
  verbose: false
streams: 
  -   destination: "${host}/${uuid}/${session}/front/medium"
      feeds: 
        - audio
        - videoFrontMedium
  -   destination: "${host}/${uuid}/${session}/front/small"
      feeds: 
        - audio
        - videoFrontSmall
  -   destination: "${host}/${uuid}/${session}/side/medium"
      feeds: 
        - audio
        - videoSideMedium
  -   destination: "${host}/${uuid}/${session}/side/small"
      feeds: 
        - audio
        - videoSideSmall
`)

var streamExample = []byte(`
destination: "${host}/${uuid}/${session}/front/small"
feeds:
  - audio
  - videoFrontSmall`)

var streamFeeds = []string{"audio", "videoFrontSmall"}

var stream0 = Stream{Destination: "${host}/${uuid}/${session}/front/medium", InputNames: []string{"audio", "videoFrontMedium"}}
var stream1 = Stream{Destination: "${host}/${uuid}/${session}/front/small", InputNames: []string{"audio", "videoFrontSmall"}}
var stream2 = Stream{Destination: "${host}/${uuid}/${session}/side/medium", InputNames: []string{"audio", "videoSideMedium"}}
var stream3 = Stream{Destination: "${host}/${uuid}/${session}/side/small", InputNames: []string{"audio", "videoSideSmall"}}

var twoFeedOutputs = Output{[]Stream{stream0, stream1, stream2, stream3}}

var hosturl = "http://127.0.0.1:8080/"
var h, err = url.Parse("http://127.0.0.1:8080/")

var inputName1 = "some/amazing/long/path"
var endpoint1 = strings.Join([]string{hosturl, inputName1}, "")

var command0 = "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoFrontMedium -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoFrontSmall"

var command1 = "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoSideMedium -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoSideSmall"

var command2 = "ffmpeg -f alsa -ar 44100 -i hw:0 -f mpegts -codec:a mp2 -b:a 128k -muxdelay 0.001 -ac 1 -filter:a ''volume=50'' http://127.0.0.1:8080/audio"

var expandedCommands = []string{command0, command1, command2}

var expectedEndpoints = Endpoints{"videoFrontMedium": "http://127.0.0.1:8080/videoFrontMedium",
	"videoFrontSmall": "http://127.0.0.1:8080/videoFrontSmall",
	"videoSideMedium": "http://127.0.0.1:8080/videoSideMedium",
	"videoSideSmall":  "http://127.0.0.1:8080/videoSideSmall"}

var cmd0 = "sleep 0.1"
var cmd1 = "sleep 0.2"
var cmd2 = "sleep 0.3"

var testCommands = Commands{Commands: []string{cmd0, cmd1, cmd2}}

var cmdSlow0 = "sleep 1"
var cmdSlow1 = "sleep 2"
var cmdSlow2 = "sleep 3"

var testCommandsSlow = Commands{Commands: []string{cmdSlow0, cmdSlow1, cmdSlow2}}

func TestExpandCaptureCommands(t *testing.T) {

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(streamExample))
	if err != nil {
		log.Fatalf("read config failed (streamExample) %v", err)
	}

	var stream Stream
	err = v.Unmarshal(&stream)
	if err != nil {
		t.Errorf("read stream failed (streamExample) %v", err)
	} else {

		//https://github.com/go-yaml/yaml/issues/282
		feedSlice, _ := stream.Feeds.([]interface{})
		for i, val := range feedSlice {
			feed := val.(string)
			if feed != streamFeeds[i] {
				t.Errorf("Got wrong feed in stream\n wanted %v\ngot: %v\n", feed, streamFeeds[i])
			}

		}

	}

	v = viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(twoFeedExample))
	if err != nil {
		t.Errorf("read config failed (twoFeedExample)")
	}

	var o Output
	err = v.Unmarshal(&o)

	if err != nil {
		t.Errorf("unmarshal outputs failed (twoFeedExample) %v", err)
	}

	populateInputNames(&o) //copy Feeds to InputNames, thereby converting interface to string

	// ordering doesn't matter in the app, so ok to rely here in test on the order of the
	// yaml entries being the same as validation object
	for k, val := range twoFeedOutputs.Streams {
		if clean(val.Destination) != clean(o.Streams[k].Destination) {
			t.Errorf("destinations don't match %v, %v", val.Destination, twoFeedOutputs.Streams[k].Destination)
		}
		if len(val.InputNames) != len(o.Streams[k].InputNames) {
			t.Errorf("\nMismatch in number of InputNames\n wanted: %v\n got: %v\n", val.InputNames, o.Streams[k].InputNames)
		} else {
			for i, f := range twoFeedOutputs.Streams[k].InputNames {
				if clean(f) != o.Streams[k].InputNames[i] {
					t.Errorf("input names don't match for destination %v (%v != %v)", val.Destination, f, o.Streams[k].InputNames[i])
				}
			}
		}

	}

	var c Commands
	err = v.Unmarshal(&c)

	if err != nil {
		t.Errorf("unmarshal commands failed (twoFeedExample) %v", err)
	}

	if hosturl != h.String() {
		t.Errorf("Error in setting up URL for test, these should match %v, %v", hosturl, h.String())
	}

	// are endpoints are correctly assembled from URL and path?
	testEndpoint := constructEndpoint(h, inputName1)
	if endpoint1 != testEndpoint {
		t.Errorf("Endpoints did not match %v, %v\n", endpoint1, testEndpoint)
	}

	// are endpoints correctly mapped ?
	endpoints := mapEndpoints(o, h)
	for k, val := range expectedEndpoints {
		if endpoints[k] != val {
			t.Errorf("\nEndpoint for %v:\nexpected %v\ngot %v\n", k, val, endpoints[k])
		}
	}

	// expand the commands in-place
	expandCaptureCommands(&c, endpoints)

	for i, expanded := range expandedCommands {
		if clean(expanded) != c.Commands[i] {
			t.Errorf("\nCommand incorrectly expanded\nexp: %v\ngot: %v\n", expanded, c.Commands[i])
		}
	}

	// try sleeping in external processes
	closed := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	runCaptureCommands(closed, &wg, testCommands)
	elapsed := time.Since(start)
	if elapsed < 300*time.Millisecond {
		t.Errorf("running sleep commands was too short %v", elapsed)
	}
	if elapsed > 350*time.Millisecond {
		t.Errorf("running sleep commands took too long %v", elapsed)
	}

	//try killing processes via closed
	wg.Add(1)
	go func() {
		fmt.Println("in closer")
		select {
		case <-time.After(500 * time.Millisecond):
			close(closed)
		}
	}()

	start = time.Now()
	runCaptureCommands(closed, &wg, testCommandsSlow)
	elapsed = time.Since(start)
	if elapsed < 500*time.Millisecond {
		t.Errorf("killing sleep commands was too short %v", elapsed)
	}
	if elapsed > 1000*time.Millisecond {
		t.Errorf("killing sleep commands took too long %v", elapsed)
	}
}
