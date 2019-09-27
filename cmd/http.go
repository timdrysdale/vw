package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/hub"
)

func startHttp(closed <-chan struct{}, wg *sync.WaitGroup, opts HTTPOptions, h *hub.Hub, running chan struct{}) {
	defer wg.Done()

	wg.Add(1)

	log.WithField("port", opts.Port).Debug("http.Server listening port set")

	srv := startHttpServer(closed, wg, opts.Port, opts, h)

	close(running) //signal that we're running

	log.Debug("Started http.Server")

	<-closed // wait for shutdown

	log.Debug("Starting to close http.Server")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.WithField("error", err).Fatal("Failure/timeout shutting down the http.Server gracefully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO make configurable
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		log.WithField("error", err).Fatal("Could not gracefully shutdown http.Server")
	}

	log.Debug("Stopped http.Server")

	return
} // startHttp

func startHttpServer(closed <-chan struct{}, wg *sync.WaitGroup, port int, opts HTTPOptions, h *hub.Hub) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/ts", func(w http.ResponseWriter, r *http.Request) { tsHandler(closed, w, r, opts, h) })
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { wsHandler(closed, w, r, opts, h) })
	//http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) { apiHandler(closed, w, r, opts, h) })

	wg.Add(1)
	go func() {
		defer wg.Done()

		//https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithField("error", err).Fatal("http.ListenAndServe")
		}
		log.Debug("Exiting http.Server")
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func tsHandler(closed <-chan struct{}, w http.ResponseWriter, r *http.Request, opts HTTPOptions, h *hub.Hub) {

	topic := strings.TrimPrefix(r.URL.Path, "/") //trim separately because net does not guarantee leading /
	topic = strings.TrimPrefix(topic, "ts")      //strip ts because we're agnostic to which handler gets the feed
	name := uuid.New().String()[:3]
	myDetails := &hub.Client{Hub: h,
		Name:  name,
		Send:  make(chan hub.Message),
		Stats: hub.NewClientStats(),
		Topic: topic}

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
					log.Errorf("%v", err) //was Fatal?
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
				msg := hub.Message{Sender: *myDetails, Type: int(ws.OpBinary), Data: frame, Sent: time.Now()}
				h.Broadcast <- msg
			}

		case <-closed:
			log.WithFields(log.Fields{"Name": name, "Topic": topic}).Info("http.muxHandler closed")
			return
		}
	}
}
