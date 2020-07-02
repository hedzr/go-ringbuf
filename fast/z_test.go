package fast

import (
	"log"
	"testing"
)

type ll struct{}

func (l *ll) Flush() error {
	return nil
}

func (l *ll) Info(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func (l *ll) Debug(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func (l *ll) Warn(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func (l *ll) Fatal(fmt string, args ...interface{}) {
	log.Fatalf(fmt, args...)
}

func TestResets(t *testing.T) {
	logger := &ll{}

	rb := New(NLtd,
		WithDebugMode(true),
		WithLogger(logger),
	)
	defer rb.Close()
	rb.ResetCounters()
	rb.Reset()
	_ = rb.Put(3)
	x := rb.(*ringBuf)
	t.Logf("qty = %v, isEmpty = %v, isFull = %v", x.qty(x.head, x.tail), rb.IsEmpty(), rb.IsFull())
	_, _ = rb.Get()

	x.head = MaxUint32
	x.tail = MaxUint32
	_ = rb.Put(3)
	_, _ = rb.Get()
}
