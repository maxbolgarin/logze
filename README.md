# logze

[![Go Version][version-img]][doc] [![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![GoReport][report-img]][report]

`logze` is a logging package that wraps the efficient [zerolog](https://github.com/rs/zerolog), providing a user-friendly interface similar to [slog](https://pkg.go.dev/golang.org/x/exp/slog). It offers the performance and features of zerolog with an interface that developers familiar with slog can easily adapt to.


## Table of Contents

- [Why you should try](#why-you-should-try)
- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Logger](#creating-a-logger)
  - [Logging Messages](#logging-messages)
  - [Logging Messages with formatting](#logging-messages-with-formatting)
  - [Global logger usage](#global-logger-usage)
  - [Configuration Options](#configuration-options)
- [Pros and Cons](#pros-and-cons)
- [Using Diode](#using-diode)
- [Contributing](#contributing)
- [License](#license)


## Why you should try

If you want to log an error in `zerolog` you will write something like this:

```go
log.Error().Err(err).Str("address", "127.0.0.1").Int("retry", n).Msg("cannot start server")
```

This is a quite long piece of code and you should remember the names of these methods. Let's look at a `logze` example:

```go
logze.Err(err, "cannot start server", "address", "127.0.0.1", "retry", n)
```

Firstly, it is shorter. Secondly, you call only one method, which names `Err`, like an `err` that you want to log.

In other hand, you may say, that it is better to use `slog` with the similar interface instead of `logze`. But `logze` also provides a `zerolog` efficency — it is 3 times faster that `slog` and only 15% slower that `zerolog` due to using `Fields()` instead of separate method for each type.



## Installation

To install `logze`, use:

```bash
go get -u github.com/maxbolgarin/logze/v2
```


## Usage

### Creating a Logger

To start using `logze` in your project, you can initialize a logger as follows:

```go
package main

import (
	"github.com/maxbolgarin/logze/v2"
)

func main() {
	logger := logze.New(logze.C().WithConsoleJSON(), "application", "logze-example")
	
	logger.Info("Starting application", "version", "1.0.0")
}

// Output: {"level":"info","message":"Starting application","version":"1.0.0","application":"logze-example","time":"2023-08-24T15:30:00Z"}
```


### Logging Messages

Once you have a logger, you can log messages at various levels:

```go
logger.Trace("Trace application", "version", "1.0.0")
logger.Debug("Debugging application start", "version", "1.0.0") 
logger.Info("Application started successfully", "number", 123) 
logger.Warn("Low memory warning", "data", map[string]int{"key1": 1, "key2": 2}) 
logger.Error("Disk space low", "error", errors.New("disk space error")) 
logger.Err(errors.New("disk space error"), "disk space low") 

/* Example Output:
{"level":"trace","caller":"/Users/alex/code/logze/logze_test.go:210","time":"2024-12-12T17:24:24+03:00","message":"Trace application","version":"1.0.0"}
{"level":"debug","message":"Debugging application start","version":"1.0.0","time":"2023-08-24T15:30:00Z"}
{"level":"info","message":"Application started successfully","number":123,"time":"2023-08-24T15:30:00Z"}
{"level":"warning","message":"Low memory warning","data":{"key1":1,"key2":2},"time":"2023-08-24T15:30:00Z"}
{"level":"error","message":"Disk space low","error":"disk space error","time":"2023-08-24T15:30:00Z"}
{"level":"error","message":"disk space low","error":"disk space error","time":"2023-08-24T15:30:00Z"}
*/
```


### Logging Messages with formatting

You can also log messages with formatting and fields in the structured way: you add formatting args and then key-value pairs. For example:

```go
logger.Debugf("Debugging application start at %s", time.Now(), "version", "1.0.0")
logger.Infof("Application %s started successfully", "name", "number", 123)
logger.Warnf("Low memory warning %d times", times)

// {"level":"debug","message":"Debugging application start at 2023-08-24T15:30:00Z","version":"1.0.0","time":"2023-08-24T15:30:00Z"}
// {"level":"info","message":"Application name started successfully","number":123,"time":"2023-08-24T15:30:00Z"}
// {"level":"warning","message":"Low memory warning 3 times","time":"2023-08-24T15:30:00Z"}
```


### Global logger usage

You can use the global logger from `logze` package instead of creating a new one:

```go
// Init global logger with trace level and one field pair (optional to provide options and fields)
logze.Init(logze.NewConfig().WithConsole().WithLevel(logze.TraceLevel), "foo", "bar")

// Trace log will print Caller
logze.Trace("trace message")

// Example of logging an error
logze.Err(errm.New("some_error"), "message")


// Here is the result	
// 10:32:57 TRC test/main.go:16 > trace message foo=bar
// 10:32:57 ERR message error=some_error foo=bar
```


### Configuration Options

You can configure the logger with various options:

- **Log Level**: Set the log level (`trace`, `debug`, `info`, `warn`, `error`, `fatal`).
- **Many Output Writers**: Direct logs to console, files or network writers, you can provide as many `io.Writer` as you want.
- **Ignore Messages**: Ignore specific log messages using `WithToIgnore`, that will check using `strings.Contains` on log message.
- **Error Counter**: Add error counters using `WithErrorCounter` or `WithSimpleErrorCounter`; it may be useful for metrics to count errors.
- **Stack Trace**: Enable/disable stack trace of errors; you can use [errm](https://github.com/maxbolgarin/errm) to get stack trace out of the box.
- **Diode Buffering**: Enable/disable and configure diode buffering.

Example:

```go
config := logze.NewConfig(w1, w2).
    WithConsole().
	WithLevel(logze.DebugLevel).
	WithToIgnore("ignore me").
	WithSimpleErrorCounter().
    WithStackTrace().
	WithDiodeSize(10000)

logger := logze.New(config)
```


## Pros and Cons

### Pros

- **High Performance**: 3 times faster than `slog`, leveraging `zerolog` efficient engine.
- **Structured Logging**: Native support for key-value pairs with any type of value.
- **Easy Transition**: Compatible interface with `slog`.
- **Easy Configuration**: Simplified configuration and initialization process.
- **Diode Buffering**: Supports non-blocking IO operations for high performance.

### Cons

- **Slight Overhead**: Approximately 15% slower than raw `zerolog` due to the usage of fields instead of typed methods.
- **Complexity**: Advanced features such as diode might be complex for new users.


## Using Diode 

By default, `logze` uses diode buffering to prevent blocking on log IO operations. This ensures high throughput but requires careful shutdown handling to prevent log loss. It's crucial to note that if your application shuts down immediately after logging, some log messages might be lost. This can be mitigated waiting for flushing logs or disabling the diode feature. Diode also can drop messages if too many logs are generated in a short span. 

To disable diode buffering:

```go
config := logze.NewConfig().WithNoDiode()
logger := logze.New(config)
```


## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.


[version-img]: https://img.shields.io/badge/Go-%3E%3D%201.19-%23007d9c
[doc-img]: https://pkg.go.dev/badge/github.com/maxbolgarin/logze
[doc]: https://pkg.go.dev/github.com/maxbolgarin/logze
[ci-img]: https://github.com/maxbolgarin/logze/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/maxbolgarin/logze/actions
[report-img]: https://goreportcard.com/badge/github.com/maxbolgarin/logze
[report]: https://goreportcard.com/report/github.com/maxbolgarin/logze
