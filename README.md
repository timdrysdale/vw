# VW

![alt text][logo]

![alt text][status]

Video over websockets, a golang alternative to the node.js-based relay in [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg).

## Why?
This implementation adds convenience features for configuration and monitoring that are helpful for unattended deployments in remote laboratory experimentsthat sit behind restrictive network firewalls, and therefore cannot directly act as publically-visible live video servers. ```vw``` facillitates relaying multiple video and audio feeds over websocket to a distant relay server. 

## Related projects
Related ```golang``` projects include [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) which calls itself [websocketd](https://github.com/joewalnes/websocketd) for TCP instead of ```STDIN``` and ```STDOUT```. Note that [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) has a websocket server, not a client as required here.

## Usage

After installation and configuration, streaming can be started by 

    $ vw stream 

If you want to use a config file in a non-standard location, then

    $ vw stream --config=/path/to/your/config.yaml


## Installation 

Download and compile the repo 

    $ go get github.com/timdrysdale/vw
    $ cd $GO_PATH/src/github.com/timdrysdale/vw
    $ go get ./...
    $ go build

### Platform specific comments 

#### Linux

Developed and tested on Centos 7 for x86_64, Kernel version 3.10, using ```v4l2```, ```alsa``` and ```ffmpeg```

#### Windows 10

Compiles and runs the core code, but I had issues trying to get ```dshow``` to work with my logitech c920 camera and ```ffmpeg```, so video streaming has not been demonstated by me (yet). Further investigation has been deferred until windows deployments become a higher priority.

#### aarch64

```vw``` has been successfully cross-compiled and used with the aarch64 ARM achitecture of the Odroid N2, running an Ubuntu 18.04.2 LTS flavour linux with kernal version 4.9.1622-22.

		 $ export GOOS = linux
		 $ export GOARCH = arm64
		 $ go build

Note that on this architecture, ```cmd.process.kill()``` is currently suspected of hanging, so cannot exit cleanly with Ctrl-C and instead need to Ctrl-Z and then ```pkill -9 vw``` to clean up defunct processes that are holding onto webcams. Note that ```pkill``` is installed with 

    $ sudo apt-get install procps  

## Configuration

On linux, a simple single-camera, no-audio, feed can be configured as shown below. There are two main parts to the configuration - commands to run to capture video and/or audio, and websocket servers to dial and forward those video/audio streams.


```
--- 
commands: 
  - "ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i ${camera0} -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 ${video0}"

variables:
  outurl: "wss://video.practable.io:443"
  camera0: /dev/video1

streams: 
  -   destination: "${outurl}/in/video1"
      feeds: 
        - video0
```

To better avoid errors associated with having to repeat yourself, configuration files rely on two kinds of variable expansion:

First, user-defined variables can be specified, and these will be expanded in both capture commands, and stream destinations, such as ```outurl``` and ```camera0``` in this example.

Second, there are implicitly defined variables associated with each feed. Each uniquely-named entry in the ```feeds``` of the ```streams``` corresponds to a unqiue, automatically-created ```http``` endpoint that varies from instance to instance. Thus, any feed name used as a variable in the capture commands is expanded to be the coordinatates of that endpoint. This expansion does NOT apply to stream names because then you'd have two URLs in one, which won't work. In this way, a ```vw``` avoids you having to manually configure http endpoints, and takes over the responsbility for ```tee```-ing identical feeds into different streams. In this example commmand, the variable ```video0``` stands for an endpoint that will be fed into the only stream being output.

Note that for testing purposes, you can run ```vw``` and ```ffmpeg``` separately. In this case, to find the dynamically-allocated endpoint, try a capture command like: 

	 - "echo \"send your MPEG TS stream to ${video0}\""

### Identifying your cameras

You can find the cameras attached to your system with ```v4l2-ctl```:

	$ v4l2-ctl --list-devices

Since ```/dev/video<N>``` designations can change from reboot to reboot, for production, it is better to configure using the serial number, such as 

	  /dev/v4l/by-id/usb-046d_Logitech_BRIO_nnnnnnnn-video-index0

If your camera needs to be initialised or otherwise set up, e.g. by using ```v4l2-ctl``` to specify a resolution or pixel format, then you may find it helpful to write a shell script that takes the dynamically generated endpoint as an argument. For example (not tested, TODO test), your script ```fmv.sh``` for producing medium-size video using ```ffmpeg``` might contain:

    #!/bin/bash
    v4l2-ctl --device $1 --set-fmt-video=width=640,height=480,pixelformat=YUYV
    ffmpeg -f v4l2 -framerate 25 -video_size 640x480 -i $1 -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -bf 0 $2

And then in your ```config.yaml``` you'd put the camera ID as the first argument, and the capture endpoint as the second:

    --- 
    commands: 
      - "./fmv.sh ${camera0} ${video0}"
    
    <snip>

Things can get a bit more tricky if you are using ```ffmpeg``` within a docker container, because you need to pass the device and network to the container and you also need to have ```sudo``` permissions in some implementations. There's a description of using alsa and pulseaudio with docker [here](https://github.com/mviereck/x11docker/wiki/Container-sound:-ALSA-or-Pulseaudio).

## Other features

### Write to file

Stream writers can be configured by adding 

    writers:
      -   file: checklong.ts
          feeds: 
            - video0
       
At present, it is an error to specify an existing file. This behaviour was adopted so that rare artefacts or a reliable test file captured on a streaming test were not inadvertently overwritten by a ham-fisted use of shell history replay. However, I found it pretty annoying myself and so this behavour is likely to be upgraded in some way (see future features)

Writing to disk is accomplished in a separate go-routine to avoid introducing variable latency in the network messaging. There is no attention paid to respecting available disk space because it is expected that the write feature will be used sparingly, and not routinely while serving (see future features).

### Exit on capture failure

Monitors can be configured to exit ```vw``` if a specified capture stream(s) fails (currently hardwired as no frames for two seconds), as an aid to using conventional process monitoring tools:

    monitor
      - video0


## Future features

These are notes to me of possible features to consider, rather than promises to implement

0. Allow endpoints to be configured manually, so that capture commands can be automatically run outside of ```vw```.
0. Configurator to assist in assigning cameras to feeds 
0. Designation of streams as required or optional, with ```vw``` exiting if required stream cannot be produced (e.g. capture failure)
0. HTTP endpoint to control whether a stream includes the audio track(s) or not
0. HTTP endpoint to report stats 
0. HTTP endpoint to offer stream pre-view
0. Allow writers to overwrite existing files automatically, and/or to append a salt like date/time.
0. Permit some form of log-rotation in the file writers to retain the last N-(duration-units) of (a) stream(s) - but consider that this could be implemented separately if local clients could receive a stream
0. Add a websocket server to let local clients receive streams (e.g. to permit a sophisticated stream recorder or image analyser to be written in a separate code-base)

## Internals

This internals diagram was drawn before implementation. At present time, the monitor does not exist, and logging is performed directly to stdout or file. However, the rest is about right.

![alt text][internals]

### Websockets

The code was initially developed with ```nhooyr/websocket``` then ```gobwas/ws``` then ```gorilla/websocket```. The usage model is for few high bandwidth connections so there is little advantage to be gained from the more recent re-implementations of websockets that focus on high connection counts with sparse activity.  


## Tests

### TODO - unit
0. Check that cmd.Process.Kill works on odroid / figure out a way to run the test suite on a different arch ....

### TODO - integration

0. Send a binary blob to the http endpoint of ```vw```, and compare with what is received at websocket server provisioned for this test.
0. Send sequentially two different binary blobs to the http endpoint of ```vw``` and forward to two websocket servers. Check both get two different blobs that match the originally sent blobs.
0. Send a MAXSIZE binary blob to the http endpoint and check it can be forwarded to a websocket server
0. Stream from a short file to ```vw``` configured to send to one server. That server saves the file to disk. Compare the files. 
0. Stream a file using ffmpeg and check the stats at the receiving websocket order.
0. Include a file that can be streamed with ffmpeg. Stream it to ```vw```, configured to send it to two servers. Those servers save the file to disk. Compare the files to ensure both obtained a complete stream.
0. Time stamp the arrival time of packets at the test websocket server, both with and without the write-to-file feature enabled.

### done - Unit

0. Send a binary blob to the http endpoint, record it to file, and compare with the send blob.
0. Send a sequence of binary blogs of non-monotonic size variation and check that they are received in the correct order at the websocket server


[status]: https://img.shields.io/badge/alpha-do%20not%20use-orange "Alpha status, do not use" 
[logo]: ./img/logo.png "VW logo"
[internals]: ./img/internals.png "Diagram of VW internals showing http server, websocket client, mux, monitor, and syscall for ffmpegs"

## Issues

Websockets do not reconnect, but should
Config file and command line integration is currently not fully implemented, and does not support all uses cases
Logging does not seem to log to file 
Port should be externally configurable for running capture commands separately
A number of parameters are fixed but could/should be made optionally configurable in the config file
Killing capture commands is not working on odroid
Configuration updates, such as sending streams to new destinations should potentially be supported without restart, so as to avoid unnecessary USB connection cycles
Some form of local preview would be useful
Some form of auditing of what has been seen, what has been transmitted, would be useful (e.g. log to file)
Tests need updating

killing ffmpeg occasionally does not work - it appears this may not be the first time this issue has been [seen](https://lists.ffmpeg.org/pipermail/ffmpeg-user/2017-February/035181.html) but there is no follow up to this 2012 issue to shed light on what was found then.