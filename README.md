# logze

> Package logze implements a [zerolog](https://github.com/rs/zerolog) wrapper providing convenient and short functions for logging

Install: `go get github.com/maxbolgarin/logze`

## How to use

  1. Firstly, you should create a new logger or init a global one: use `New` function or `Init`.

  2. To create a new logger, you should provide it a `Config`: use `NewConfig` and `With*` methods.

  3. After creating a new logger, you can use it's methods.
     Their names match the names of the logging levels, e.g. `Debug` or `Error`.

  4. To log formatted message, use methods with "f" suffix. You can also add field pairs to such
     formatted messages, just add them after format args, like:
    `logze.Infof("msg: %s", variable, "key", "value")`

## Local logger example
```go
	lg := logze.New(logze.NewConfig().WithConsole())
	lg.Debug("some debug message")

	lg = lg.WithFields("foo", "bar")
	a := "message"
	lg.Infof("some info message with format %s", a, "another_key", "another_value")

	// 10:28:52 DBG some debug message
	// 10:28:52 INF some info message with format message another_key=another_value foo=bar
```

## Global logger example
```go
	logze.Init(logze.NewConfig().WithConsole().WithLevel(logze.TraceLevel), "foo", "bar")
	logze.Trace("trace message")
	logze.Error(errm.New("some_error"), "message")

	// 10:32:57 TRC test/main.go:16 > trace message foo=bar
	// 10:32:57 ERR message error=some_error foo=bar
```
