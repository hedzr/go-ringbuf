package fast

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
	"unsafe"
)

// Opt interface the functional options
type Opt func(buf *ringBuf)

// New returns the RingBuffer object
func New(capacity uint32, opts ...Opt) RingBuffer {
	if !isInitialized() {
		return nil
	}

	size := roundToPower2(capacity)
	// logger := initLogger("all.log", "debug")
	logger := initLoggerConsole(zapcore.DebugLevel)
	rb := &ringBuf{
		data:       make([]rbItem, size),
		head:       0,
		tail:       0,
		cap:        size,
		capModMask: size - 1, // = 2^n - 1
		logger:     logger,
	}

	for _, opt := range opts {
		opt(rb)
	}

	if rb.debugMode {
		rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
	}

	for i := 0; i < (int)(size); i++ {
		rb.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
	}

	return rb
}

// WithDebugMode enables the internal debug mode for more logging output, and the metrics for debugging
func WithDebugMode(debug bool) Opt {
	return func(buf *ringBuf) {
		buf.debugMode = debug
	}
}

// Dbg exposes some internal fields for debugging
type Dbg interface {
	GetGetWaits() uint64
	GetPutWaits() uint64
	Debug(enabled bool) (lastState bool)
	ResetCounters()
}

func (rb *ringBuf) GetGetWaits() uint64 {
	return atomic.LoadUint64(&rb.getWaits)
}

func (rb *ringBuf) GetPutWaits() uint64 {
	return atomic.LoadUint64(&rb.putWaits)
}

func (rb *ringBuf) ResetCounters() {
	atomic.StoreUint64(&rb.getWaits, 0)
	atomic.StoreUint64(&rb.putWaits, 0)
}

func (rb *ringBuf) Close() (err error) {
	if rb.logger != nil {
		err = rb.logger.Sync()
		rb.logger = nil
	}
	return
}

func (rb *ringBuf) qty(head, tail uint32) (quantity uint32) {
	if tail >= head {
		quantity = tail - head
	} else {
		quantity = rb.capModMask + (tail - head)
	}
	return quantity
}

func (rb *ringBuf) Quantity() uint32 {
	return rb.Size()
}

func (rb *ringBuf) Size() uint32 {
	var quantity uint32
	// head = atomic.LoadUint32(&fast.head)
	// tail = atomic.LoadUint32(&fast.tail)
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)

	if tail >= head {
		quantity = tail - head
	} else {
		quantity = rb.capModMask + (tail - head)
	}

	return quantity
}

func (rb *ringBuf) Cap() uint32 {
	return rb.cap
}

func (rb *ringBuf) IsEmpty() (b bool) {
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)
	// var tail, head uint32
	// head = atomic.LoadUint32(&fast.head)
	// tail = atomic.LoadUint32(&fast.tail)
	b = head == tail
	return
}

func (rb *ringBuf) IsFull() (b bool) {
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)
	// var tail, head uint32
	// head = atomic.LoadUint32(&fast.head)
	// tail = atomic.LoadUint32(&fast.tail)
	b = ((tail + 1) & rb.capModMask) == head
	return
}

func (rb *ringBuf) Debug(enabled bool) (lastState bool) {
	lastState = rb.debugMode
	rb.debugMode = enabled
	return
}

func roundToPower2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}