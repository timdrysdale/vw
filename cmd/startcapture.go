package cmd

import (
	"fmt"
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

	// Start a process:
	tokens := strings.Split(command, " ")
	fmt.Printf("Starting capture with:\n%s\n", command)
	for i, tk := range tokens {
		fmt.Printf("%d(%s)\n", i, tk)
	}
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
			fmt.Printf("Exited runCommand %v\n", wg)
			//TODO shutdown the program so that we can use daemon monitoring to spot errors with capture
			return

		}

	}

}
