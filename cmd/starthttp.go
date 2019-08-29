package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/google/uuid"
)

func startHttp(closed <-chan struct{}, wg *sync.WaitGroup, listen url.URL, opts HTTPOptions, msgChan chan message) {
	defer wg.Done()

	port, err := strconv.Atoi(listen.Port())
	if err != nil {
		panic("Error Converting port into int")
	}

	wg.Add(1)
	fmt.Printf("\n Listening on :%d\n", port)
	srv := startHttpServer(closed, wg, port, opts, msgChan)

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

func startHttpServer(closed <-chan struct{}, wg *sync.WaitGroup, port int, opts HTTPOptions, msgChan chan message) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { muxingHandler(closed, w, r, opts, msgChan) })

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

func muxingHandler(closed <-chan struct{}, w http.ResponseWriter, r *http.Request, opts HTTPOptions, msgChan chan message) {
	myDetails := clientDetails{uuid.New().String()[:3], r.URL.Path, make(chan message)}

	//receive MPEGTS in 188 byte chunks
	//ffmpeg uses one tcp packet per frame

	maxFrameBytes := 1024000 //TODO make configurable

	var frameBuffer mutexBuffer

	rawFrame := make([]byte, maxFrameBytes)

	glob := make([]byte, maxFrameBytes)

	frameBuffer.b.Reset() //else we send whole buffer on first flush

	reader := bufio.NewReader(r.Body)

	tCh := make(chan int)

	//Method for detecting packet boundaries: identify empty buffer via delay on reading a byte
	//after 13.738µs got 188 bytes
	//after 13.027µs got 120 bytes
	//after 13.883µs got 68 bytes
	//after 9.027µs got 188 bytes
	//after 8.876µs got 188 bytes
	//after 9.027µs got 104 bytes
	//<ffmpeg frame reported>
	//after 42.418638ms got 84 bytes  <============= NOTE THE ~40ms delay=====================
	//after 87.442µs got 188 bytes
	//after 43.555µs got 167 bytes
	//after 44.251µs got 21 bytes
	//after 23.267µs got 101 bytes
	//after 23.976µs got 49 bytes

	// Read from the buffer, blocking if empty
	go func() {

		for {

			tCh <- 0 //kick the monitoring routine

			n, err := io.ReadAtLeast(reader, glob, 1)

			if err == nil {

				frameBuffer.mux.Lock()

				_, err = frameBuffer.b.Write(glob[:n])

				frameBuffer.mux.Unlock()

				if err != nil {
					log.Fatalf("%v", err)
					return
				}

			} else {

				select {

				case <-closed:

					return //game over if closed

				default:

					// try again in case it's recoverable
				}
			}
		}
	}()

	for {

		select {

		case <-tCh:

			// do nothing, just received data from buffer

		case <-time.After(1 * time.Millisecond):
			// no new data for >= 1mS weakly implies frame has been fully sent to us
			// this is two orders of magnitude more delay than when reading from
			// non-empty buffer so _should_ be ok, but recheck if errors crop up on
			// lower powered system. Assume am on same computer as capture routine

			//flush buffer to internal send channel
			frameBuffer.mux.Lock()

			n, err := frameBuffer.b.Read(rawFrame)

			frame := rawFrame[:n]

			frameBuffer.b.Reset()

			frameBuffer.mux.Unlock()

			if err == nil && n > 0 {
				msg := message{sender: myDetails, op: ws.OpBinary, data: frame}
				msgChan <- msg
			}

		case <-closed:
			fmt.Printf("\nMuxHandler got closed\n")
			return
		}
	}
}
