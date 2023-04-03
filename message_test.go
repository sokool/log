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
		{"without type gives info", "hi", nil, "[INF] hi"},
		{"no type and with first err argument gives err type", "oh no %s", []any{fmt.Errorf("it's not working")}, "[ERR] oh no it's not working"},
		{"dbg type", "dbg hi", nil, "[DBG] hi"},
		{"inf type", "inf hi", nil, "[INF] hi"},
		{"err type", "err hi", nil, "[ERR] hi"},
		{"abc type", "abc hi", nil, "[INF] abc hi"},
		{"just tag", "payments: Tim balance updated", nil, "[INF] [payments] Tim balance updated"},
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
