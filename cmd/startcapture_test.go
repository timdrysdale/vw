package cmd

import (
	"bytes"
	"log"
	"testing"

	"github.com/spf13/viper"
)

var twoFeedExample = []byte(`--- 
commands: 
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontMedium} -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontMedium} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontSmall}"
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoFrontMedium} -s 640x480 -b:v 1024k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideMedium} -s 320x240 -b:v 512k -bf 0 -f mpegts -codec:v mpeg1video ${videoSideSmall}"
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
        - videoFrontSmall`)

var stream0 = Stream{Name: "", Destination: "${host}/${uuid}/${session}/front/medium", InputNames: []string{"audio", "videoFrontMedium"}}
var stream1 = Stream{Name: "", Destination: "${host}/${uuid}/${session}/front/small", InputNames: []string{"audio", "videoFrontSmall"}}
var stream2 = Stream{Name: "", Destination: "${host}/${uuid}/${session}/side/medium", InputNames: []string{"audio", "videoSideMedium"}}
var stream3 = Stream{Name: "", Destination: "${host}/${uuid}/${session}/side/small", InputNames: []string{"audio", "videoSideSmall"}}

var twoFeedOutputs = Output{[]Stream{stream0, stream1, stream2, stream3}}

func TestExpandCaptureCommands(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	err := v.ReadConfig(bytes.NewBuffer(twoFeedExample))
	if err != nil {
		log.Fatalf("read config failed (twoFeedExample)")
	}

	var o Output
	err = unmarshalConfig(v, &o)

	if err != nil {
		t.Errorf("unmarshal outputs failed (twoFeedExample)")
	}

	// order _should_ be the same in this test, won't matter in usage
	for k, v := range o.Streams {
		if clean(v.Destination) != clean(twoFeedOutputs.Streams[k].Destination) {
			t.Errorf("didn't unmarshal outputs %v, %v", v.Destination, twoFeedOutputs.Streams[k].Destination)
		}

	}

}
