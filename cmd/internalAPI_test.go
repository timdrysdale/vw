package cmd

import (
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/timdrysdale/agg"
	"github.com/timdrysdale/hub"
	"github.com/timdrysdale/rwc"
)

// Commands that we are testing ...
// {"verb":"add","what":"destination","rule":{"stream":"video0","destination":"wss://<some.relay.server>/in/video0","id":"0"}}
// {"verb":"delete","what":"destination","which":"<id>">}
// {"verb":"list","what":"destination","which":"all"}
// {"verb":"list","what":"destination","which":"<id>">}

// {"verb":"add","what":"stream","rule":{"stream":"video0","feeds":["video0","audio0"]}}
// {"verb":"delete","what":"stream","which":"<which>"}
// {"verb":"list","what":"stream","which":"all"}
// {"verb":"list","what":"stream","which":"<name>"}

// do one test with the internalAPI to check it is wired up ok, then
// test the handler directly for the rest of the tests
func TestInternalAPICommunicates(t *testing.T) {

	app = App{Hub: agg.New(), Closed: make(chan struct{})}
	app.Websocket = rwc.New(app.Hub)

	name := "api"
	go app.internalAPI(name)

	client, ok := <-app.Hub.Register

	if !ok {
		t.Errorf("Problem receiving internalAPI registration")
	}

	if client.Topic != name {
		t.Errorf("internalAPI registered with wrong name (%s/%s)\n", name, client.Topic)
	}

	cmd := []byte(`{"verb":"list","what":"destination","which":"all"}`)

	go func() {
		client.Send <- hub.Message{Sender: hub.Client{}, Data: cmd, Type: websocket.TextMessage, Sent: time.Now()}
	}()

	time.Sleep(1 * time.Millisecond)

	select {
	case msg, ok := <-client.Hub.Broadcast:
		if ok {
			if string(msg.Data) != "{}" {
				t.Error("Unexpected reply from internalAPI")
			}
		} else {
			t.Error("Problem with messaging channel")
		}

	case <-time.After(1 * time.Millisecond):
		t.Error("timeout waiting for internalAPI to reply")
	}

	close(app.Closed)
}

func TestInternalAPIBadCommand(t *testing.T) {

	a := testApp(false)

	cmd := []byte(`Not even JSON`)

	// note prefix / on stream is removed
	expected := errBadCommand

	_, err := a.handleAdminMessage(cmd)
	if err == nil {
		t.Error("Failed to throw error")
	} else if !reflect.DeepEqual(expected, err) {
		t.Errorf("Got wrong err %s/%s\n", expected, err)
	}

}

func TestInternalAPIDestinationAdd(t *testing.T) {

	a := testApp(false)

	rule := `{"id":"00","stream":"/stream/large","destination":"wss://video.practable.io:443/large"}`
	cmd := []byte(`{"verb":"add","what":"destination","rule":` + rule + `}`)

	// note prefix / on stream is removed
	expected := []byte(`{"id":"00","stream":"stream/large","destination":"wss://video.practable.io:443/large"}`)

	go func() {
		reply, err := a.handleAdminMessage(cmd)
		if err != nil {
			t.Error("unexpected error")
			return
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Got wrong rule %s/%s\n", expected, reply)
		}
	}()

	got := <-a.Websocket.Add

	if got.Stream != "stream/large" {
		t.Error("Wrong stream")
	}
	if got.Destination != "wss://video.practable.io:443/large" {
		t.Error("Wrong destination")
	}
	if got.Id != "00" {
		t.Error("Wrong Id")
	}
}

func TestInternalAPIDestinationDelete(t *testing.T) {

	a := testApp(false)

	cmd := []byte(`{"verb":"delete","what":"destination","which":"00"}`)

	// note prefix / on stream is removed
	expected := []byte(`{"deleted":"00"}`)

	go func() {
		reply, err := a.handleAdminMessage(cmd)
		if err != nil {
			t.Error("unexpected error")
			return
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Got wrong rule %s/%s\n", expected, reply)
		}
	}()

	got := <-a.Websocket.Delete

	if got != "00" {
		t.Error("Wrong Id")
	}
}

func TestInternalAPIDestinationDeleteAll(t *testing.T) {

	a := testApp(false)

	cmd := []byte(`{"verb":"delete","what":"destination","which":"all"}`)

	// note prefix / on stream is removed
	expected := []byte(`{"deleted":"deleteAll"}`)

	go func() {
		reply, err := a.handleAdminMessage(cmd)
		if err != nil {
			t.Error("unexpected error")
			return
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Got wrong rule %s/%s\n", expected, reply)
		}
	}()

	got := <-a.Websocket.Delete

	if got != "deleteAll" {
		t.Error("Wrong Id")
	}
}

func TestInternalAPIDestinationShow(t *testing.T) {

	a := testApp(false)
	a.Websocket.Rules = make(map[string]rwc.Rule)
	a.Websocket.Rules["00"] = rwc.Rule{Destination: "wss://video.practable.io:443/large", Stream: "stream/large", Id: "00"}

	cmd := []byte(`{"verb":"list","what":"destination","which":"00"}`)
	expected := []byte(`{"id":"00","stream":"stream/large","destination":"wss://video.practable.io:443/large"}`)

	reply, err := a.handleAdminMessage(cmd)
	if err != nil {
		t.Error("unexpected error")
		return
	}
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Got wrong rule %s/%s\n", expected, reply)
	}

}

func TestInternalAPIDestinationShowAll(t *testing.T) {

	a := testApp(false)
	a.Websocket.Rules = make(map[string]rwc.Rule)
	a.Websocket.Rules["stream/large"] = rwc.Rule{Stream: "stream/large",
		Destination: "wss://somewhere",
		Id:          "00"}
	a.Websocket.Rules["stream/medium"] = rwc.Rule{Stream: "stream/medium",
		Destination: "wss://overthere",
		Id:          "01"}

	cmd := []byte(`{"verb":"list","what":"destination","which":"all"}`)
	expected := []byte(`{"stream/large":{"id":"00","stream":"stream/large","destination":"wss://somewhere"},"stream/medium":{"id":"01","stream":"stream/medium","destination":"wss://overthere"}}`)

	reply, err := a.handleAdminMessage(cmd)
	if err != nil {
		t.Error("unexpected error")
		return
	}
	if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Got wrong rule %s/%s\n", expected, reply)
	}

}

/*



// These tests do not start the hub or the websocket client
// Their channels can be read by the test code, saving mocking
// and simpler than inspecting the side effects of a running
// Hub and Websocket

func TestInternalAPIStreamAdd(t *testing.T) {

	rule := []byte(`{"stream":"/stream/large","feeds":["audio","video0"]}`)

	req, err := http.NewRequest("PUT", "/api/streams", bytes.NewBuffer(rule))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.InternalAPIrFunc(a.handleStreamAdd)

	go func() {
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		//note prefix / on stream is removed
		expected := `{"stream":"stream/large","feeds":["audio","video0"]}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	}()

	got := <-a.Hub.Add

	if got.Stream != "stream/large" {
		t.Error("Wrong stream")
	}

	if got.Feeds[0] != "audio" {
		t.Error("Wrong feeds")
	}
	if got.Feeds[1] != "video0" {
		t.Error("Wrong feeds")
	}

}

func TestInternalAPIStreamDelete(t *testing.T) {

	req, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		t.Error(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"stream": "video0",
	})

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.InternalAPIrFunc(a.handleStreamDelete)

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

func TestInternalAPIStreamDeleteAll(t *testing.T) {

	req, err := http.NewRequest("DELETE", "", nil)
	if err != nil {
		t.Error(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"stream": "all",
	})

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.InternalAPIrFunc(a.handleStreamDeleteAll)

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

func TestInternalAPIStreamShow(t *testing.T) {

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
	handler := http.InternalAPIrFunc(a.handleStreamShow)

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

func TestInternalAPIStreamShowAll(t *testing.T) {

	req, err := http.NewRequest("PUT", "", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	a := testApp(false)
	handler := http.InternalAPIrFunc(a.handleStreamShowAll)

	a.Hub.Rules = make(map[string][]string)
	a.Hub.Rules["stream/large"] = []string{"audio", "video0"}
	a.Hub.Rules["stream/medium"] = []string{"audio", "video1"}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"stream/large":["audio","video0"],"stream/medium":["audio","video1"]}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}

}
*/
