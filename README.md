# VW

![alt text][logo]

![alt text][status]

Video over websockets, written in golang as an alternative to the node.js-based relay in [phoboslab/jsmpeg](https://github.com/phoboslabs/msjpeg).

## Why?
The current ```jsmpeg``` provides an insecure websocket server that clients can log into. One of the main use-cases for ```jsmpeg``` is when network permissions are so constrained that webRTC is not possible.  Since those network policy conditions also usually preclude offering an externally-visible server, then we need to do _something_ to join up the dots with an external user, such as websocket bridge to another server. I wrote the experimental [streamer](https://github.com/timdrysdale/streamer) to fulfil this role, but for some of the constrained targets on which I would like to deploy ```jsmpeg```, it would be attractive to avoid having to install node.js, and duplicate the handling of each message. On a c5.large AWS server, ```streamer``` (operating in server-server mode) uses approx 1% of CPU per single-in single-out video stream at 1000-1300kbps, and slightly less than 1% of memory (Intel Xeon 8100 series with 4GB RAM), including in both cases the overhead for nginx to manage the reverse-proxying (approx 3/10 of that 1% resource usage). That's a powerful CPU, so on a more constrained target, there will be a definite benefit to avoiding double handling, as well as approximately halving the installation size by not needing node.




[status]: https://img.shields.io/badge/alpha-do%20not%20use-orange "Alpha status, do not use" 
[logo]: ./img/logo.png "VW logo"
