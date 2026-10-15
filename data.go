package log

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type Data map[string]any

// Flat returns an iterator of flattened key-value pairs, sorted by key.
func (d Data) Flat(delim ...string) iter.Seq2[string, string] {
	delimChar := "."
	if len(delim) > 0 && delim[0] != "" {
		delimChar = delim[0]
	}

	return func(yield func(string, string) bool) {
		d.walk("", d, delimChar, yield)
	}
}

// properties returns a string of key=value pairs, optionally colored.
func (d Data) properties(color bool, delim ...string) string {
	var s strings.Builder

	for n, v := range d.Flat(delim...) {
		// Quote values if needed
		if strings.Contains(v, " ") {
			v = fmt.Sprintf(`"%s"`, v)
		}

		// Apply color if requested
		if color {
			n = fmt.Sprintf("\u001B[90;1m%s\u001B[0m", n)
			v = fmt.Sprintf("\u001B[37;3m%s\u001B[0m", v)
		}

		s.WriteString(fmt.Sprintf("%s=%v ", n, v))
	}

	return strings.TrimSpace(s.String())
}

// walk recursively flattens value v, sorting map keys at each level.
func (d Data) walk(p string, v any, delim string, yield func(string, string) bool) bool {
	switch x := v.(type) {

	case map[string]any:
		// 1. Collect keys
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		// 2. Sort keys for deterministic iteration
		slices.Sort(keys)

		// 3. Iterate over sorted keys
		for _, k := range keys {
			n := d.join(p, k, delim)
			if !d.walk(n, x[k], delim, yield) {
				return false
			}
		}

	case Data:
		// Same logic for Data (which is a distinct type from map[string]any in switches)
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		for _, k := range keys {
			n := d.join(p, k, delim)
			if !d.walk(n, x[k], delim, yield) {
				return false
			}
		}

	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				elem := rv.Index(i).Interface()
				n := d.join(p, strconv.Itoa(i), delim)
				if !d.walk(n, elem, delim, yield) {
					return false
				}
			}
			return true
		}

		return yield(p, fmt.Sprint(v))
	}

	return true
}

func (d Data) join(a, b, delim string) string {
	if a == "" {
		return b
	}
	return a + delim + b
}

//func (a Data) String() string {
//	b, _ := json.MarshalIndent(a, "", "\t")
//	return string(b)
//}
