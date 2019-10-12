# VW

![alt text][logo]

![alt text][status] ![alt text][coverage]

Video over websockets, a dynamically-reconfigurable golang alternative to the node.js-based relay in [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg).

## Why?

```vw```'s dynamic reconfigurability is intended to make life easier when deploying remote laboratory experiments that use MPEGTS video streams. 

### Background

Sending MPEGTS streams over websockets is attractive for remote laboratory experiments because they often cannot have public IP addresses, and must sit behind restrictive firewalls that often permit only outgoing ```80/tcp``` and ```443/tcp```. Therefore these experiments cannot act as servers, and must instead rely on a remote relay that can receive their websocket video stream and arrange for it to be forwarded to a user. That remote relay must be readily changeable in order to facilitate cost savings like using short-lived spot-priced servers or for coping with other remote relay failures. In some cases, the reconfigurability is required within the experiment, for cost and privacy reasons. The MPEGTS streams can be comprised of separately-encoded video and audio tracks, perhaps of selectable bitrate for bandwidth efficiency. For privacy reasons, it is desirable to be able add and subtract individual tracks from the outgoing stream, for example to facilitate a room-wide mute function across multiple experiments when humans are present. This is particularly important for beginning hosters who are using their offices or homes for experiments, but don't want to entirely disable the audio (but just mute it when they are present). Obviously, the existing [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg) relay could be made configurable with a config file, and simply restarted as needed. Since there is a delay to starting ffmpeg, it would instead be preferable to hot-swap the destination with a dynamically reconfigurable local relay so that the video connection can be maintained without a 1-sec outage due to ffmpeg restarting. This then requires the local relay to run continuously, and if the destination must be updated because the remote relay has changed, or the audio mute has been toggled, then the local relay must be dynamically reconfigurable. This also happens to resonate with the 12-factor approach of avoiding configuration files in favour of a RESTful API for configuration.

### Related projects
Related ```golang``` projects include [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) which calls itself [websocketd](https://github.com/joewalnes/websocketd) for TCP instead of ```STDIN``` and ```STDOUT```. Note that [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) has a websocket server, not a client as required here.

## Getting started

Download and build in the usual manner. The code relies on additional libraries such as ```timdrysdale/[hub,agg,reconws,rwc]``` so these must be downloaded too.

    $ go get github.com/timdrysdale/vw
	$ cd $GOPATH/src/github/timdrysdale/vw
	$ go get -d ./...
	$ go build

Install the executable in a location that is on your path, e.g.

	$ cp ./vw ~/bin
	$ export PATH=$PATH:~/bin

Set the configuration variables, e.g. the port that ```vw``` will listen on (default is 8888):
	
	$ export VW_PORT=8888

Set the logging level if you wish to see greater or fewer messages

    $ export VW_LOGLEVEL=ERROR

Note that other configuration variables are available, but are for developer use only (see ```cmd/stream.go```). 

Start ```vw``` with the ```stream``` command:

    $ ./vw stream 

Start an ffmpeg video stream and direct it to ```vw```

	 $ ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i /dev/video0 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 http://localhost:8888/ts/video0


Start an ffmpeg audio stream and direct it to ```vw```

	 $ ffmpeg <your audio streaming settings here> http://localhost:8888/ts/audio0

See [ffmpeg/ALSA](https://trac.ffmpeg.org/wiki/Capture/ALSA) for information on identifying your audio input devices.

TODO: provide example settings

Configure the streams

    $ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","feeds":["video0","audio0"]}' http://localhost:8888/api/streams
	$ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://<some.relay.server>/in/video0","id":"0"}' http://localhost:8888/api/destinations

You should immediately be able to see your video streams if your browser is connected to your relay server. See [timdrysdale/crossbar](https://github.com/timdrysdale/crossbar) for a relay server.

## API

The API allows rules to be added, deleted, and listed.

### Rules for feeds and Streams

```feeds``` are individual tracks from ffmpeg, which can be forwarded by using a single destination rule:
	
    $ curl -X POST -H "Content-Type: application/json" -d '{"stream":"video0","destination":"wss://<some.relay.server>/in/video0","id":"0"}' http://localhost:8888/api/destinations 

```streams``` can combine multiple tracks (typically one video and one audio track are all that a player can handle), and require both a stream rule and a destination rule. Streams are prefixed with a mandatory ```stream/``` - note that if you supply a leading '/' then you will not be able to delete the stream because you cannot access an endpoint with '//' in it.

    $ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","feeds":["video0","audio0"]}' http://localhost:8888/api/streams
	$ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://<some.relay.server>/in/video0","id":"0"}' http://localhost:8888/api/destinations

### Updating rules

Existing rules can be updated by simply adding them again, e.g. to mute the audio:

    $ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","feeds":["video0"]}' http://localhost:8888/api/streams

 or to change the destination (note the rule ```id``` is kept the same):

	$ curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://<some.relay.server>/in/video1","id":"0"}' http://localhost:8888/api/destinations

### Seeing existing rules

If you want to see the ```streams``` you have set up:
   
    $ curl -X GET http://localhost:8888/api/streams/all
     {"stream/front/large":["video0","audio0"]}	 

If you want to see the ```destinations``` you have set up:

    $ curl -X GET http://localhost:8888/api/destinations/all
	  {"0":{"id":"0","stream":"stream/front/large","destination":"wss://video.practable.io:443/in/video2"}}

### Deleting individual rules
   
If you want to delete a ```stream``` (response confirms which stream was deleted):

    $ curl -X DELETE http://localhost:8888/api/streams/stream/front/large
     "stream/front/large"

If you want to delete a ```destination``` then refer to its ```id``` (response confirms which ```id``` was deleted)

    $ curl -X DELETE http://localhost:8888/api/destinations/0
      "0"

### Deleting all rules

Simply use the <>/all endpoint, e.g. 

    $ curl -X DELETE http://localhost:8888/api/streams/all

or

    $ curl -X DELETE http://localhost:8888/api/destinations/all


## Identifying your devices

### Cameras

You can find the cameras attached to your system with ```v4l2-ctl```:

	$ v4l2-ctl --list-devices

Since ```/dev/video<N>``` designations can change from reboot to reboot, for production, it is better to configure using the serial number, such as 

	  /dev/v4l/by-id/usb-046d_Logitech_BRIO_nnnnnnnn-video-index0

If your camera needs to be initialised or otherwise set up, e.g. by using ```v4l2-ctl``` to specify a resolution or pixel format, then you may find it helpful to write a shell script that takes the dynamically generated endpoint as an argument. For example (not tested, TODO test), your script ```fmv.sh``` for producing medium-size video using ```ffmpeg``` might contain:

    #!/bin/bash
    v4l2-ctl --device $1 --set-fmt-video=width=640,height=480,pixelformat=YUYV
    ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i $1 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 $2

### Microphones

Use you can use this command on linux:

    $ arecord -l

TODO: This command is not currently known to work (but a previous version of the command I don't have to hand, did work)

    $ ffmpeg -f alsa -i hw:2 -filter:a "volume=50" -f mpegts -codec:a libmp3lame -b:a 24k http://localhost:8888/ts/audio0

Things can get a bit more tricky if you are using ```ffmpeg``` within a docker container, because you need to pass the device and network to the container and you also need to have ```sudo``` permissions in some implementations. There's a description of using alsa and pulseaudio with docker [here](https://github.com/mviereck/x11docker/wiki/Container-sound:-ALSA-or-Pulseaudio).

## Platform specific comments 

#### Linux

Developed and tested on Centos 7 for x86_64, Kernel version 3.10, using ```v4l2```, ```alsa``` and ```ffmpeg```

#### Windows 10

I managed to compile and run an earlier version of the code, but I had issues trying to get ```dshow``` to work with my logitech c920 camera and ```ffmpeg```, so I am deferring further testing until windows-hosted experiments become a higher priority.

#### aarch64

```vw``` has been successfully cross-compiled and used with the aarch64 ARM achitecture of the Odroid N2, running an Ubuntu 18.04.2 LTS flavour linux with kernel version 4.9.1622-22.

		 $ export GOOS = linux
		 $ export GOARCH = arm64
		 $ go build


## Future features 

These are notes to me of possible features to consider, rather than promises to implement

0. Configurator to assist in assigning cameras to feeds 
0. HTTP endpoint to report stats 
0. HTTP endpoint to offer stream pre-view

## Issues

There is an insignificant leakage of goroutines (0.1%) - after 86,400 streams were added for one-second each in one 24hour test, there were an extra 114 RelayOut() and 86 RelayTo() goroutines present (out of some 172,800 that had been started and killed), but memory usage had not appreciably increased (static at 0.1% of total) and CPU usage remained constant (ca 20%). The cause seems fairly subtle but after further golang experience it'll be pretty obvious why this is. There are also some races detected that I don't understand how they occur, which may or may not be related. 

Panics ...

Soak testing + mods have ruled out a number of panics.

close of closed channel
A new one just cropped up -  agg.go:42/hug.go:75



## Internals

This diagram has been superseded by the new design ...

![alt text][internals]


## Historical notes

In case future self wonders why I didn't try this or that ... when I already did.

### yaml
An earlier version of this code required a ```yaml``` configuration file that was read once on starting ```vww```. ```vw``` took responsibility for starting ```ffmpeg```. This version suffered from terminal issues around platform-specific variances in the operation of ```cmd.process.kill()``` that left orphan processes holding onto the USB camera on the odroid-N2. Investigating further, it seemed like an Linux-wide issue that killing grandparents did not kill parents or children, and this would prevent ```vw``` from meeting its promises to cleanly kill ```ffmpeg```. Also, the static configuration turned out to be way less flexible than I hoped and the instant I had got it working I realised I didn't think it would work down the line for large deployments. So I could sovle both issues by splitting the responsibility for running ```ffmpeg``` away from ```vw```, and going for dynamic reconfiguration.

### websocket library
The code was initially developed with ```nhooyr/websocket``` then ```gobwas/ws``` then ```gorilla/websocket```, mainly in response to errors possibly relating to websocket reuse with the first two (I didn't really understand what was going wrong, but things got immediately better with ```gorilla/websocket```). There is no comparative performance disadvantage to ```gorilla/websocket``` in my use case because the ```vw``` usage model is for a few high bandwidth connections, rather than the numerous+sparse use case targeted by the other implementations.



[status]: https://img.shields.io/badge/alpha-do%20not%20use-orange "Alpha status, do not use" 
[logo]: ./img/logo.png "VW logo"
[internals]: ./img/internals.png "Diagram of VW internals showing http server, websocket client, mux, monitor, and syscall for ffmpegs"
[coverage]: https://img.shields.io/badge/coverage-71%25-yellowgreen "coverage 71%"
