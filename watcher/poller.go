package watcher

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/radovskyb/watcher"
)

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
type Poller struct {
	Autorestart bool
	Interval    time.Duration
	Cmd         string
	Args        []string
	bin         *exec.Cmd
	watcher     *watcher.Watcher
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func NewPoller(autorestart bool, interval int, cmd string, args []string) *Poller {
	if interval == 0 {
		interval = 250
	}
	return &Poller{
		Autorestart: autorestart,
		Interval:    time.Duration(interval) * time.Millisecond,
		Cmd:         cmd,
		Args:        args,
		watcher:     watcher.New(),
	}
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (p *Poller) Add(path string) error {
	return errors.Wrapf(p.watcher.Add(path), "failed to add path %s", path)
}

func (p *Poller) sleep(d time.Duration, events chan watcher.Event) {
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
func (p *Poller) Watch() {
	for {
		p.bin = exec.Command(p.Cmd, p.Args...)
		p.bin.Stdout = os.Stdout
		p.bin.Stderr = os.Stderr
		p.bin.Stdin = os.Stdin
		must(p.bin.Start(), "bin.Start()")

		// Watch for exit.
		exited := make(chan error)
		go func() {
			exited <- p.bin.Wait()
		}()

		select {
		case <-p.watcher.Event:
			_ = kill(p.bin, "executable changed; reloading...")
			p.sleep(p.Interval, p.watcher.Event)
		case err := <-p.watcher.Error:
			if err != watcher.ErrWatchedFileDeleted {
				must(err, "error while polling files")
			}
		case err := <-exited:
			var exitCode int
			if err, ok := err.(*exec.ExitError); ok {
				if p.Autorestart {
					_ = kill(p.bin, "executable quit; reloading...")
					p.sleep(p.Interval, p.watcher.Event)
					continue
				}
				if status, ok := err.Sys().(syscall.WaitStatus); ok {
					// Retry on bus error, as these are occasionally
					// encountered when restarting a binary hosted on a
					// docker volume.
					if status.Signal() == syscall.SIGBUS {
						_ = kill(p.bin, "retrying on bus error...")
						p.sleep(p.Interval, p.watcher.Event)
						continue
					}
					exitCode = status.ExitStatus()
				} else {
					exitCode = 1
				}
			}
			os.Exit(exitCode)
		case <-p.watcher.Closed:
			_ = kill(p.bin, "executable quit; reloading...")
			return
		}
	}
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (p *Poller) Start() error {
	return errors.Wrap(p.watcher.Start(p.Interval), "poller.Start")
}

// Deprecated: please use github.com/cosmtrek/air or another tool instead.
func (p *Poller) Close() error {
	p.watcher.Close()
	return nil
}
