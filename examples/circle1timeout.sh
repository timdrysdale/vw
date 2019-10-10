#!/bin/sh 
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","feeds":["video0","audio0"]}' http://localhost:8888/api/streams
for (( ; ; ))
do
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://video.practable.io:444/in/video0","id":"0"}' http://localhost:8888/api/destinations
sleep 5s
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://video.practable.io:444/in/video1","id":"0"}' http://localhost:8888/api/destinations
sleep 5s
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://video.practable.io:444/in/video2","id":"0"}' http://localhost:8888/api/destinations
sleep 5s
curl -X POST -H "Content-Type: application/json" -d '{"stream":"/stream/front/large","destination":"wss://video.practable.io:444/in/video3","id":"0"}' http://localhost:8888/api/destinations  
sleep 5s
done
