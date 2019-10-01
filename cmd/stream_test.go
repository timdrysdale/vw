package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"testing"
)

//curl  --request POST --data-binary @bin.dat http://localhost:${VW_PORT}/ts/test

func TestStream(t *testing.T) {

	t.Skip("Skipping - Need a way to configure vw before this test can be run :-)")

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
