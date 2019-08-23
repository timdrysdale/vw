package cmd

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/phayes/freeport"
)

var configBinaryCurl = `--- 
commands: 
  - "curl  --request POST --data-binary @bin.dat ${binarydata}"
urlout: "ws://127.0.0.1:%d"
session: 7525cb39-554e-43e1-90ed-3a97e8d1c6bf
uuid: 49270598-9da2-4209-98da-e559f0c587b4
streams: 
  -   destination: "${urlout}/${uuid}/${session}/front/medium"
      feeds: 
        - binarydata
`

func TestRoot(t *testing.T) {

	port, err := freeport.GetFreePort()
	if err != nil {
		fmt.Printf("Error getting free port %v", err)
	}

	writeConfig(port)

	data, err := writeDataFile(1024, "./bin.dat")
	if err != nil {
		fmt.Printf("Error writing data file %v", err)
	}
	
	go func(){
		//give wsReceiver a chance to start 
		time.Sleep(10 * time.Millisecond)
		Execute()
	}
	
	msg := wsReceiver(port, t)

	if len(msg) != len(data) {
		t.Errorf("messages have different lengths: sent %v, received %v", len(data), len(msg))
	}

	//if !Equal(msg, data) {
	//	t.Errorf("Messages have different contents:\nsent: %v...\ngot : %v...", data[:10], msg[:10])
	//}

}

func writeConfig(port int) error {

	config := []byte(fmt.Sprintf(configBinaryCurl, port))

	name := "./vw.yaml"

	err := ioutil.WriteFile(name, config, 0644)

	return err
}

func writeDataFile(size int, name string) ([]byte, error) {

	data := make([]byte, size)
	rand.Read(data)

	err := ioutil.WriteFile(name, data, 0644)

	return data, err

}

func wsReceiver(port int, t *testing.T) []byte {

	var msg []byte

	addr := fmt.Sprintf(":%d", port)

	http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			t.Errorf("Error starting wsReceiver for test: %v", err)
		}
		defer conn.Close()

		msg, _, _ = wsutil.ReadClientData(conn)
		if err != nil {
			t.Errorf("Error starting wsReceiver for test: %v", err)
		} else {
			err = ioutil.WriteFile("./rx.dat", msg, 0644)
			if err != nil {
				t.Errorf("Error writing data to file: %v", err)

			}
		}
	}))
	return msg
}
