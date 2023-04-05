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
log := log.New(os.Stdout).Tag("test")
log.Printf("first: hi")
log.Printf("second:dbg it's some details")
log.Printf("third:err oh no, that's not working")

```
and here is an output with all `Options` such as `Date`, `Time`, `Tag`, `Type` and file `Location`
![example](https://imgtr.ee/images/2023/04/05/kaRCJ.png)