package cmd

import (
	"sync"
	"testing"
	"time"

	"github.com/timdrysdale/vw/config"
)

var cmd0 = "sleep 0.1"
var cmd1 = "sleep 0.2"
var cmd2 = "sleep 0.3"

var testCommands = config.Commands{Commands: []string{cmd0, cmd1, cmd2}}

var cmdSlow0 = "sleep 1"
var cmdSlow1 = "sleep 2"
var cmdSlow2 = "sleep 3"

var testCommandsSlow = config.Commands{Commands: []string{cmdSlow0, cmdSlow1, cmdSlow2}}

func TestRunCommands(t *testing.T) {
	// try sleeping in external processes
	closed := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()
	runCaptureCommands(closed, &wg, testCommands)
	elapsed := time.Since(start)
	if elapsed < 300*time.Millisecond {
		t.Errorf("running sleep commands was too short %v", elapsed)
	}
	if elapsed > 350*time.Millisecond {
		t.Errorf("running sleep commands took too long %v", elapsed)
	}

	//try killing processes via closed
	wg.Add(1)
	go func() {
		select {
		case <-time.After(500 * time.Millisecond):
			close(closed)
		}
	}()

	start = time.Now()
	runCaptureCommands(closed, &wg, testCommandsSlow)
	elapsed = time.Since(start)
	if elapsed < 500*time.Millisecond {
		t.Errorf("killing sleep commands was too short %v", elapsed)
	}
	if elapsed > 1000*time.Millisecond {
		t.Errorf("killing sleep commands took too long %v", elapsed)
	}

}
