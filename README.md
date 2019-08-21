# VW

![alt text][logo]

![alt text][status]

Video over websockets, written in golang as an alternative to the node.js-based relay in [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg).

## Why?
Mainly for my own amusement. The current ```jsmpeg``` provides an websocket server that a browser client can log into to get a stream of MPEGTS video. My use-case for```jsmpeg``` is for networks too restrictive to support WebRTC, which usually also precludes hosting a server that is visible to external users outside the firewall. However, if we use switch from using a websocket server to using a websocket client then we can connect to an external server. It'd be possible to tweak the ```node.js``` demo to achieve this, but frankly, I wanted a practice project for ```go``` and it might yield some deployment convenience and a small performance bump. Related ```golang``` projects include [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) which calls itself [websocketd](https://github.com/joewalnes/websocketd) for TCP instead of ```STDIN``` and ```STDOUT```, but [ws-tcp-relay](https://github.com/isobit/ws-tcp-relay) has a websocket server, not a client. 

## Usage

     TODO - put help message here

      --listen
      --forward
      --control
      --video_cmd (may need to include v4l2-ctl to set camera options)
      --audio_cmd (may need to include alsa to set audio options)



## Installation 

Download and compile the repo 

    $ go get github.com/timdrysdale/vw
    $ cd $GO_PATH/src/github.com/timdrysdale/vw
    $ go get ./...
    $ go build


## Internals

![alt text][internals]


                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             


## Configuration

sysdefault:CARD=C920


whether to record stream to disk
where to record stream to disk
how much stream to save, and whether that is in frames, or 

## Internals

The websocket client is implemented using ```nhooyr/websocket``` because it is more recent than ```gorilla/websocket```, and has a relatively straightforward, idomatic API.


writing to disk is accomplished in a separate go-routine to avoid introducing variable latency in the network messaging.

## Tests

TODO

Send a binary blob to the http endpoint, record it to file, and compare with the send blob.

Send a binary blob to the http endpoint of ```vw```, and compare with what is received at websocket server provisioned for this test.

Send sequentially two different binary blobs to the http endpoint of ```vw``` and forward to two websocket servers. Check both get two different blobs that match the originally sent blobs.

Send a MAXSIZE binary blob to the http endpoint and check it can be forwarded to a websocket server

Send a sequence of binary blogs of non-monotonic size variation and check that they are received in the correct order at the websocket server

Stream a file using ffmpeg and check the stats at the receiving websocket order.

Stream from a short file to ```vw``` configured to send to one server. That server saves the file to disk. Compare the files. 

Include a file that can be streamed with ffmpeg. Stream it to ```vw```, configured to send it to two servers. Those servers save the file to disk. Compare the files to ensure both obtained a complete stream.

Time stamp the arrival time of packets at the test websocket server, both with and without the write-to-file feature enabled.


## Device configuration

There's a description of using alsa and pulseaudio with docker [here](https://github.com/mviereck/x11docker/wiki/Container-sound:-ALSA-or-Pulseaudio).


[status]: https://img.shields.io/badge/alpha-do%20not%20use-orange "Alpha status, do not use" 
[logo]: ./img/logo.png "VW logo"
[internals]: ./img/internals.png "Diagram of VW internals showing http server, websocket client, mux, monitor, and syscall for ffmpegs"