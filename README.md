# logze

[![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![GoReport][report-img]][report]

Package `logze` implements a [zerolog](https://github.com/rs/zerolog) wrapper providing a convenient and short interface for structural logging

Install: `go get github.com/maxbolgarin/logze`

- [Why you should try](#why-you-should-try)
- [How to use](#how-to-use)
- [Contributing](#contributing)


## Why you should try

If you want to log an error in `zerolog`, you will write something like this:

```go
log.Error().Err(err).Str("address", "127.0.0.1").Msg("cannot start server")
```

This is a quite long piece of code. You also should remember all of these methods. Let's look at a `logze` example:

```go
logze.Error(err, "cannot start server", "address", "127.0.0.1")
```

Firstly, it is shorter. Secondly, you call only one method, which calls `Error`. That is quite intuitive — if you want to log error, you will use `Error` method.

In other hand, you may say, that it is better to use `slog` with the similar interface instead of `logze`. But `logze` provides a `zerolog` efficency with a convinient interface.


## How to use

1. Firstly, you should create a new logger or init a global one: use `New` function or `Init`
2. To create a new logger, you should provide it a `Config`: use `NewConfig` and `With*` methods
3. After creating a new logger, you can use it's methods. Their names match the names of the logging levels, e.g. `Debug` or `Error`
4. To log formatted message, use methods with `f` suffix. You can also add field pairs to such formatted messages, just add them after format args, like: `logze.Infof("msg: %s", variable, "key", "value")`

**Warning!** Pretty logging on the console is made possible using the provided `WithConsole()` method in `Config`, but it is inefficient. Use `WithConsoleJSON` or `NewConfig(os.Stderr)` to use all `zerolog` advantages.


### Local logger example
 
```go
	// Create a new logger instance with pretty logging to console (useful for developing)
	lg := logze.New(logze.NewConfig().WithConsole())

	// Log a debug message
	lg.Debug("some debug message")

	// Create a new logger with fields, that will be added to all log messages
	lg = lg.WithFields("foo", "bar")

	// Example of formatting and adding fields simultaniously
	a := "message"
	lg.Infof("some info message with format %s", a, "key", "value")


	// Here is the result
	// 2024-09-18 22:00:54 DBG some debug message
	// 2024-09-18 22:00:55 INF some info message with format message key=value foo=bar
```

## Global logger example

```go
	// Init global logger with trace level and one field pair
	logze.Init(logze.NewConfig().WithConsole().WithLevel(logze.TraceLevel), "foo", "bar")

	// Trace log will print Caller
	logze.Trace("trace message")

	// Example of logging an error
	logze.Error(errm.New("some_error"), "message")


	// Here is the result	
	// 10:32:57 TRC test/main.go:16 > trace message foo=bar
	// 10:32:57 ERR message error=some_error foo=bar
```


## Contributing

If you'd like to contribute to `logze`, make a fork and submit a pull request!

Released under the [MIT License]

[MIT License]: LICENSE.txt
[doc-img]: https://pkg.go.dev/badge/github.com/maxbolgarin/logze
[doc]: https://pkg.go.dev/github.com/maxbolgarin/logze
[ci-img]: https://github.com/maxbolgarin/logze/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/maxbolgarin/logze/actions
[report-img]: https://goreportcard.com/badge/github.com/maxbolgarin/logze
[report]: https://goreportcard.com/report/github.com/maxbolgarin/logze
