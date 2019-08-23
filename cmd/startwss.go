package cmd

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

func startWss(closed <-chan struct{}, wg *sync.WaitGroup, clientMap ClientMap) {
	defer wg.Done()
	for url, channels := range clientMap {
		wg.Add(1)
		name := "wssClient(" + uuid.New().String()[:3] + "):"
		go wssClient(closed, wg, url, channels, name)
		log.Printf("%s spawned", name)
	}
}

func wssClient(closed <-chan struct{}, wg *sync.WaitGroup, url string, messageChannels []chan Packet, name string) {

	defer wg.Done()
	const minTimeout = 500 * time.Millisecond //TODO make configurable
	timeout := minTimeout

	for {
		log.Printf("%s dialing %s", name, url)
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), url)

		if err != nil {
			log.Printf("%s can not connect to %s: %v", name, url, err)
			time.Sleep(timeout)
			timeout = 2 * timeout //polynomial backoff
		} else {

			timeout = minTimeout //we've connected so reset timeout
			log.Printf("%s connected to %s", name, url)

			for {
				select {

				case <-closed:
					err = conn.Close()
					if err != nil {
						log.Printf("%s can not close: %v", name, err)
					} else {
						log.Printf("%s closed\n", name)
					}

				default:

					for _, channel := range messageChannels {
						select {
						case packet := <-channel:
							err = wsutil.WriteClientMessage(conn, ws.OpBinary, packet.Data)
							if err != nil {
								log.Printf("%s send error: %v", name, err)
							}
						default:
						}
					}
				}
			}
		}
	}
}
