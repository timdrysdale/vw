package cmd

import (
	"crypto/rand"
	"io/ioutil"
	"log"
	"testing"
)

var unix = "foo\r\n"
var windows = "foo\n"
var cleaned = "foo"

func TestUtilsClean(t *testing.T) {
	if clean(unix) != clean(windows) {
		log.Fatalf("did not clean strings %v %v\n", []byte(unix), []byte(windows))
	}

	if clean(unix) != cleaned {
		log.Fatalf("clean(string) did not match a cleaned string %v %v\n", []byte(unix), []byte(cleaned))
	}
}

func writeDataFile(size int, name string) ([]byte, error) {

	data := make([]byte, size)
	rand.Read(data)

	err := ioutil.WriteFile(name, data, 0644)

	return data, err

}
