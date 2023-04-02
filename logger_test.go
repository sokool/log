package log_test

import (
	"testing"

	"github.com/sokool/gokit/log"
)

func TestName(t *testing.T) {
	pl := log.Default.WithTag("payments")
	pl.Print("hi")

	dl := log.Default
	dl.Print("payments:inf hi")

	log.Default.Print("dbg hi")
	log.Default.Print("err hi")

}
