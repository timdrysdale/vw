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

func TestSlashify(t *testing.T) {

	if "/foo" != slashify("foo") {
		t.Errorf("Slashify not prefixing slash ")
	}
	if "//foo" == slashify("/foo") {
		t.Errorf("Slashify prefixing additional slash")
	}
	if "/foo" != slashify("/foo/") {
		t.Errorf("Slashify not removing trailing slash")
	}
	if "/foo" != slashify("foo/") {
		t.Errorf("Slashify not both removing trailing slash AND prefixing slash")
	}

	b := "foo/bar/rab/oof/"
	if "/foo/bar/rab/oof" != slashify(b) {
		t.Errorf("Slashify not coping with internal slashes %s -> %s", b, slashify(b))
	}

}

// Utilities only needed during testing - put them here
// to reduce build size in production

// mod from https://yourbasic.org/golang/compare-slices/
// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func clientExists(topics *topicDirectory, client clientDetails) bool {

	topics.Lock()
	existingClients := topics.directory[client.topic]
	topics.Unlock()

	for _, existingClient := range existingClients {
		if client.name == existingClient.name {
			return true

		}
	}

	return false

}
