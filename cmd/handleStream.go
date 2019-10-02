package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *App) handleStreamAllShow(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("StreamAllShow")
}

func (app *App) handleStreamShow(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fmt.Sprintf("StreamShow for %s", r.URL.Path))
}

func (app *App) handleStreamAdd(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fmt.Sprintf("StreamAdd for %s", r.URL.Path))
}

func (app *App) handleStreamDelete(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(fmt.Sprintf("StreamDelete for %s", r.URL.Path))

}
