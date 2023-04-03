package log_test

import (
	"fmt"
	"testing"

	"github.com/sokool/log"
)

func TestMessage_Render(t *testing.T) {
	type scenario struct {
		description string
		text        string
		args        []any
		msg         string
	}

	cases := []scenario{
		{"no type and text", "", nil, "[INF]"},
		{"no type", "hi", nil, "[INF] hi"},
		{"no type and with first err argument gives", "oh no %s", []any{fmt.Errorf("it's not working")}, "[ERR] oh no it's not working"},
		{"dbg type no text", "dbg", nil, "[DBG]"},
		{"dbg type and text", "dbg hi", nil, "[DBG] hi"},
		{"dbg type and text and arguments", "dbg it's a test of %s and %s", []any{"debug", "args"}, "[DBG] it's a test of debug and args"},
		{"inf type and text", "inf hi", nil, "[INF] hi"},
		{"err type and text", "err hi", nil, "[ERR] hi"},
		{"abc type and text", "abc hi", nil, "[INF] abc hi"},
		{"with tag", "payments: Tim balance updated", nil, "[INF] [payments] Tim balance updated"},
		{"with tag and no text", "payments:", nil, "[INF] [payments]"},
		{"with multiple tags and text", "payments:billing: Tim balance updated", nil, "[INF] [payments:billing] Tim balance updated"},
		{"tag with type", "payments:dbg hi again", nil, "[DBG] [payments] hi again"},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			m := log.NewMessage(c.text)
			if c.args != nil {
				m = log.NewMessage(c.text, c.args...)
			}
			if s := m.Render(log.Tag | log.Type); c.msg != s {
				t.Fatalf("expected `%s`, got `%s`", c.msg, s)
			}
		})
	}
}
