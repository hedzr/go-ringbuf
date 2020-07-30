package fast

import (
	"github.com/hedzr/log"
)

// New returns the RingBuffer object
func New(capacity uint32, opts ...Opt) (ringBuffer RingBuffer) {
	if isInitialized() {
		size := roundUpToPower2(capacity)

		rb := &ringBuf{
			data:       make([]rbItem, size),
			head:       0,
			tail:       0,
			cap:        size,
			capModMask: size - 1, // = 2^n - 1
			//logger:     logger,
		}

		ringBuffer = rb

		for _, opt := range opts {
			opt(rb)
		}

		//if rb.debugMode && rb.logger != nil {
		//	// rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
		//}

		for i := 0; i < (int)(size); i++ {
			rb.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
			if rb.initializer != nil {
				rb.data[i].value = rb.initializer.PreAlloc(i)
			}
		}
	}
	return
}

// Opt interface the functional options
type Opt func(buf *ringBuf)

// WithItemInitialilzer provides your custom initializer for each data item.
func WithItemInitializer(initializeable Initializeable) Opt {
	return func(buf *ringBuf) {
		buf.initializer = initializeable
	}
}

// WithDebugMode enables the internal debug mode for more logging output, and collect the metrics for debugging
func WithDebugMode(debug bool) Opt {
	return func(buf *ringBuf) {
		buf.debugMode = debug
	}
}

// WithLogger setup a logger
func WithLogger(logger log.Logger) Opt {
	return func(buf *ringBuf) {
		buf.logger = logger
	}
}
