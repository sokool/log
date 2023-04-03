package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Message struct {
	Tag       string
	Type      string
	Text      string
	CreatedAt time.Time
}

func NewMessage(format string, args ...any) Message {
	m := Message{
		Text:      fmt.Sprintf(format, args...),
		Type:      "INF",
		CreatedAt: time.Now(),
	}
	if p := strings.Index(m.Text, ":"); p != -1 {
		m.Text, m.Tag = m.Text[p+1:], m.Text[0:p]
	}
	if p := strings.Index(m.Text, "dbg "); p != -1 {
		m.Text, m.Type = m.Text[p+3:], "DBG"
	} else if p = strings.Index(m.Text, "err "); p != -1 {
		m.Text, m.Type = m.Text[p+3:], "ERR"
	} else if p := strings.Index(m.Text, "inf "); p != -1 {
		m.Text, m.Type = m.Text[p+3:], "INF"
	} else if len(args) != 0 {
		if _, ok := args[0].(error); ok {
			m.Type = "ERR"
		}
	}
	m.Text = strings.TrimSpace(m.Text)

	return m
}

func (m Message) Render(o Option) string {
	var s string
	if o&Date != 0 {
		s += fmt.Sprintf("%s ", m.CreatedAt.Format("2006/01/02"))
	}
	if o&Time != 0 {
		s += fmt.Sprintf("%s ", m.CreatedAt.Format("15:04:05.000000"))
	}
	if o&Type != 0 {
		s += fmt.Sprintf("[%s] ", m.typ(o&Colors != 0))
	}
	if o&Tag != 0 && m.Tag != "" {
		s += fmt.Sprintf("[%s] ", m.tag(o&Colors != 0))
	}
	return fmt.Sprintf("%s%s", s, m.Text)
}

func (m Message) String() string {
	return m.Render(All)
}

func (m Message) WriteTo(w io.Writer) (int64, error) {
	var s string
	if w == os.Stdout {
		s = m.Render(All)
	} else {
		s = m.Render(All)
	}

	n, err := w.Write([]byte(s))
	return int64(n), err
}

func (m Message) tag(colors bool) string {
	if colors {
		return fmt.Sprintf("\x1b[36;1m%s\x1b[0m", m.Tag)
	}
	return m.Tag
}

func (m Message) typ(colors bool) string {
	if !colors {
		return m.Type
	}
	color := "%s"
	if colors {
		switch m.Type {
		case "INF":
			color = "\x1b[32;1m%s\x1b[0m" // green
		case "ERR":
			color = "\x1b[31;1m%s\x1b[0m" // red
		case "DBG":
			color = "\x1b[33;1m%s\x1b[0m" // yellow
		}
	}
	return fmt.Sprintf(color, m.Type)
}

type Handler func(Message)

type Option int64

const (
	Date Option = 1 << iota
	Time
	Type
	Tag
	Colors

	All = Date | Time | Type | Tag | Colors
)
