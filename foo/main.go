package main

import (
	"context"
	"os"

	"github.com/sokool/log"
)

func main() {
	// Use a dedicated key type to avoid collisions with other packages.
	type key string
	type data = map[string]any
	const traceID key = "trace_id"
	ctx := context.WithValue(context.Background(), traceID, "1875fhsohfh08203")

	// Set one global reader for contexts passed through %+v.
	// The recognized trace_id becomes Message.TraceID, not an inline property.
	log.SetContextData(func(ctx context.Context) data {
		return data{"trace_id": ctx.Value(traceID)}
	})

	// Default options include colors, properties, trace ID and source location.
	lgr := log.New(os.Stdout).Tag("foo")

	// %s formats ordinary text; %+v extracts data from the context.
	// The trace ID appears as #1875fhso before the source location.
	lgr.Printf("hello %s %+v", "world", ctx)

	// Prefixes add message tags and select the log level.
	lgr.Printf("second:dbg: with tag name and debug level %+v", ctx)
	lgr.Printf("third:err: explicit error level %v", data{"message": "it's bad", "code": 859})
	lgr.Printf("fourth:wrn: warning level")

	// Configuration methods return a new logger without changing the original.
	// This variant omits date and trace information, and filters out INFO and DEBUG.
	lgr = lgr.
		Tag("bar").
		Options(log.Levels | log.Tags | log.Time | log.Properties | log.Colors).
		Verbosity(log.WARNING)

	lgr.Infof("ignored")
	lgr.Debugf("ignored")

	// Object properties render inline as sorted key=value pairs.
	lgr.Errorf("with %+v attributes",
		log.Data{"code": 404, "message": "not found"})

	// Logger implements io.Writer; message prefixes still select the level.
	lgr.Write([]byte("wrn: by io.Writer method"))

	// JSON preserves nested properties and includes the full trace ID.
	// Verbosity remains independent of the output format.
	lgr = lgr.
		Tag("ror:nik").
		Options(log.JSON).
		Verbosity(log.DEBUG)

	lgr.Printf("with %+v messages %+v",
		log.Data{"wat": "hi"},
		log.Data{"qot": "there", "details": log.Data{"count": 2}})

	lgr.Printf("request finished %+v", ctx)
}
