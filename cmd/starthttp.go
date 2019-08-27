package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

func startHTTP(closed <-chan struct{}, wg *sync.WaitGroup, listen url.URL, feedmap FeedMap) {
	defer wg.Done()

	port, err := strconv.Atoi(listen.Port())
	if err != nil {
		panic("Error Converting port into int")

	}

	wg.Add(1)
	fmt.Printf("\n Listening on :%d\n", port)
	srv := startHTTPServer(closed, wg, port, feedmap)

	for {
		select {
		case <-closed:
			fmt.Printf("Starting to close HTTP SERVER %v\n", wg)
			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Panicf("failure/timeout shutting down the http server gracefully: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO make configurable
			defer cancel()

			srv.SetKeepAlivesEnabled(false)
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
			}
			if os.Getenv("DEBUG") == "true" {
				fmt.Printf("\nExiting startHTTP %v\n", wg)
			}
			return
		default:
		}
	}
}

func startHTTPServer(closed <-chan struct{}, wg *sync.WaitGroup, port int, feedmap FeedMap) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { muxingHandler(closed, w, r, feedmap) })

	wg.Add(1)
	go func() {
		defer wg.Done()

		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
			log.Fatalf("ListenAndServe(): %s", err)
		}
		fmt.Printf("Exiting HTTPServer %v\n", wg)
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func muxingHandler(closed <-chan struct{}, w http.ResponseWriter, r *http.Request, feedmap FeedMap) {

	if channelSlice, ok := feedmap[r.URL.Path]; !ok {
		fmt.Printf(`\n
---------------------------------
Unknown stream, check config for:%s\n
--------------------------------\n`, r.URL.Path)
		return
	} else {
		//receive MPEGTS in 188 byte chunks
		//ffmpeg uses one tcp packet per frame

		maxFrameBytes := 1024000 //TODO make configurable
		chunkSize := 4096        //188
		chunk := make([]byte, chunkSize)
		frameBufferArray := make([]byte, maxFrameBytes) //owned by buffer, don't re-use
		frameBuffer := bytes.NewBuffer(frameBufferArray)
		frame := make([]byte, maxFrameBytes) // use for reading from frameBuffer
		flushPeriod := 5 * time.Millisecond  //TODO make configurable
		//statsPeriod := 1 * time.Second       //TODO make configurable
		tickerFlush := time.NewTicker(flushPeriod)
		//tickerStats := time.NewTicker(statsPeriod)
		chunkCount := 0
		frameBuffer.Reset() //else we send whole buffer on first flush
		//mute := true

		for {
			select {
			case <-time.After(1 * time.Second):
				//frameBuffer.Reset() //flush first second of video with any non-188 byte aligned header
				//mute = false
			case <-tickerFlush.C:

				//flush buffer to internal send channel
				n, err := frameBuffer.Read(frame)

				if err == nil && n > 0 {
					packet := Packet{Data: frame[:n]} //slice length is high-low
					chunkCount = chunkCount + (n / chunkSize)
					for _, channel := range channelSlice {
						channel <- packet
					} //for
					frameBuffer.Reset()
				} else {
					//fmt.Printf("\nFrame buffer read error %v\n", err)
				}

			//case <-tickerStats.C:
			// print stats
			//fmt.Printf("\n Chunks total: %d\n", chunkCount)
			case <-closed:
				fmt.Printf("\nMuxHandler got closed\n")
				return
			default:
				// get a chunk from the response body
				n, err := io.ReadFull(r.Body, chunk)
				if err == nil {
					//	if n == 188 {
					_, _ = frameBuffer.Write(chunk)
					//fmt.Printf("\n%v\n", chunk)
				}
				if n < chunkSize {
					time.Sleep(1 * time.Millisecond) //reduce CPU? nope.
				}
				//		if err != nil {
				//			fmt.Printf("\nFailed to write chunk to frameBuffer;' only wrote %d because %v\n", n, err)
				//		}

				//	} else {
				//		fmt.Printf("\nBad chunk, wanted %d, got %d\n", chunkSize, n)
				//	} //if/else
				//} //if
			} //select
		}
	}
}

/*
func HandleConnections(closed <-chan struct{}, wg *sync.WaitGroup, clientActionsChan chan clientAction, messagesFromMe chan message, host *url.URL)

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var HTTPUpgrader ws.HTTPUpgrader

		// chrome needs this to connect
		HTTPUpgrader.Protocol = func(str string) bool { return true }

		conn, _, _, err := HTTPUpgrader.Upgrade(r, w)

		if err != nil {
			log.Fatalf("WS upgrade failed because %v\n", err)
			return
		}

		//subscribe this new client
		messagesForMe := make(chan message, 2)
		var name = uuid.New().String()
		var topic = r.URL.Path

		client := clientDetails{name: name, topic: topic, messagesChan: messagesForMe}

		clientActionsChan <- clientAction{action: clientAdd, client: client}

		defer func() {
			clientActionsChan <- clientAction{action: clientDelete, client: client}
			fmt.Printf("Disconnected %v, deleting from topics\n", client)
		}()

		var localWG sync.WaitGroup

		localWG.Add(2)
		// read from client
		go func() {
			defer localWG.Done()
			for {

				msg, op, err := wsutil.ReadClientData(conn)
				if err == nil {
					messagesFromMe <- message{sender: client, op: op, data: msg}
				} else {
					log.Printf("Error on read because %v\n", err)
					return
				}
			}
		}()

		//write to client
		go func() {
			defer conn.Close()
			defer localWG.Done()
			for {
				select {
				case msg := <-messagesForMe:
					err = wsutil.WriteServerMessage(conn, msg.op, msg.data)
					if err != nil {
						log.Printf("Fatal error on write because %v", err)
						return
					}

				case <-closed:
					return
				} //select
			} //for
		}() //func

		localWG.Wait()
	}) //end of fun definition
*/
