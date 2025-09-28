package mpmc

// New returns the RingBuffer object.
//
// It returns [ErrQueueFull] when you're trying to put a new
// element into a full ring buffer.
func New[T any](capacity uint32, opts ...Opt[T]) (ringBuffer RingBuffer[T]) {
	return newRingBuffer(func(capacity uint32, opts ...Opt[T]) (ringBuffer RingBuffer[T]) {
		size := roundUpToPower2(capacity)
		rb := &ringBuf[T]{
			data:       make([]rbItem[T], size),
			cap:        size,
			capModMask: size - 1, // = 2^n - 1
		}
		for _, opt := range opts {
			opt(rb)
		}
		ringBuffer = rb
		return
	}, capacity, opts...)
}

// NewOverlappedRingBuffer make a new instance of the overlapped ring
// buffer, which overwrites its head element if it's full.
//
// When an unexpected state occurred, the returned value could be nil.
//
// In this case, checking it for unavailable value is recommended.
// This could happen if a physical core fault detected in high-freq,
// high-pressure, and high-temperature place.
//
// For the normal runtime environment, unexpected state should be
// impossible, so ignore it is safe.
func NewOverlappedRingBuffer[T any](capacity uint32, opts ...Opt[T]) (ringBuffer RichOverlappedRingBuffer[T]) {
	return newOverlappedRingBuffer(func(capacity uint32, opts ...Opt[T]) (ringBuffer RichOverlappedRingBuffer[T]) {
		size := roundUpToPower2(capacity)
		rb := &orbuf[T]{
			ringBuf[T]{
				data:       make([]rbItem[T], size),
				cap:        size,
				capModMask: size - 1, // = 2^n - 1
			},
		}
		for _, opt := range opts {
			opt(&rb.ringBuf)
		}
		ringBuffer = rb
		return
	}, capacity, opts...)
}

// Creator _
type Creator[T any] func(capacity uint32, opts ...Opt[T]) (ringBuffer RingBuffer[T])
type OverlappedCreator[T any] func(capacity uint32, opts ...Opt[T]) (ringBuffer RichOverlappedRingBuffer[T])

func newOverlappedRingBuffer[T any](creator OverlappedCreator[T], capacity uint32, opts ...Opt[T]) (ringBuffer RichOverlappedRingBuffer[T]) {
	if isInitialized() {
		ringBuffer = creator(capacity, opts...)
	}
	return
}

func newRingBuffer[T any](creator Creator[T], capacity uint32, opts ...Opt[T]) (ringBuffer RingBuffer[T]) {
	if isInitialized() {
		ringBuffer = creator(capacity, opts...)
	}
	return
}

// Opt interface the functional options
type Opt[T any] func(rb *ringBuf[T])

// WithItemInitializer provides your custom initializer for each data item.
func WithItemInitializer[T any](initializeable Initializeable[T]) Opt[T] {
	return func(buf *ringBuf[T]) {
		buf.initializer = initializeable
	}
}

// WithDebugMode enables the internal debug mode for more logging output, and collect the metrics for debugging
func WithDebugMode[T any](_ bool) Opt[T] {
	return func(_ *ringBuf[T]) {
		// buf.debugMode = debug
	}
}

// // WithLogger setup a logger
// func WithLogger[T any](logger log.Logger) Opt[T] {
// 	return func(buf *ringBuf[T]) {
// 		// buf.logger = logger
// 	}
// }
