package cmd

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"
)

//curl  --request POST --data-binary @bin.dat http://localhost:${VW_PORT}/ts/test

func testStream(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cmd := exec.Command("rm", "./bin.dat")
	err := cmd.Run()
	cmd = exec.Command("rm", "./rx.dat")
	err = cmd.Run()

	_, err = writeDataFile(1024000, "./bin.dat")
	if err != nil {
		fmt.Printf("Error writing data file %v", err)
	}

	t.Error("TestStream not implemented") // stream bin.dat here....

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

func writeDataFile(size int, name string) ([]byte, error) {

	data := make([]byte, size)
	rand.Read(data)

	err := ioutil.WriteFile(name, data, 0644)

	return data, err

}
