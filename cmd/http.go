package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (app *App) startHttp() {
	app.WaitGroup.Add(1)
	defer app.WaitGroup.Done()

	log.WithField("port", app.Opts.Port).Debug("http.Server listening port set")

	srv := app.startHttpServer(app.Opts.Port)

	log.Debug("Started http.Server")

	<-app.Closed // wait for shutdown

	log.Debug("Starting to close http.Server")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.WithField("error", err).Fatal("Failure/timeout shutting down the http.Server gracefully")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(app.Opts.HttpWaitMs)*time.Millisecond)
	defer cancel()

	srv.SetKeepAlivesEnabled(false)
	if err := srv.Shutdown(ctx); err != nil {
		log.WithField("error", err).Fatal("Could not gracefully shutdown http.Server")
	}

	log.Debug("Stopped http.Server")

	return
} // startHttp

func (app *App) startHttpServer(port int) *http.Server {
	app.WaitGroup.Add(1)
	defer app.WaitGroup.Done()

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{Addr: addr}

	var router = mux.NewRouter()

	router.HandleFunc("/api", app.handleApi)
	router.HandleFunc("/api/streams/{stream}", app.handleStreamShow).Methods("GET")
	router.HandleFunc("/api/streams/{stream}", app.handleStreamAdd).Methods("PUT", "UPDATE")
	router.HandleFunc("/api/streams/{stream}", app.handleStreamDelete).Methods("DELETE")
	router.HandleFunc("/ts", app.handleTs)
	router.HandleFunc("/ws", app.handleWs)

	/*	router.HandleFunc("/api/destinations/{id}/create", handleCreateDestination).Methods("POST")
		router.HandleFunc("/api/destinations/{id}/delete", handleDeleteDestination).Methods
		router.HandleFunc("/api/destinations/{id}/delete", handleDeleteDestination)

		http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) { apiHandler(w, r, a) })

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { indexHandler(closed, w, r, a.Hub) })
	*/
	app.WaitGroup.Add(1)
	go func() {
		defer app.WaitGroup.Done()

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
