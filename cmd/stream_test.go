package cmd

//This is named z... so it runs last - else it conflicts with streamtest

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/websocket"
	"github.com/phayes/freeport"
)

var configBinaryCurl = `--- 
commands: 
  - "curl  --request POST --data-binary @bin.dat ${binarydata}"
outurl: "ws://127.0.0.1:%d"
session: 7525cb39-554e-43e1-90ed-3a97e8d1c6bf
uuid: 49270598-9da2-4209-98da-e559f0c587b4
streams: 
  -   destination: "${outurl}/${uuid}/${session}/front/medium"
      feeds: 
        - binarydata
`

var upgrader = websocket.Upgrader{}

func TestStream(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cmd := exec.Command("rm", "./bin.dat")
	err := cmd.Run()
	cmd = exec.Command("rm", "./rx.dat")
	err = cmd.Run()

	port, err := freeport.GetFreePort()
	if err != nil {
		fmt.Printf("Error getting free port %v", err)
	}

	writeConfig(port)

	_, err = writeDataFile(1024000, "./bin.dat")
	if err != nil {
		fmt.Printf("Error writing data file %v", err)
	}

	go func() {
		//give wsReceiver a chance to start
		time.Sleep(100 * time.Millisecond)
		streamCmd.Execute()
	}()

	go wsReceiver(port, t)

	time.Sleep(10 * time.Millisecond)

	var outbuf bytes.Buffer

	cmd = exec.Command("diff", "./bin.dat", "./rx.dat")
	cmd.Stdout = &outbuf
	err = cmd.Run()
	stdout := outbuf.String()
	if stdout != "" {
		t.Errorf("Data sent and received is different: %v", stdout)
	}
	cmd = exec.Command("rm", "./bin.dat")
	err = cmd.Run()
	cmd = exec.Command("rm", "./rx.dat")
	err = cmd.Run()
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

func wsReceiver(port int, t *testing.T) {

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
}

/*

	// handle memprofile
	go func() {

		if memprofile != "" {

			time.Sleep(time.Duration(duration) * time.Second)

			f, err := os.Create(memprofile)

			if err != nil {
				log.WithField("error", err).Fatal("Could not create memory profile")
			}

			defer f.Close()

			if err := pprof.WriteHeapProfile(f); err != nil {
				log.WithField("error", err).Fatal("Could not write memory profile")
			}

			defer pprof.StopCPUProfile()
			close(closed)
		}
	}()

	// handle cpuprofile
	if cpuprofile != "" {

		f, err := os.Create(cpuprofile)

		if err != nil {
			log.WithField("error", err).Fatal("Could not create CPU profile")
		}

		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			log.WithField("error", err).Fatal("Could not start CPU profile")
		}

		defer pprof.StopCPUProfile()

	}
*/

/*

 TODO

 start sending a stream, then cause websocket to die
 profile streaming only in test
 benchmark timing/latency of packets

*/
