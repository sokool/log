# Logger

Helps your implementation to track your code activity

## Why?

- four levels `debug`, `information`, `warning` and `error`
- `tags` that help messages be organized into logical groups
- `colored` human readable output text
- structural `json` output for machines, ie on production environment
- `traceable` will add info about file name, location, function and line number
  of called log
- implements `io.Writer` in order to implement your own logger or just hide
  dependencies for external logging
- uses `Printf(format string, args ...any)` function idiom, well know in go
  ecosystem

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
lgr.Printf("second:dbg: with tag name and as debug")
lgr.Printf("third:err: with tage name and explicit error level %v", log.Data{"message": "it's bad", "code": 859})
lgr.Printf("fourth:wrn: just warning level level")

lgr = lgr.
    Tag("bar").                                             // all messages are marked with bar tag name
    Options(log.Levels | log.Tags | log.Time | log.Colors). // colored text with tags and levels
    Verbosity(log.WARNING)                                  // do not show DBG and INF logs

lgr.Infof("ignored")
lgr.Debugf("ignored")
lgr.Errorf("with list of %v attributes", map[string]any{"code": 404, "message": "not found"})
lgr.Write([]byte("wrn: by io.Writer method"))

lgr = lgr.
    Tag("ror:nik").
    Options(log.JSON). //
    Verbosity(log.DEBUG)

lgr.Printf("with %v messages %v", log.Data{"wat": "hi"}, log.Data{"qot": "there"})
```

and here is an output with `Options` such as `Date`, `Time`, `Tags`, `Levels`, `Trace` and `JSON`

[![kaD03.png](https://yourimageshare.com/ib/z6gS4dadrN.webp)](https://yourimageshare.com/ib/z6gS4dadrN.webp)