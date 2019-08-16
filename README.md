# VW

![alt text][logo]

![alt text][status]

Video over websockets, written in golang as an alternative to the node.js-based relay in [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg).

## Why?
The current ```jsmpeg``` provides an insecure websocket server that clients can log into. One of the main use-cases for ```jsmpeg``` is when network permissions are so constrained that webRTC is not possible.  Since those network policy conditions also usually preclude offering an externally-visible server, then we need to do _something_ to join up the dots with an external user, such as websocket bridge to another server. I wrote the experimental [streamer](https://github.com/timdrysdale/streamer) to fulfil this role, but for some of the constrained targets on which I would like to deploy ```jsmpeg```, it would be attractive to avoid having to install node.js, and duplicate the handling of each message. On a c5.large AWS server, ```streamer``` (operating in server-server mode) uses approx 1% of CPU per single-in single-out video stream at 1000-1300kbps, and slightly less than 1% of memory (Intel Xeon 8100 series with 4GB RAM), including in both cases the overhead for nginx to manage the reverse-proxying (approx 3/10 of that 1% resource usage). That's a powerful CPU, so on a more constrained target, there will be a definite benefit to avoiding double handling, as well as approximately halving the installation size by not needing node.

## Usage

### configuration

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



[status]: https://img.shields.io/badge/alpha-do%20not%20use-orange "Alpha status, do not use" 
[logo]: ./img/logo.png "VW logo"
