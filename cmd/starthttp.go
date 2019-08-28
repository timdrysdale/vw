package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func startHttp(closed <-chan struct{}, wg *sync.WaitGroup, listen url.URL, feedmap FeedMap) {
	defer wg.Done()

	port, err := strconv.Atoi(listen.Port())
	if err != nil {
		panic("Error Converting port into int")

	}

	wg.Add(1)
	fmt.Printf("\n Listening on :%d\n", port)
	srv := startHttpServer(closed, wg, port, feedmap)

	<-closed
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
	fmt.Printf("Exiting START HTTP SERVER %v\n", wg)
	return

} // startHttp

//mux := http.NewServeMux()
//mux.Handler("/request", requesthandler)
//http.ListenAndServe(":9000", nil)

func startHttpServer(closed <-chan struct{}, wg *sync.WaitGroup, port int, feedmap FeedMap) *http.Server {
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
				//fmt.Printf("\n%v\n", frame[:n])
				if err == nil && n > 0 {
					packet := Packet{Data: frame[:n]} //slice length is high-low
					chunkCount = chunkCount + (n / chunkSize)
					for _, channel := range channelSlice {
						channel <- packet
					} //for
					//reset buffer

					//lasti := 0
					//for i, val := range packet.Data {
					//	if val == 0x47 {
					//		fmt.Printf("http: %d spaced at %v\n", val, i-lasti)
					//		lasti = i
					//	}
					//}

					//for i, val := range frame[:n] {
					//	if val == 0x47 {
					//	fmt.Printf("%d %v\n", i, val)
					//	}
					//}

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
				_, err := io.ReadFull(r.Body, chunk)
				if err == nil {
					//	if n == 188 {
					_, _ = frameBuffer.Write(chunk)
					//fmt.Printf("\n%v\n", chunk)
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
