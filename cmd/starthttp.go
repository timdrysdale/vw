package cmd

import (
	"bufio"
	"context"
	"fmt"
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
	chunkSize := 188         //4096                               //188                                //188
	//frameBufferArray := make([]byte, maxFrameBytes) //owned by buffer, don't re-use
	//frameBuffer := bytes.NewBuffer(frameBufferArray)

	var frameBuffer mutexBuffer
	//frameBuffer.mux.Lock()
	//frameBuffer.b = bytes.NewBuffer(frameBufferArray)
	//frameBuffer.mux.Unlock()

	rawFrame := make([]byte, maxFrameBytes) // use for reading from frameBuffer
	flushPeriod := 2 * time.Millisecond     //time.Duration(opts.FlushMS) * time.Millisecond
	tickerFlush := time.NewTicker(flushPeriod)
	syncTS := byte('G')

	frameBuffer.b.Reset() //else we send whole buffer on first flush

	reader := bufio.NewReader(r.Body)

	//tCh := make(chan int)

	go func() {
		for {
			//tCh <- 0
			glob, err := reader.ReadBytes(syncTS)
			if err == nil {
				frameBuffer.mux.Lock()
				_, err = frameBuffer.b.Write(glob)
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
		//case <-tCh:
		// keep waiting
		case <-tickerFlush.C: //<-time.After(flushPeriod):
			//flush buffer to internal send channel
			frameBuffer.mux.Lock()
			n, err := frameBuffer.b.Read(rawFrame)

			frame := rawFrame //append([]byte{syncTS}, rawFrame[:n]...)

			offsetFrequency := make(map[int]int) //map of frequency of implied offset of first sync
			for i, char := range frame[:n] {
				if char == syncTS {
					impliedOffset := i % chunkSize
					if val, ok := offsetFrequency[impliedOffset]; ok {
						offsetFrequency[impliedOffset] = val + 1
					} else {
						offsetFrequency[impliedOffset] = 1
					}
				}
			}
			estimatedOffset := 0
			estimatedFrequency := 0
			for offset, frequency := range offsetFrequency {
				if frequency > estimatedFrequency {
					estimatedFrequency = frequency
					estimatedOffset = offset
				}
			}

			//if estimatedOffset == 187 { //we're missing the first sync
			//
			//	outframe = append([]byte{syncTS}, frame[:n]...)
			//	estimatedOffset = 0
			//}

			potentiallyReady := n + 1 - estimatedOffset
			trimFromEnd := potentiallyReady % chunkSize

			forNextFrame := frame[(n + 1 - trimFromEnd):(n + 1)]
			forThisFrame := frame[estimatedOffset:(n + 1 - trimFromEnd)]
			//fmt.Printf("-------------------------------------------------------------------------------------------")
			//fmt.Printf("FrameSize: %d, skipped: %d, forNextFrame: %d", n, estimatedOffset, len(forNextFrame))
			//fmt.Printf("offset: %d, trim %d\n", estimatedOffset, trimFromEnd)
			//fmt.Printf("\n%v\n", offsetFrequency)
			//fmt.Printf("Chunks this write: %d\n", (n+1-estimatedOffset-trimFromEnd)/chunkSize)
			//fmt.Printf("First chars are: %v\n", frame[0])
			//fmt.Printf("Trimmed chars are: %v\n", forNextFrame)
			//check
			if len(forThisFrame)%188 != 0 {
				fmt.Printf("Frame length not have integral multiple of chunk size\n")
			}

			frameBuffer.b.Reset()
			frameBuffer.b.Write(forNextFrame)

			frameBuffer.mux.Unlock()

			if err == nil && n > 0 {
				msg := message{sender: myDetails, op: ws.OpBinary, data: forThisFrame}
				msgChan <- msg
			}

		case <-closed:
			fmt.Printf("\nMuxHandler got closed\n")
			return
		}
	}
}
