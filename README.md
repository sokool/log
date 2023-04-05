# Logger

Helps your implementation to track your code activity

## Why?

- `tag names` to group logs into groups
- three levels to print out `information`, `debug` or `error` messages
- `colored` when printing out to standard output
- implements `io.Writer` in order to implement your own logger or just hide
  dependencies for external logging
- has `Printf(format string, args ...any)`

## How to use it?

just download it to your source code

```shell
go get github.com/sokool/log
```

and then run it with that sample snippet

```go
import "github.com/sokool/log"

lgr := log.
    New(os.Stdout). // output is written to standard output
    Tag("foo")      // all messages are marked with tag name

lgr.Printf("hello %s", "world")
lgr.Printf("first: with tag name")
lgr.Printf("second:dbg with tag name and as debug")
lgr.Printf("third:err with tage name and explicit error level")

lgr = lgr.
    Tag("bar").                               // all messages are marked with bar tag name
    Options(log.Type | log.Tag | log.Colors). // no date and file location in output
    Verbose(false)                            // do not show DBG logs

lgr.Printf("now it's simple colored text with log type and tag")
lgr.Printf("dbg this will not be printed out")
lgr.Printf("%s", fmt.Errorf("are always printed even in non-verbose option"))
lgr.Write([]byte("by io.Writer method"))

```
and here is an output with `Options` such as `Date`, `Time`, `Tag`, `Type` and file `Location`

[![kaD03.png](https://imgtr.ee/images/2023/04/05/kaD03.png)](https://imgtr.ee/i/kaD03)