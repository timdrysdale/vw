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

	maxFrameBytes := 1024000                        //TODO make configurable
	chunkSize := 188                                //4096                               //188                                //188
	readBuf := make([]byte, chunkSize)              //maxFrameBytes)
	frameBufferArray := make([]byte, maxFrameBytes) //owned by buffer, don't re-use
	frameBuffer := bytes.NewBuffer(frameBufferArray)
	frame := make([]byte, maxFrameBytes) // use for reading from frameBuffer
	flushPeriod := 5 * time.Millisecond  //time.Duration(opts.FlushMS) * time.Millisecond
	tickerFlush := time.NewTicker(flushPeriod)

	chunkCount := 0
	frameBuffer.Reset() //else we send whole buffer on first flush

	for {
		select {
		case <-tickerFlush.C:
			//flush buffer to internal send channel

			n, err := frameBuffer.Read(frame)

			if err == nil && n > 0 {
				chunkCount = chunkCount + (n / chunkSize)
				msg := message{sender: myDetails, op: ws.OpBinary, data: frame[:n]}
				msgChan <- msg
				frameBuffer.Reset()
			}

		case <-closed:
			fmt.Printf("\nMuxHandler got closed\n")
			return

		default:
			// get a chunk from the response body
			//n, err := io.ReadAtLeast(r.Body, readBuf, chunkSize)
			n, err := io.ReadFull(r.Body, readBuf)
			if err == nil {
				_, _ = frameBuffer.Write(readBuf[:n]) //readBuf[:n])

			}
		}
	}
}
