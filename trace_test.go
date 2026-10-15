package log_test

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/sokool/log"
)

func TestTraceRecognition(t *testing.T) {
	log.SetTraceKey()
	t.Cleanup(func() { log.SetTraceKey() })
	cases := []struct {
		name string
		data log.Data
		want string
	}{
		{"canonical", log.Data{"trace_id": "abc"}, "abc"},
		{"camel case", log.Data{"traceId": "abc"}, "abc"},
		{"pascal case", log.Data{"TraceID": "abc"}, "abc"},
		{"dot", log.Data{"trace.id": "abc"}, "abc"},
		{"dash", log.Data{"trace-id": "abc"}, "abc"},
		{"vendor", log.Data{"dd.trace_id": "abc"}, "abc"},
		{"canonical priority", log.Data{"trace_id": "first", "TraceID": "second"}, "first"},
		{"alphabetical priority", log.Data{"traceId": "second", "TraceID": "first"}, "first"},
		{"skip empty", log.Data{"trace_id": "", "traceId": "abc"}, "abc"},
		{"skip number", log.Data{"trace_id": 123, "traceId": "abc"}, "abc"},
		{"unrelated", log.Data{"trace_grid": "abc", "trace_id_backup": "def"}, ""},
		{"request id", log.Data{"request_id": "abc"}, ""},
		{"nested", log.Data{"nested": log.Data{"trace_id": "abc"}}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := log.NewMessage("request %+v", 0, c.data)
			if m.TraceID != c.want {
				t.Fatalf("want %q, got %q", c.want, m.TraceID)
			}
			if c.want != "" && strings.Contains(m.Text(false, true), "="+c.want) {
				// Other aliases are intentionally retained; test consumption with a single field.
				if len(c.data) == 1 {
					t.Fatal("trace ID repeated in text")
				}
			}
		})
	}
}

func TestTraceKeyConfiguration(t *testing.T) {
	t.Cleanup(func() { log.SetTraceKey() })
	data := log.Data{"trace_id": "automatic", "operation_Id": "azure", "custom": "first"}
	names := []string{"custom", "operation_Id"}
	log.SetTraceKey(names...)
	names[0] = "trace_id"
	m := log.NewMessage("%+v", 0, data)
	if m.TraceID != "first" {
		t.Fatal(m.TraceID)
	}
	log.SetTraceKey("operation_Id")
	if got := log.NewMessage("%+v", 0, data).TraceID; got != "azure" {
		t.Fatal(got)
	}
	log.SetTraceKey("missing", "operation_Id")
	if got := log.NewMessage("%+v", 0, data).TraceID; got != "azure" {
		t.Fatal(got)
	}
	log.SetTraceKey("missing")
	missing := log.NewMessage("%+v", 0, data)
	if missing.TraceID != "" {
		t.Fatal("unexpected automatic fallback")
	}
	if missing.Properties()["attr"].(log.Data)["trace_id"] != "automatic" {
		t.Fatal("trace_id is not reserved")
	}
	log.SetTraceKey()
	if got := log.NewMessage("%+v", 0, data).TraceID; got != "automatic" {
		t.Fatal(got)
	}
	if m.TraceID != "first" {
		t.Fatal("existing message changed")
	}
}

func TestTraceRendering(t *testing.T) {
	log.SetTraceKey()
	cases := []struct {
		name   string
		id     string
		option log.Option
		want   string
	}{
		{"shortened", "1875fhsohfh08203", log.Trace | log.Tags | log.Levels,
			"[DBG] [location:point] test #1875fhso main.go:24"},
		{"short", "abc", log.Trace, "test #abc main.go:24"},
		{"unicode", "ąęśćżźółabc", log.Trace, "test #ąęśćżźół main.go:24"},
		{"absent", "", log.Trace, "test main.go:24"},
		{"disabled", "abc", log.Tags | log.Properties, "[location:point] test"},
		{"colored", "abc", log.Trace | log.Colors,
			"test \x1b[36;1m#abc\x1b[0m \x1b[35;1mmain.go:24\x1b[0m"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := log.NewMessage("location:point:dbg: test %+v", 0, log.Data{"trace_id": c.id})
			m.File, m.Line = "/app/main.go", 24
			b, err := m.Render(c.option)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != c.want {
				t.Fatalf("want %q, got %q", c.want, b)
			}
		})
	}
}

func TestTraceJSON(t *testing.T) {
	log.SetTraceKey()
	for _, id := range []string{"1875fhsohfh08203", ""} {
		m := log.NewMessage("%+v", 0, log.Data{"trace_id": id})
		b, err := m.Render(log.JSON)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(b, &fields); err != nil {
			t.Fatal(err)
		}
		if id != "" && fields["trace_id"] != id {
			t.Fatal(fields)
		}
		if id == "" {
			if _, ok := fields["trace_id"]; ok {
				t.Fatal("empty metadata should be omitted")
			}
		}
	}
}

func TestConcurrentTraceConfiguration(t *testing.T) {
	t.Cleanup(func() { log.SetTraceKey() })
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			for range 20 {
				log.SetTraceKey("trace_id", "custom")
				m := log.NewMessage("%+v", 0, log.Data{"trace_id": "abc"})
				if m.TraceID != "abc" {
					t.Error(m.TraceID)
				}
			}
		}()
	}
	group.Wait()
}
