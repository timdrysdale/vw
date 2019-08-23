package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
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
	srv := startHttpServer(wg, port, feedmap)

	for {
		select {
		case <-closed:

			if err := srv.Shutdown(context.TODO()); err != nil {
				log.Panicf("failure/timeout shutting down the http server gracefully: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //TODO make configurable
			defer cancel()

			srv.SetKeepAlivesEnabled(false)
			if err := srv.Shutdown(ctx); err != nil {
				log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
			}
			return
		default:
		} // select
	} // for
} // startHttp

func startHttpServer(wg *sync.WaitGroup, port int, feedmap FeedMap) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { muxingHandler(w, r, feedmap) })

	wg.Add(1)
	go func() {
		defer wg.Done()

		// returns ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			//https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
			log.Fatalf("ListenAndServe(): %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func muxingHandler(w http.ResponseWriter, r *http.Request, feedmap FeedMap) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("request", err)
	}
	packet := Packet{Data: buf}

	if channelSlice, ok := feedmap[r.URL.Path]; ok {
		for _, channel := range channelSlice {
			channel <- packet
		}
	} else {
		fmt.Printf("didn't find %s in feedmap", r.URL.Path)
	}
}
