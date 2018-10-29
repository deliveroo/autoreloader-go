package watcher

import (
	"fmt"
	"log"
	"os/exec"
)

type Watcher interface {
	Add(string) error
	Close() error
	Watch()
	Start() error
}

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
