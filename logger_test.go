package log_test

import (
	"bytes"
	"testing"

	"github.com/sokool/log"
)

func TestNew(t *testing.T) {
	var b bytes.Buffer
	l := log.New(&b, log.Levels|log.Tags|log.Colors).Tag("log")
	l.Printf("new:err: system %s", "failure")
	o := log.Levels | log.Tags | log.Trace
	if s := b.String(); s != "[\u001B[31;1mERR\u001B[0m] [\u001B[34;1mlog:new\u001B[0m] system failure\n" {
		t.Fatal()
	}
	if s, _ := log.NewMessage("err: oh no", 0).Render(o); string(s) != "[ERR] oh no logger_test.go:18" {
		t.Fatalf(string(s))
	}
	b.Reset()
	l = l.Options(o)
	func(m string, args ...any) {
		l.Trace(1).Printf(m, args...)
	}("foo test")
	if s := b.String(); s != "[INF] [log] foo test logger_test.go:25\n" {
		t.Fatalf(s)
	}
	b.Reset()

	if l.Tag("test").Printf("err: oh no"); b.String() != "[ERR] [test] oh no logger_test.go:31\n" {
		t.Fatalf(b.String())
	}
}

func TestMessage_Fields(t *testing.T) {
	var b bytes.Buffer
	type data = log.Data
	j := data{
		"foo": "yo",
		"baz": data{
			"hoz": data{
				"izy": []string{"one", "two"},
				"diz": []data{
					{"koz": "nice", "bar": "elo"},
				},
			},
		},
	}
	log := log.Default.Options(log.All).Writer(&b)
	log.Printf("datacenter:location:err: system %v\nlocation %v\ninfo %v", data{"test": "yo"}, j, data{"one": "jeden"})
	//fmt.Println(&b)

	//fmt.Println(b.String())
}

type data map[string]any
