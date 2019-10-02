package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/rwc"
)

func (app *App) handleStreamShowAll(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("StreamAllShow")

}

func (app *App) handleStreamShow(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fmt.Sprintf("StreamShow for %s", r.URL.Path))
}

/*  Add a new stream rule

Example:

curl -X POST -H "Content-Type: application/json" \
-d '{"stream":"/stream/front/large","feeds":["video0","audio0"]}'\
http://localhost:8888/api/streams/video

*/
func (app *App) handleStreamAdd(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var rule agg.Rule
	err = json.Unmarshal(b, &rule)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	app.Hub.Add <- rule

	output, err := json.Marshal(rule)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func (app *App) handleStreamDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stream := vars["stream"]

	app.Hub.Delete <- stream

	output, err := json.Marshal(stream)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func (app *App) handleStreamDeleteAll(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fmt.Sprintf("StreamDelete for %s", r.URL.Path))

}

// Destination

// curl -X GET http://localhost:8888/api/destinations/all
func (app *App) handleDestinationShowAll(w http.ResponseWriter, r *http.Request) {

	output, err := json.Marshal(app.Websocket.Rules)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

// curl -X GET http://localhost:8888/api/destinations/01
func (app *App) handleDestinationShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	output, err := json.Marshal(app.Websocket.Rules[id])
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)

}

/*  Add a new stream rule

Example:

curl -X POST -H "Content-Type: application/json" \
-d '{"stream":"/stream/front/large","feeds":["video0","audio0"]}'\
http://localhost:8888/api/streams/video

*/
func (app *App) handleDestinationAdd(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var rule rwc.Rule
	err = json.Unmarshal(b, &rule)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	app.Websocket.Add <- rule

	output, err := json.Marshal(rule)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

// curl -X DELETE http://localhost:8888/api/destinations/00
func (app *App) handleDestinationDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	app.Websocket.Delete <- id

	output, err := json.Marshal(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}

func (app *App) handleDestinationDeleteAll(w http.ResponseWriter, r *http.Request) {

	id := "deleteAll"

	app.Websocket.Delete <- id

	output, err := json.Marshal(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
}
