ffmpeg -re -i sample00.ts -f mpegts -codec:v mpeg1video -s 640x480 -b:v 1000k -r 24 -bf 0 http://127.0.0.1:35708/video0
