package log_test

import (
	"fmt"
	"testing"

	"github.com/sokool/log"
)

func TestMessage_Render(t *testing.T) {
	type scenario struct {
		description string
		input       string
		args        []any
		output      string
	}
	cases := []scenario{
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
			description: "dbg type and text",
			input:       "dbg: hi",
			output:      "[DBG] hi",
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
			output:      "[ERR] [test] some string and json one=1",
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
			input:       "bar: some %s and %d int with args %s",
			args:        []any{"string", 834, map[string]any{"number": 999, "string": "hello world"}},
			output:      `[INF] [bar] some string and 834 int with args number=999 string="hello world"`,
		},
		{
			description: "just json",
			input:       `app:block: {"message":"nice!", "foo": "bar"}`,
			output:      `[INF] [app:block] message=nice! foo=bar`,
		},
		{
			description: "type text tag and as json",
			input:       `live:err:{"message":"hello world", "number": 42}`,
			output:      `[ERR] [live] message="hello world" number=42`,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			m := log.NewMessage(c.input)
			if c.args != nil {
				m = log.NewMessage(c.input, c.args...)
			}
			if s := m.Render(log.Tag | log.Type); c.output != s {
				t.Fatalf("expected `%s`, got `%s`", c.output, s)
			}
		})
	}
}
