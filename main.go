package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/deliveroo/autoreloader-go/watcher"
)

const interval = 250 * time.Millisecond

func main() {
	var (
		autorestart   = flag.Bool("autorestart", false, "automatically restarts the binary upon non-zero exit code")
		enablePolling = flag.Bool("poll", false, "use polling, not fsnotify, to monitor binary")
		help          = flag.Bool("?", false, "prints the usage")
	)
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()
	if *help || len(flag.Args()) == 0 {
		usage()
	}

	var (
		cmd  = flag.Arg(0)
		argv = flag.Args()[1:]
	)

	// Find the full path for the command.
	cmdFullPath, err := exec.LookPath(cmd)
	must(err, "")

	var w watcher.Watcher
	if *enablePolling {
		w = watcher.NewPoller(*autorestart, interval, cmd, argv)
	} else {
		n, err := watcher.NewNotifier(*autorestart, interval, cmd, argv)
		must(err, "")
		w = n
	}
	defer mustClose(w)

	mustNotNil(w, "watcher not initialized")
	must(w.Add(cmdFullPath), "failed to watch")
	go w.Watch()
	go must(w.Start(), "failed to start")
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

// usage prints the usage and quits.
func usage() {
	fmt.Printf("usage: %s command [arguments]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func mustClose(closer io.Closer) {
	if closer != nil {
		must(closer.Close(), "failed to close")
	}
}

func mustNotNil(v interface{}, msg string) {
	if v == nil {
		log.Fatal(msg)
	}
}
