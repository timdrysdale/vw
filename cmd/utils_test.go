package cmd

import (
	"log"
	"testing"
)

var unix = "foo\r\n"
var windows = "foo\n"
var cleaned = "foo"

func TestClean(t *testing.T) {
	if clean(unix) != clean(windows) {
		log.Fatalf("did not clean strings %v %v\n", []byte(unix), []byte(windows))
	}

	if clean(unix) != cleaned {
		log.Fatalf("clean(string) did not match a cleaned string %v %v\n", []byte(unix), []byte(cleaned))
	}
}
