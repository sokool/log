package log_test

import (
	"bytes"
	"testing"

	"github.com/sokool/log/log"
)

func TestNew(t *testing.T) {
	var b bytes.Buffer
	log.New(&b, log.Type|log.Tag|log.Colors).Tag("log").Printf("new:err system %s", "failure")
	if s := b.String(); s != "[\u001B[31;1mERR\u001B[0m] [\u001B[36;1mlog:new\u001B[0m] system failure\n" {
		t.Fatal()
	}
}
