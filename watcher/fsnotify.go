package watcher

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// Deprecated: pluease use github.com/cosmtrek/air or another tool instead.
type Notifier struct {
	Autorestart bool
	Interval    time.Duration
	Cmd         string
	Args        []string
	bin         *exec.Cmd
	watcher     *fsnotify.Watcher
	done        chan struct{}
}

// NewNotifier returns a Notifier with the given parameters, using
// fsnotify as the underlying watcher.
//
// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func NewNotifier(autorestart bool, interval int, cmd string, args []string) (*Notifier, error) {
	if interval == 0 {
		interval = 250
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Notifier{
		Autorestart: autorestart,
		Interval:    time.Duration(interval) * time.Millisecond,
		Cmd:         cmd,
		Args:        args,
		watcher:     w,
	}, nil
}

// Add returns an error if the given path is invalid.
//
// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (n *Notifier) Add(path string) error {
	return errors.Wrapf(n.watcher.Add(path), "failed to add path %s", path)
}

// sleep blocks for the given duration.
func (n *Notifier) sleep(d time.Duration, events chan fsnotify.Event) {
	timer := time.After(d)
	for {
		select {
		case <-events:
		case <-timer:
			return
		}
	}
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (n *Notifier) Watch() {
	for {
		n.bin = exec.Command(n.Cmd, n.Args...)
		n.bin.Stdout = os.Stdout
		n.bin.Stderr = os.Stderr
		n.bin.Stdin = os.Stdin
		must(n.bin.Start(), "bin.Start()")

		// Watch for exit.
		exited := make(chan error)
		go func() {
			exited <- n.bin.Wait()
		}()

		select {
		case <-n.watcher.Events:
			_ = kill(n.bin, "executable changed; reloading...")
			n.sleep(n.Interval, n.watcher.Events)
		case err := <-n.watcher.Errors:
			must(err, "error while polling files")
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if n.Autorestart {
					_ = kill(n.bin, "executable quit; reloading...")
					n.sleep(n.Interval, n.watcher.Events)
					continue
				}
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					// Retry on bus error, as these are occasionally
					// encountered when restarting a binary hosted on a
					// docker volume.
					if status.Signal() == syscall.SIGBUS {
						_ = kill(n.bin, "retrying on bus error...")
						n.sleep(n.Interval, n.watcher.Events)
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

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (n *Notifier) Start() error {
	for range n.done {
		return nil
	}
	return nil
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (n *Notifier) Close() error {
	n.watcher.Close()
	return nil
}
