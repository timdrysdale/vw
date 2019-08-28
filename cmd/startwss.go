package cmd

import (
	"context"
	"fmt"
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
	const timeout = 1000 * time.Millisecond //TODO make configurable

	for {
		fmt.Printf("%s dialing %s\n", name, url) //TODO revert to log
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		conn, _, _, err := ws.DefaultDialer.Dial(ctx, url)

		fmt.Printf("%s dialed %s getting %v\n", name, url, err)

		if err != nil {

			log.Printf("%s can not connect to %s: %v\n", name, url, err)
			select {
			case <-time.After(timeout):
			case <-closed:
				fmt.Printf("wssClient detected closed\n")
				err = conn.Close()
				if err != nil {
					log.Printf("%s can not close: %v", name, err)
				} else {
					log.Printf("%s closed\n", name)
				}
				fmt.Printf("wssClient has closed\n")
				return
			}

		} else {

			log.Printf("%s connected to %s\n", name, url)

			for {
				select {

				case <-closed:
					fmt.Printf("wssClient detected closed\n")
					err = conn.Close()
					if err != nil {
						log.Printf("%s can not close: %v", name, err)
					} else {
						log.Printf("%s closed\n", name)
					}
					fmt.Printf("wssClient has closed\n")
					return
				default:
					//fmt.Printf("%s ready to start sending messages\n", name)
					for _, channel := range messageChannels {
						select {
						case packet := <-channel:
							//lasti := 0
							//for i, val := range packet.Data {
							//	if val == 0x47 {
							//		lasti = i
							//		break
							//	}
							//}
							//for i := lasti; i < len(packet.Data); i = i + 188 {
							//	if packet.Data[i] != 0x47 {
							//		fmt.Printf("\nSync not found at %d, neighbouring chars are %v\n", i, packet.Data[i-5:i+5])
							//	}
							//}

							err = wsutil.WriteClientMessage(conn, ws.OpBinary, packet.Data)
							//fmt.Printf("\n%s sent %d bytes\n", name, len(packet.Data))
							if err != nil {
								log.Printf("%s send error: %v", name, err)
							}
						case <-closed:
							fmt.Printf("wssClient detected closed\n")
							err = conn.Close()
							if err != nil {
								log.Printf("%s can not close: %v", name, err)
							} else {
								log.Printf("%s closed\n", name)
							}
							fmt.Printf("wssClient has closed\n")
						}
					}
				}
			}
		}
	}
}
