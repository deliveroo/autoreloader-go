package watcher

import (
	"fmt"
	"log"
	"os/exec"
)

type Watcher interface {
	// Add watches the given path, returning an error if the path is
	// not valid.
	Add(string) error

	// Close ensures the underlying watcher is closed.
	Close() error

	// Watch handles the given command and coordinates its
	// management when changed.
	Watch()
	Start() error
}

// kill terminates the given command, printing a reason.
func kill(bin *exec.Cmd, reason string) error {
	fmt.Println(reason)
	return bin.Process.Kill()
}

// must calls log.Fatal if the error is non-nil, prepending an optional
// message.
func must(err error, msg string) {
	if err != nil {
		s := err.Error()
		if msg != "" {
			s = msg + ": " + s
		}
		log.Fatal(s)
	}
}
