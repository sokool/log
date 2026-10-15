package log

import "fmt"

type part struct {
	text     string
	args     []any
	argument *argument
}

func (p part) render(colors, show bool) string {
	if p.argument == nil {
		return p.text
	}
	a := *p.argument
	a.colors, a.show = colors, show
	args := append([]any(nil), p.args...)
	args[len(args)-1] = &a
	return fmt.Sprintf(p.text, args...)
}
