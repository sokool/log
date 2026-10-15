package log_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sokool/log"
)

func TestMessage_Render(t *testing.T) {
	type key string
	type data = map[string]any
	const traceID key = "trace_id"
	const method key = "method"
	const userID key = "user_id"
	ctx := context.WithValue(context.Background(), traceID, "1875fhsohfh08203")
	ctx = context.WithValue(ctx, method, "GET")
	ctx = context.WithValue(ctx, userID, 42)
	log.SetContextData(func(ctx context.Context) data {
		return log.Data{
			"trace_id": ctx.Value(traceID),
			"method":   ctx.Value(method),
			"user_id":  ctx.Value(userID),
		}
	})
	t.Cleanup(func() { log.SetContextData(nil) })

	type scenario struct {
		description string
		input       string
		args        []any
		output      string
		options     log.Option
	}
	cases := []scenario{
		{
			description: "context with trace id",
			input:       "http: request finished %+v",
			args:        []any{ctx},
			options:     log.Tags | log.Levels | log.Trace,
			output:      "[INF] [http] request finished #1875fhso main.go:24",
		},
		{
			description: "context with inline properties",
			input:       "handling %+v finished",
			args:        []any{ctx},
			output:      "[INF] handling method=GET user_id=42 finished",
		},
		{
			description: "no type and text",
			output:      "[INF]",
		},
		{
			description: "no type",
			input:       "hi",
			output:      "[INF] hi",
		},
		{
			description: "no type and with first err argument gives",
			input:       "oh no %s",
			args:        []any{fmt.Errorf("it's not working")},
			output:      "[ERR] oh no it's not working",
		},
		{
			description: "text with err message in it",
			input:       "some info example with err word inside message",
			output:      "[INF] some info example with err word inside message",
		},
		{
			description: "dbg type no text",
			input:       "dbg:",
			output:      "[DBG]",
		},
		{
			description: "debug type and text",
			input:       "dbg: hi",
			output:      "[DBG] hi",
		},
		{
			description: "warning type and text",
			input:       "wrn: hi",
			output:      "[WRN] hi",
		},
		{
			description: "dbg type and text and arguments",
			input:       "dbg: it's a test of %s and %s",
			args:        []any{"debug", "args"},
			output:      "[DBG] it's a test of debug and args",
		},
		{
			description: "inf type and text",
			input:       "inf: hi",
			output:      "[INF] hi",
		},
		{
			description: "err type and text",
			input:       "err: hi",
			output:      "[ERR] hi",
		},
		{
			description: "abc tag and text",
			input:       "abc: hi",
			output:      "[INF] [abc] hi",
		},
		{
			description: "abc tag and textx",
			input:       `test:err: some string and json {"one":1}`,
			output:      `[ERR] [test] some string and json {"one":1}`,
		},
		{
			description: "with arrow at beginning",
			input:       "⇨ http server started on [::]:9000",
			output:      "[INF] ⇨ http server started on [::]:9000",
		},
		{
			description: "tag and text with leading white spaces",
			input:       "payments:    Tim balance updated",
			output:      "[INF] [payments] Tim balance updated",
		},
		{
			description: "tag and no text",
			input:       "payments:",
			output:      "[INF] [payments]",
		},
		{
			description: "multiple comas",
			input:       "http:dbg: GET:%s",
			args:        []any{"https://test.pl"},
			output:      "[DBG] [http] GET:https://test.pl",
		},
		{
			description: "multiple comas",
			input:       "elo:szmero:err: failed:tricky string",
			output:      "[ERR] [elo:szmero] failed:tricky string",
		},
		{
			description: "multiple tags and text",
			input:       "payments:billing: Tim balance updated",
			output:      "[INF] [payments:billing] Tim balance updated",
		},
		{
			description: "tag and dbg type with text",
			input:       "payments:dbg: hi again",
			output:      "[DBG] [payments] hi again",
		},
		{
			description: "tag with spaces are ignored",
			input:       "tricky:name:err: is here",
			output:      "[ERR] [tricky:name] is here",
		},
		{
			description: "text with tag and err message in it",
			input:       "foo:inf:with no err type",
			output:      "[INF] [foo] with no err type",
		},
		{
			description: "text with attributes",
			input:       "bar: some %s and %d int with args %v and here is the %d value",
			args:        []any{"string", 834, map[string]any{"number": 999, "string": "hello world"}, 5},
			output:      `[INF] [bar] some string and 834 int with args number=999 string="hello world" and here is the 5 value`,
		},
		{
			description: "custom type attribute",
			input:       "%v",
			args:        []any{data{"a": "hello", "b": "world"}},
			output:      `[INF] a=hello b=world`,
		},
		{
			description: "v string",
			input:       "v",
			output:      `[INF] v`,
		},
		{
			description: "string as attribute",
			input:       "%v",
			args:        []any{"hi"},
			output:      `[INF] hi`,
		},
		{
			description: "json as a string to attributes",
			input:       `app:block: {"message":"nice!", "foo": "bar"}`,
			output:      `[INF] [app:block] {"message":"nice!", "foo": "bar"}`,
		},
		{
			description: "type text tag and as json",
			input:       `live:err:{"message":"hello world", "number": 42}`,
			output:      `[ERR] [live] {"message":"hello world", "number": 42}`,
		},
		{
			description: "without properties in output text",
			input:       `some:kind:wrn: raz %s dwa %v trzy %d hi %v`,
			args:        []any{"1", log.Data{"hello": "elo"}, 3, log.Data{"foo": "bar"}},
			options:     log.Tags | log.Levels,
			output:      `[WRN] [some:kind] raz 1 dwa trzy 3 hi`,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			m := log.NewMessage(c.input, 0)
			if c.args != nil {
				m = log.NewMessage(c.input, 0, c.args...)
			}
			if c.options == 0 {
				c.options = log.Tags | log.Levels | log.Properties
			}
			if c.options&log.Trace != 0 {
				m.File, m.Line = "main.go", 24
			}
			if s, _ := m.Render(c.options); c.output != string(s) {
				t.Fatalf("expected `%s`, got `%s`", c.output, s)
			}
		})
	}
}
