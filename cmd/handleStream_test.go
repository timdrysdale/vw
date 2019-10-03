package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/rwc"
)

//add/delete/deleteAll/show/showAll

func TestHandleStreamAddv1(t *testing.T) {

	rule := []byte(`{"stream":"/stream/large","feeds":["audio","video0"]}`)

	req, err := http.NewRequest("PUT", "/api/streams", bytes.NewBuffer(rule))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	a := testApp(true)
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

func TestHandleStreamAdd(t *testing.T) {

	rule := []byte(`{"stream":"/stream/large","feeds":["audio","video0"]}`)

	req, err := http.NewRequest("PUT", "/api/streams", bytes.NewBuffer(rule))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.HandlerFunc(a.handleStreamAdd)

	go func() {
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
	}()

	got := <-a.Hub.Add

	if got.Stream != "/stream/large" {
		t.Error("Wrong stream")
	}

	if got.Feeds[0] != "audio" {
		t.Error("Wrong feeds")
	}
	if got.Feeds[1] != "video0" {
		t.Error("Wrong feeds")
	}

}

func TestHandleStreamDelete(t *testing.T) {

	req, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		t.Error(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"stream": "video0",
	})

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.HandlerFunc(a.handleStreamDelete)

	go func() {
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	}()

	got := <-a.Hub.Delete

	if got != "video0" {
		t.Error("Wrong stream")
	}

}

func TestHandleStreamDeleteAll(t *testing.T) {

	req, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		t.Error(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"stream": "all",
	})

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.HandlerFunc(a.handleStreamDeleteAll)

	go func() {
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	}()

	got := <-a.Hub.Delete

	if got != "deleteAll" {
		t.Errorf("handler send wrong message on Hub.Delete: got %v want %v",
			got, "deleteAll")
	}

}

func TestHandleStreamShow(t *testing.T) {

	req, err := http.NewRequest("PUT", "", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{
		"stream": "stream/large",
	})

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.HandlerFunc(a.handleStreamShow)

	a.Hub.Rules = make(map[string][]string)
	a.Hub.Rules["stream/large"] = []string{"audio", "video0"}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := "[\"audio\",\"video0\"]"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

}

func testApp(running bool) *App {
	a := &App{Hub: agg.New(), Closed: make(chan struct{})}
	a.Websocket = rwc.New(a.Hub)
	if running {
		go a.Hub.Run(a.Closed)
		go a.Websocket.Run(a.Closed)
	}
	return a
}
