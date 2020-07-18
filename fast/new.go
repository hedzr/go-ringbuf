package fast

// New returns the RingBuffer object
func New(capacity uint32, opts ...Opt) (ringBuffer RingBuffer) {
	if isInitialized() {
		size := roundUpToPower2(capacity)
		//// logger := initLogger("all.log", "debug")
		//logger := initLoggerConsole(zapcore.DebugLevel)
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

		if rb.debugMode && rb.logger != nil {
			//	rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
		}

		for i := 0; i < (int)(size); i++ {
			rb.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
		}
	}
	return
}

// Opt interface the functional options
type Opt func(buf *ringBuf)

// WithDebugMode enables the internal debug mode for more logging output, and collect the metrics for debugging
func WithDebugMode(debug bool) Opt {
	return func(buf *ringBuf) {
		buf.debugMode = debug
	}
}

// WithLogger setup a logger
func WithLogger(logger Logger) Opt {
	return func(buf *ringBuf) {
		buf.logger = logger
	}
}
