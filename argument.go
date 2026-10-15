package log

import (
	"context"
	"encoding/json"
	"fmt"
)

type argument struct {
	value    any
	data     Data
	colors   bool
	show     bool
	explicit bool
}

func (a *argument) prepare() {
	if ctx, ok := a.value.(context.Context); ok && a.explicit {
		a.data = Data{}
		for k, v := range contexts.data(ctx) {
			a.data[k] = v
		}
		return
	}
	if a.explicit {
		a.data = a.objectProperties()
		return
	}
	a.data = a.deprecatedProperties()
}

// Deprecated: retained for legacy %v property extraction.
func (a *argument) deprecatedProperties() Data {
	switch a.value.(type) {
	case fmt.Formatter, error:
		return nil
	}
	return a.objectProperties()
}

func (a *argument) objectProperties() Data {
	b, err := json.Marshal(a.value)
	if err != nil || len(b) == 0 || b[0] != '{' {
		return nil
	}
	var data Data
	if json.Unmarshal(b, &data) != nil {
		return nil
	}
	return data
}

func (a *argument) Format(s fmt.State, verb rune) {
	if a.data == nil {
		fmt.Fprintf(s, "%%!%c(%T=%v)", verb, a.value, a.value)
		return
	}
	if a.show {
		fmt.Fprintf(s, fmt.FormatString(s, 's'), a.data.properties(a.colors))
	}
}
