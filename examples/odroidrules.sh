#!/bin/sh 
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/odroid0","feeds":["video0","audio0"]}' http://localhost:8888/api/streams
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/odroid0","destination":"wss://video.practable.io:443/in/odroid0","id":"0"}' http://localhost:8888/api/destinations
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/odroid1","feeds":["video1","audio1"]}' http://localhost:8888/api/streams
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/odroid1","destination":"wss://video.practable.io:443/in/odroid1","id":"1"}' http://localhost:8888/api/destinations
