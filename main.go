package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	var (
		cmd  = os.Args[1]
		argv = os.Args[2:]
	)

	// Find the full path for the command.
	cmdFullPath, err := exec.LookPath(cmd)
	must(err, "error looking path")

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
			time.Sleep(250 * time.Millisecond)
		case err := <-watcher.Errors:
			must(err, "error while watching files")
		case err := <-exited:
			exitCode := 1
			if err, ok := err.(*exec.ExitError); ok {
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					exitCode = status.ExitStatus()
				}
			}
			os.Exit(exitCode)
		}
	}
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func usage() {
	fmt.Println("usage: autoreloader command [args...]")
	os.Exit(1)
}