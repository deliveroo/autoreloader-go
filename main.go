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
	"github.com/radovskyb/watcher"
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

	// Prefer polling, if enabled.
	if *enablePolling {
		poller := watcher.New()
		must(poller.Add(cmdFullPath), "failed to watch")
		go poll(cmd, argv, poller, *autorestart)
		go must(poller.Start(interval), "poller.Start()")
		return
	}

	// Start watcher.
	watcher, err := fsnotify.NewWatcher()
	must(err, "failed to create watcher")
	defer watcher.Close()
	must(watcher.Add(cmdFullPath), "failed to watch")
	watch(cmd, argv, watcher, *autorestart)
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

// run executes the given command with the specified arguments.
func run(cmd string, argv []string) *exec.Cmd {
	bin := exec.Command(cmd, argv...)
	bin.Stdout = os.Stdout
	bin.Stderr = os.Stderr
	bin.Stdin = os.Stdin
	must(bin.Start(), "bin.Start()")
	return bin
}

// poll runs the polling-based watcher.
func poll(cmd string, argv []string, poller *watcher.Watcher, autorestart bool) {
	var bin *exec.Cmd
	defer mustKill(bin)

	for {
		// Run the binary.
		bin = run(cmd, argv)

		// Watch for exit.
		exited := make(chan error)
		go func() {
			exited <- bin.Wait()
		}()

		select {
		case <-poller.Event:
			fmt.Println("executable changed; reloading...")
			_ = bin.Process.Kill()
		case err := <-poller.Error:
			if err != watcher.ErrWatchedFileDeleted {
				must(err, "error while polling files")
			}
			time.Sleep(interval)
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if autorestart {
					fmt.Println("executable quit; reloading...")
					continue
				}
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					if status.Signal() == syscall.SIGBUS {
						// Retry on bus error, as these are occasionally
						// encountered when restarting a binary hosted on a
						// docker volume.
						fmt.Println("retrying on bus error...")
						continue
					}
					exitCode = status.ExitStatus()
				} else {
					exitCode = 1
				}
			}
			os.Exit(exitCode)
		case <-poller.Closed:
			return
		}
	}
}

// watch runs the fsnotify-based watcher.
func watch(cmd string, argv []string, watcher *fsnotify.Watcher, autorestart bool) {
	var bin *exec.Cmd
	defer mustKill(bin)

	for {
		bin = run(cmd, argv)

		// Watch for exit.
		exited := make(chan error)
		go func() {
			exited <- bin.Wait()
		}()

		select {
		case <-watcher.Events:
			log.Println("executable changed; reloading...")
			_ = bin.Process.Kill()

			// Sleep for 250ms, ignoring any events received in the
			// interim, because several events may be received when the
			// binary changes (e.g. CHMOD, WRITE).
			sleep(250*time.Millisecond, watcher.Events)
		case err := <-watcher.Errors:
			must(err, "error while watching files")
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if autorestart {
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

func mustKill(bin *exec.Cmd) {
	if bin != nil {
		must(bin.Process.Kill(), "error killing process")
	}
}
