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
	var (
		autorestart = flag.Bool("autorestart", false, "automatically restarts the binary upon non-zero exit code")
		help        = flag.Bool("?", false, "prints the usage")
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

	// Start watcher.
	watcher, err := fsnotify.NewWatcher()
	must(err, "failed to create watcher")
	defer watcher.Close()
	must(watcher.Add(cmdFullPath), "failed to watch")

	for {
		// Launch the process.
		cmd := exec.Command(cmd, argv...)
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

			// Sleep for 250ms, ignoring any events received in the
			// interim, because several events may be received when the
			// binary changes (e.g. CHMOD, WRITE).
			sleep(250*time.Millisecond, watcher.Events)
		case err := <-watcher.Errors:
			must(err, "error while watching files")
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if *autorestart {
					fmt.Println("executable quit; reloading...")
					sleep(250*time.Millisecond, watcher.Events)
					continue
				}
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					if status.Signal() == syscall.SIGBUS {
						// Retry on bus error, as these are occasionally
						// encountered when restarting a binary hosted on a
						// docker volume.
						fmt.Println("retrying on bus error...")
						sleep(250*time.Millisecond, watcher.Events)
						continue
					}
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
