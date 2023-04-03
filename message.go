package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	tag       string
	typ       string
	text      string
	createdAt time.Time
}

func NewMessage(format string, args ...any) Message {
	m := Message{
		text:      fmt.Sprintf(format, args...),
		typ:       "INF",
		createdAt: time.Now(),
	}
	if p := strings.Index(m.text, ":"); p != -1 {
		m.text, m.tag = m.text[p+1:], m.text[0:p]
	}
	if p := strings.Index(m.text, "dbg "); p != -1 {
		m.text, m.typ = m.text[p+3:], "DBG"
	} else if p = strings.Index(m.text, "err "); p != -1 {
		m.text, m.typ = m.text[p+3:], "ERR"
	} else if p := strings.Index(m.text, "inf "); p != -1 {
		m.text, m.typ = m.text[p+3:], "INF"
	} else if len(args) != 0 {
		if _, ok := args[0].(error); ok {
			m.typ = "ERR"
		}
	}
	m.text = strings.TrimSpace(m.text)

	return m
}

func (m Message) Render(o Option) string {
	var s string
	if o&Date != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("2006/01/02"))
	}
	if o&Time != 0 {
		s += fmt.Sprintf("%s ", m.createdAt.Format("15:04:05.000000"))
	}
	if o&Type != 0 {
		s += fmt.Sprintf("[%s] ", m.Type(o&Colors != 0))
	}
	if o&Tag != 0 && m.tag != "" {
		s += fmt.Sprintf("[%s] ", m.Tag(o&Colors != 0))
	}
	return fmt.Sprintf("%s%s", s, m.text)
}

func (m Message) String() string {
	return m.Render(All)
}

func (m Message) Text() string {
	return m.text
}

func (m Message) Tag(colors bool) string {
	if colors {
		return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", m.tag)
	}
	return m.tag
}

func (m Message) Type(colors bool) string {
	if !colors {
		return m.typ
	}
	color := "%s"
	if colors {
		switch m.typ {
		case "INF":
			color = "\x1b[32;1m%s\x1b[0m" // green
		case "ERR":
			color = "\x1b[31;1m%s\x1b[0m" // red
		case "DBG":
			color = "\x1b[33;1m%s\x1b[0m" // yellow
		}
	}
	return fmt.Sprintf(color, m.typ)
}

func (m Message) CreatedAt() time.Time {
	return m.createdAt
}

func (m Message) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Render(Date | Time | Tag | Type))
}
