package cmd

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func runCaptureCommands(closed <-chan struct{}, wgExternal *sync.WaitGroup, c Commands) {
	wgExternal.Done()

	var wg sync.WaitGroup

	for _, command := range c.Commands {
		wg.Add(1)
		go runCommand(closed, &wg, command)
	}

	wg.Wait()
}

func runCommand(closed <-chan struct{}, wg *sync.WaitGroup, command string) {
	defer wg.Done()

	tokens := strings.Split(command, " ")
	cmd := exec.Command(tokens[0], tokens[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return
	}

	finished := make(chan struct{})

	go func() {
		_ = cmd.Wait()
		close(finished)
	}()

	for {
		select {

		case <-closed:
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill process: ", err)
			}
			return

		case <-finished:
			return
		}

	}

}
