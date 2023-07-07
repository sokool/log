package log

import "fmt"

const (
	DEBUG   Level = 4
	INFO    Level = 3
	WARNING Level = 2
	ERROR   Level = 1
)

type Level int

func (l Level) Render(short, color bool) string {
	s := "%s"
	if color {
		switch l {
		case ERROR:
			s = "\x1b[31;1m%s\x1b[0m"
		case WARNING:
			s = "\x1b[33;1m%s\x1b[0m"
		case INFO:
			s = "\x1b[32;1m%s\x1b[0m"
		case DEBUG:
			s = "\x1b[36;1m%s\x1b[0m"
		default:
			s = "\x1b[39;1m%s\x1b[0m"
		}
	}
	if short {
		return fmt.Sprintf(s, l.GoString())
	}
	return fmt.Sprintf(s, l.String())
}

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (l Level) GoString() string {
	switch l {
	case DEBUG:
		return "DBG"
	case INFO:
		return "INF"
	case WARNING:
		return "WRN"
	case ERROR:
		return "ERR"
	default:
		return "UNK"
	}

}

var levels = map[string]Level{"dbg": DEBUG, "err": ERROR, "inf": INFO, "wrn": WARNING}
