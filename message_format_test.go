package log_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sokool/log"
)

func TestPropertyCollisions(t *testing.T) {
	m := log.NewMessage("%+v %+v", 0,
		log.Data{"level": "business", "attr": "original", "id": 1},
		log.Data{"id": 2, "nested": log.Data{"name": "a"}})
	p := m.Properties()
	if p["level"] != "INFO" || p["id"] != float64(2) {
		t.Fatal(p)
	}
	conflicts := p["attr"].(log.Data)
	if conflicts["level"] != "business" || conflicts["attr"] != "original" {
		t.Fatal(conflicts)
	}
	if !strings.Contains(m.Text(false, true), "nested.name=a") {
		t.Fatal(m.Text(false, true))
	}
}

func TestStandardFormatting(t *testing.T) {
	object := struct{ Name string }{"Jane"}
	cases := []struct {
		format string
		args   []any
	}{
		{"%v %v %v", []any{"hello", 42, true}},
		{"%08x %.2f %q %%", []any{42, 1.234, "hello"}},
		{"%*.*s", []any{8, 3, "abcdef"}},
		{"%8.3s", []any{"abcdef"}},
		{"%*.3s", []any{8, "abcdef"}},
		{"%8.*s", []any{3, "abcdef"}},
		{"%.*s", []any{0, "abcdef"}},
		{"%*s", []any{-8, "abc"}},
		{"%[3]*.[2]*[1]f", []any{1.234, 2, 8}},
		{"%[2]s %[1]d %s", []any{42, "hello", "world"}},
		{"%T %p %#v", []any{object, &object, object}},
		{"%s", []any{"a", 42}},
		{"%% %s", []any{"a"}},
		{"%[1]T %[1]p", []any{&object}},
		{"%s %d", []any{"hello"}},
		{"%[4]s", []any{"hello"}},
		{"%[1]s %s", []any{"hello"}},
		{"hello", []any{42}},
		{"%%", []any{42}},
		{"%s", nil},
	}
	for _, c := range cases {
		t.Run(c.format, func(t *testing.T) {
			want := fmt.Sprintf(c.format, c.args...)
			m := log.NewMessage(c.format, 0, c.args...)
			if got := m.Text(false, true); got != want {
				t.Fatalf("want %q, got %q", want, got)
			}
			if got := m.Text(false, false); got != want {
				t.Fatalf("without properties: want %q, got %q", want, got)
			}
		})
	}
}

func TestPropertyOperandReuse(t *testing.T) {
	v := struct{ Name string }{"Jane"}
	m := log.NewMessage("%[1]v %[1]T %[1]v", 0, v)
	want := "Name=Jane " + fmt.Sprintf("%T", v) + " Name=Jane"
	if got := m.Text(false, true); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestMissingOperandAfterProperties(t *testing.T) {
	m := log.NewMessage("%v %s", 0, log.Data{"id": 1})
	if got := m.Text(false, true); got != "id=1 %!s(MISSING)" {
		t.Fatal(got)
	}
}

func TestLiteralJSON(t *testing.T) {
	text := `response {"level":"ok"}`
	m := log.NewMessage(text, 0)
	if got := m.Text(false, true); got != text {
		t.Fatal(got)
	}
	if _, exists := m.Properties()["attr"]; exists {
		t.Fatal("literal JSON became properties")
	}
}

func TestPropertySnapshot(t *testing.T) {
	nested := log.Data{"id": 1}
	data := log.Data{"nested": nested}
	m := log.NewMessage("%+v", 0, data)
	nested["id"] = 2
	data["extra"] = true
	if got := m.Text(false, true); got != "nested.id=1" {
		t.Fatal(got)
	}
}

type formatCounter int

func (c *formatCounter) String() string {
	(*c)++
	return "snapshot"
}

func TestFormattingSnapshot(t *testing.T) {
	var calls formatCounter
	m := log.NewMessage("value %s", 0, &calls)
	for range 2 {
		if got := m.Text(false, true); got != "value snapshot" {
			t.Fatal(got)
		}
	}
	if calls != 1 {
		t.Fatalf("formatted %d times, want 1", calls)
	}
}
