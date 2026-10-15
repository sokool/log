package log

import (
	"context"
	"sync"
)

var contexts keys

type keys struct {
	mu     sync.RWMutex
	reader func(context.Context) map[string]any
	names  []string
}

// SetContextData replaces the reader used for contexts formatted with %+v.
// The reader must support concurrent calls. A nil reader disables extraction.
// Changes affect new messages only.
func SetContextData(reader func(context.Context) map[string]any) {
	contexts.setReader(reader)
}

// SetTraceKey replaces the trace ID field names, checked in the given order.
// With no names, it restores automatic recognition. It affects new messages only.
func SetTraceKey(names ...string) {
	contexts.setNames(names...)
}

func (k *keys) setReader(reader func(context.Context) map[string]any) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.reader = reader
}

func (k *keys) setNames(names ...string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.names = append([]string(nil), names...)
}

func (k *keys) data(ctx context.Context) Data {
	k.mu.RLock()
	reader := k.reader
	k.mu.RUnlock()
	if reader == nil {
		return nil
	}
	return reader(ctx)
}

func (k *keys) traceNames() []string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return append([]string(nil), k.names...)
}
