# Autoreloader

```
usage: autoreloader-go command [arguments]
```

Autoreloader launches the specified command, and waits for it to exit. If
the executable changes in that time, the process is killed and restarted.
This is useful in a development environment to allow a service to restart
everytime it's rebuilt.
