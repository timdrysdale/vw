package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/rwc"
)

//add/delete/deleteAll/show/showAll

func TestHandleStreamAdd(t *testing.T) {

	rule := []byte(`{"stream":"/stream/large","feeds":["audio","video0"]}`)

	req, err := http.NewRequest("PUT", "/api/streams", bytes.NewBuffer(rule))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	a := testApp()
	handler := http.HandlerFunc(a.handleStreamAdd)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := string(rule)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

	time.Sleep(time.Millisecond)

	feeds := []string{"audio", "video0"}
	isFound := []bool{false, false}

	if _, ok := a.Hub.Rules["/stream/large"]; ok {
		if len(feeds) != len(a.Hub.Rules["/stream/large"]) {
			t.Errorf("Number of feeds in rule is wrong")
		} else {

			for j := 0; j < len(feeds); j++ {
				for i := 0; i < len(feeds); i++ {
					if feeds[i] == a.Hub.Rules["/stream/large"][j] {
						isFound[i] = true
					}
				}
			}
		}
	} else {
		t.Errorf("Rule not added to Hub")
	}

	for i, val := range isFound {
		if !val {
			t.Errorf("Didn't find feed %s in rule", feeds[i])
		}
	}

}

func testApp() *App {
	a := &App{Hub: agg.New(), Closed: make(chan struct{})}
	a.Websocket = rwc.New(a.Hub)
	go a.Hub.Run(a.Closed)
	go a.Websocket.Run(a.Closed)
	return a
}
