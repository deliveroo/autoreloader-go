# autoreloader-go

⚠️ ⚠️ ⚠️ **THIS TOOL IS DEPRECATED! PLEASE USE [github.com/cosmtrek/air](https://github.com/cosmtrek/air) OR SIMILAR TOOL INSTEAD.** ⚠️ ⚠️ ⚠️

A small Go app for watching and reloading executables.

```
usage: autoreloader-go command [arguments]
  -?	prints the usage
  -autorestart
    	automatically restarts the binary upon non-zero exit code
  -poll
    	use polling, not fsnotify, to monitor binary
```

Autoreloader launches the specified command, and waits for it to exit. If the
executable changes in that time, the process is killed and restarted.  This is
useful in a development environment to allow a service to restart every time
it's rebuilt.
