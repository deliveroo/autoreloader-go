package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	log.SetFlags(0)

	help := flag.Bool("?", false, "prints the usage")
	flag.Usage = usage
	flag.Parse()
	if *help || len(os.Args) < 2 {
		usage()
	}

	var (
		cmd  = os.Args[1]
		argv = os.Args[2:]
	)

	// Find the full path for the command.
	cmdFullPath, err := exec.LookPath(cmd)
	must(err, "")

	// Start watcher.
	watcher, err := fsnotify.NewWatcher()
	must(err, "failed to create watcher")
	defer watcher.Close()
	must(watcher.Add(cmdFullPath), "failed to watch")

	for {
		// Launch the process.
		cmd := exec.Command(cmdFullPath, argv...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		defer cmd.Process.Kill()
		must(cmd.Start(), "cmd.Start()")

		// Watch for exit.
		exited := make(chan error)
		go func() {
			exited <- cmd.Wait()
		}()

		select {
		case <-watcher.Events:
			log.Println("executable changed; reloading...")
			cmd.Process.Kill()
			sleep(250*time.Millisecond, watcher.Events)
		case err := <-watcher.Errors:
			must(err, "error while watching files")
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				} else {
					exitCode = 1
				}
			}
			os.Exit(exitCode)
		}
	}
}

// sleep pauses the current goroutine for at least duration d, swallowing
// all fsnotify events received in the interim.
func sleep(d time.Duration, events chan fsnotify.Event) {
	timer := time.After(d)
	for {
		select {
		case <-events:
		case <-timer:
			return
		}
	}
}

func must(err error, msg string) {
	if err != nil {
		s := err.Error()
		if msg != "" {
			s = msg + ": " + s
		}
		log.Fatal(s)
	}
}

func usage() {
	fmt.Printf("usage: %s command [arguments]\n", os.Args[0])
	os.Exit(1)
}
