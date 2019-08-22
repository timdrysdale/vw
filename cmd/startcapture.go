package cmd

import (
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func constructEndpoint(h *url.URL, inputName string) string {

	h.Path = inputName
	return h.String()
}

func populateInputNames(o *Output) {

	//for each stream, copy each item in Feeds as string into InputNames
	for i, s := range o.Streams {
		feedSlice, _ := s.Feeds.([]interface{})
		for _, feed := range feedSlice {
			o.Streams[i].InputNames = append(o.Streams[i].InputNames, feed.(string))
		}

	}

}

func mapEndpoints(o Output, h *url.URL) Endpoints {
	//go through feeds to collect inputs into map

	var e = make(Endpoints)

	for _, v := range o.Streams {
		for _, f := range v.InputNames {
			e[f] = constructEndpoint(h, f)
		}
	}

	return e
}

func expandCaptureCommands(c *Commands, e Endpoints) {

	// we rely on e being in scope for the mapper when it runs
	mapper := func(placeholderName string) string {
		if val, ok := e[placeholderName]; ok {
			return val
		} else {
			return placeholderName //don't change what we don't know about
		}

	}

	for i, raw := range c.Commands {
		c.Commands[i] = os.Expand(raw, mapper)
	}

}

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

	cmd := exec.Command(tokens[0], tokens[1:]...)
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

		default:
		}

	}

}
