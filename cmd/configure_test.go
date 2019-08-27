package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/timdrysdale/vw/config"
)

var twoFeedExample = []byte(`--- 
commands: 
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontMedium/some/other/path} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontSmall}"
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i ${myspecialvideo} -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideMedium} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideSmall}"
  - "ffmpeg -f alsa -ar 44100 -i hw:0 -f mpegts -codec:a mp2 -b:a 128k -muxdelay 0.001 -ac 1 -filter:a ''volume=50'' ${audio}"
control: 
  path: control
  scheme: http
outurl: "wss://video.practable.io:443"
log: ./vw.log
retry_wait: 1000
strict: false
tuning: 
bufferSize: 1024000
uuid: 49270598-9da2-4209-98da-e559f0c587b4
session: 7525cb39-554e-43e1-90ed-3a97e8d1c6bf
verbose: false
streams: 
  -   destination: "${outurl}/${uuid}/${session}/front/medium"
      feeds: 
        - audio
        - "/videoFrontMedium/some/other/path/"
  -   destination: "${outurl}/${uuid}/${session}/front/small"
      feeds: 
        - audio
        - "/videoFrontSmall"
  -   destination: "${outurl}/${uuid}/${session}/side/medium"
      feeds: 
        - audio
        - "videoSideMedium/"
  -   destination: "${outurl}/${uuid}/${session}/side/small"
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

var stream0 = config.StreamDetails{To: "${outurl}/${uuid}/${session}/front/medium", InputNames: []string{"/audio", "/videoFrontMedium/some/other/path"}}
var stream1 = config.StreamDetails{To: "${outurl}/${uuid}/${session}/front/small", InputNames: []string{"/audio", "/videoFrontSmall"}}
var stream2 = config.StreamDetails{To: "${outurl}/${uuid}/${session}/side/medium", InputNames: []string{"/audio", "/videoSideMedium"}}
var stream3 = config.StreamDetails{To: "${outurl}/${uuid}/${session}/side/small", InputNames: []string{"/audio", "/videoSideSmall"}}

var twoFeedOutputs = config.Streams{[]config.StreamDetails{stream0, stream1, stream2, stream3}}

var expectedChannelCountForClientMap = map[string]int{"wss://somwhere.nice:123/x8786x/y987y/front/medium": 2,
	"wss://somwhere.nice:123/x8786x/y987y/front/small": 2,
	"wss://somwhere.nice:123/x8786x/y987y/side/medium": 2,
	"wss://somwhere.nice:123/x8786x/y987y/side/small":  2}

var expectedChannelCountForFeedMap = map[string]int{"/audio": 4, "/videoFrontMedium/some/other/path": 1, "/videoFrontSmall": 1, "/videoSideMedium": 1, "/videoSideSmall": 1}

var hosturl = "http://127.0.0.1:8080/"
var h, err = url.Parse("http://127.0.0.1:8080/")

var inputName1 = "some/amazing/long/path"
var endpoint1 = strings.Join([]string{hosturl, inputName1}, "")

var command0 = "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoFrontMedium/some/other/path -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoFrontSmall"

//note env var ${myspecialvideo} that should not be altered by vw
var command1 = "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i ${myspecialvideo} -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoSideMedium -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video http://127.0.0.1:8080/videoSideSmall"

var command2 = "ffmpeg -f alsa -ar 44100 -i hw:0 -f mpegts -codec:a mp2 -b:a 128k -muxdelay 0.001 -ac 1 -filter:a ''volume=50'' http://127.0.0.1:8080/audio"

var expandedCommands = []string{command0, command1, command2}

var expectedEndpoints = config.Endpoints{"/videoFrontMedium/some/other/path": "http://127.0.0.1:8080/videoFrontMedium/some/other/path",
	"/videoFrontSmall": "http://127.0.0.1:8080/videoFrontSmall",
	"/videoSideMedium": "http://127.0.0.1:8080/videoSideMedium",
	"/videoSideSmall":  "http://127.0.0.1:8080/videoSideSmall",
	"/audio":           "http://127.0.0.1:8080/audio"}

func TestUnmarshallStream(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(streamExample))
	if err != nil {
		log.Fatalf("read config failed (streamExample) %v", err)
	}

	var stream config.StreamDetails
	err = v.Unmarshal(&stream)
	if err != nil {
		t.Errorf("read stream failed (streamExample) %v", err)
	} else {

		//https://github.com/go-yaml/yaml/issues/282
		feedSlice, _ := stream.From.([]interface{})
		for i, val := range feedSlice {
			feed := val.(string)
			if feed != streamFeeds[i] {
				t.Errorf("Got wrong feed in stream\n wanted %v\ngot: %v\n", feed, streamFeeds[i])
			}
		}
	}
}

func TestConfigureForCommands(t *testing.T) {

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(streamExample))
	if err != nil {
		log.Fatalf("read config failed (streamExample) %v", err)
	}

	var stream config.StreamDetails
	err = v.Unmarshal(&stream)
	if err != nil {
		t.Errorf("read stream failed (streamExample) %v", err)
	} else {

		//https://github.com/go-yaml/yaml/issues/282
		feedSlice, _ := stream.From.([]interface{})
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

	var conf config.Configuration
	err = v.Unmarshal(&conf)

	var o config.Streams

	fmt.Printf("Streams:\n%v\n", o)

	if err != nil {
		t.Errorf("unmarshal outputs failed (twoFeedExample) %v", err)
	}

	populateInputNames(&o) //copy Feeds to InputNames, thereby converting interface to string

	// ordering doesn't matter in the app, so ok to rely here in test on the order of the
	// yaml entries being the same as validation object
	fmt.Printf("twoFeedOutputs %v:\n", twoFeedOutputs.Stream[0])
	fmt.Printf("twoFeedOutputs %v:\n", twoFeedOutputs.Stream[1])
	fmt.Printf("twoFeedOutputs %v:\n", twoFeedOutputs.Stream[2])
	fmt.Printf("twoFeedOutputs %v:\n", twoFeedOutputs.Stream[3])

	for k, val := range twoFeedOutputs.Stream {
		fmt.Printf("k is %d\n", k)
		if clean(val.To) != clean(o.Stream[k].To) {
			t.Errorf("destinations don't match %v, %v", val.To, twoFeedOutputs.Stream[k].To)
		}
		if len(val.InputNames) != len(o.Stream[k].InputNames) {
			t.Errorf("\nMismatch in number of InputNames\n wanted: %v\n got: %v\n", val.InputNames, o.Stream[k].InputNames)
		} else {
			for i, f := range twoFeedOutputs.Stream[k].InputNames {
				if clean(f) != o.Stream[k].InputNames[i] {
					t.Errorf("input names don't match for destination %v (%v != %v)", val.To, f, o.Stream[k].InputNames[i])
				}
			}
		}
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
}

func TestExpandCaptureCommands(t *testing.T) {

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(twoFeedExample))
	if err != nil {
		t.Errorf("read config failed (twoFeedExample)")
	}

	var c config.Commands
	err = v.Unmarshal(&c)

	if err != nil {
		t.Errorf("unmarshal commands failed (twoFeedExample) %v", err)
	}

	// expand the commands in-place
	expandCaptureCommands(&c, expectedEndpoints)

	for i, expanded := range expandedCommands {
		if clean(expanded) != c.Commands[i] {
			t.Errorf("\nCommand incorrectly expanded\nexp: %v\ngot: %v\n", expanded, c.Commands[i])
		}
	}
}

func TestConfigureChannels(t *testing.T) {

	//known-good output configuration from earlier
	o := twoFeedOutputs

	channelBufferLength := 2 //we're not doing much with them, make bigger in production

	channelList := make([]config.ChannelDetails, 0)

	outurl := "wss://somwhere.nice:123"
	uuid := "x8786x"
	session := "y987y"
	configureChannels(o, channelBufferLength, &channelList, outurl, uuid, session)

	if len(channelList) != 8 {
		t.Errorf("Wrong number of channels configured; expected 8, got %d", len(channelList))
	}

	// check a channel is useable, and not duplicated
	channelList[3].Channel <- config.Packet{Data: []byte("h")}
	channelList[3].Channel <- config.Packet{Data: []byte("h")}

	for i := 0; i <= 1; i++ {
		select {

		case <-channelList[0].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[1].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[2].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[3].Channel:
			//got the right channel, phew.

		case <-channelList[4].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[5].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[6].Channel:
			t.Errorf("Wrong channel got the message")

		case <-channelList[7].Channel:
			t.Errorf("Wrong channel got the message")

		case <-time.After(time.Millisecond):
			t.Errorf("Channels timed out")
		}
	}

}

func TestMakeFeedMap(t *testing.T) {

	//known-good output configuration from earlier
	o := twoFeedOutputs

	channelBufferLength := 2 //we're not doing much with them, make bigger in production

	channelList := make([]config.ChannelDetails, 0)
	outurl := "wss://somwhere.nice:123"
	uuid := "x8786x"
	session := "y987y"
	configureChannels(o, channelBufferLength, &channelList, outurl, uuid, session)

	feedMap := make(config.FeedMap)

	configureFeedMap(&channelList, feedMap)

	if len(feedMap) != len(expectedChannelCountForFeedMap) {
		t.Errorf("Wrong number of entries in feedMap; expected %d, got %d", len(expectedChannelCountForFeedMap), len(feedMap))
	}

	for feed, channels := range feedMap {
		if len(channels) != expectedChannelCountForFeedMap[feed] {
			t.Errorf("Wrong number of channels associated with feed %s; expected %d got %d", feed, expectedChannelCountForFeedMap[feed], len(channels))
		}
	}
}

func TestMakeClientMap(t *testing.T) {

	//known-good output configuration from earlier
	o := twoFeedOutputs

	channelBufferLength := 2 //we're not doing much with them, make bigger in production

	channelList := make([]config.ChannelDetails, 0)

	outurl := "wss://somwhere.nice:123"
	uuid := "x8786x"
	session := "y987y"
	configureChannels(o, channelBufferLength, &channelList, outurl, uuid, session)

	clientMap := make(config.ClientMap)

	configureClientMap(&channelList, clientMap)

	if len(clientMap) != len(expectedChannelCountForClientMap) {
		t.Errorf("Wrong number of entries in feedMap; expected %d, got %d", len(expectedChannelCountForClientMap), len(clientMap))
	}

	for feed, channels := range clientMap {
		if len(channels) != expectedChannelCountForClientMap[feed] {
			t.Errorf("Wrong number of channels associated with feed %s; expected %d got %d", feed, expectedChannelCountForClientMap[feed], len(channels))
		}
	}
}

func TestGetUrlOut(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(bytes.NewBuffer(twoFeedExample))
	if err != nil {
		log.Fatalf("read config failed (twoFeedExample) %v", err)
	}

	if !v.IsSet("outurl") {
		t.Errorf("Outgoing URL outurl is not set\n")
	}
	outurl := v.GetString("outurl")
	expected := "wss://video.practable.io:443"
	if outurl != expected {
		t.Errorf("Error getting host from config. Wanted %v, got %v\n", expected, outurl)
	}

}
