package log_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sokool/log"
)

func TestLoggerJSONProperties(t *testing.T) {
	var output bytes.Buffer
	l := log.New(&output, log.JSON)
	l.Errorf("failed %v %+v", 42, log.Data{"id": 7, "level": "custom"})
	var fields map[string]any
	if err := json.Unmarshal(output.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	if fields["level"] != "ERROR" || fields["text"] != "failed 42" || fields["id"] != float64(7) {
		t.Fatal(fields)
	}
}
