package cmd

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/agg"
)

func startHttp(closed <-chan struct{}, wg *sync.WaitGroup, opts HTTPOptions, h *agg.Hub, running chan struct{}) {
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

func startHttpServer(closed <-chan struct{}, wg *sync.WaitGroup, port int, opts HTTPOptions, h *agg.Hub) *http.Server {
	defer wg.Done()
	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	http.HandleFunc("/ts", func(w http.ResponseWriter, r *http.Request) { tsHandler(closed, w, r, h) })
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { wsHandler(closed, w, r, h) })
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
