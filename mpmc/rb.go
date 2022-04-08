package mpmc

import (
	"fmt"
	"github.com/hedzr/log"
	"net"
	"runtime"
	"sync/atomic"
)

// New returns the RingBuffer object
func New[T any](capacity uint32, opts ...Opt[T]) (ringBuffer RingBuffer[T]) {
	if isInitialized() {
		size := roundUpToPower2(capacity)

		rb := &ringBuf[T]{
			data:       make([]rbItem[T], size),
			cap:        size,
			capModMask: size - 1, // = 2^n - 1
		}

		ringBuffer = rb

		for _, opt := range opts {
			opt(rb)
		}

		// if rb.debugMode && rb.logger != nil {
		//	// rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
		// }

		for i := 0; i < (int)(size); i++ {
			rb.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
			if rb.initializer != nil {
				rb.data[i].value = rb.initializer.PreAlloc(i)
			}
		}
	}
	return
}

func newRingBuf[T any](capacity uint32, opts ...Opt[T]) (ringBuffer *ringBuf[T]) {
	if isInitialized() {
		size := roundUpToPower2(capacity)

		ringBuffer = &ringBuf[T]{
			data:       make([]rbItem[T], size),
			cap:        size,
			capModMask: size - 1, // = 2^n - 1
		}

		for _, opt := range opts {
			opt(ringBuffer)
		}

		// if rb.debugMode && rb.logger != nil {
		//	// rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
		// }

		for i := 0; i < (int)(size); i++ {
			ringBuffer.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
			if ringBuffer.initializer != nil {
				ringBuffer.data[i].value = ringBuffer.initializer.PreAlloc(i)
			}
		}
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
func WithDebugMode[T any](debug bool) Opt[T] {
	return func(buf *ringBuf[T]) {
		// buf.debugMode = debug
	}
}

// WithLogger setup a logger
func WithLogger[T any](logger log.Logger) Opt[T] {
	return func(buf *ringBuf[T]) {
		// buf.logger = logger
	}
}

// ringBuf implements a circular buffer. It is a fixed size,
// and new writes will be blocked when queue is full.
type ringBuf[T any] struct {
	cap        uint32
	capModMask uint32
	_          [CacheLinePadSize - 8]byte
	head       uint32
	_          [CacheLinePadSize - 4]byte
	tail       uint32
	_          [CacheLinePadSize - 4]byte
	putWaits   uint64
	_          [CacheLinePadSize - 8]byte
	getWaits   uint64
	_          [CacheLinePadSize - 8]byte
	data       []rbItem[T]
	// debugMode  bool
	// logger     log.Logger
	// logger     *zap.Logger
	// _         cpu.CacheLinePad
	initializer Initializeable[T]
}

type rbItem[T any] struct {
	readWrite uint64 // 0: writable, 1: readable, 2: write ok, 3: read ok
	value     T      // ptr
	_         [CacheLinePadSize - 8 - 8]byte
	// _         cpu.CacheLinePad
}

func (rb *ringBuf[T]) Put(item T) (err error) { return rb.Enqueue(item) }

func (rb *ringBuf[T]) Enqueue(item T) (err error) {
	var tail, head, nt uint32
	var holder *rbItem[T]
	for {
		head = atomic.LoadUint32(&rb.head)
		tail = atomic.LoadUint32(&rb.tail)
		nt = (tail + 1) & rb.capModMask

		isFull := nt == head
		if isFull {
			err = ErrQueueFull
			return
		}
		isEmpty := head == tail
		if isEmpty && head == MaxUint32 {
			err = ErrQueueNotReady
			return
		}

		holder = &rb.data[tail]

		atomic.CompareAndSwapUint32(&rb.tail, tail, nt)
	retry:
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 0, 2) {
			if atomic.LoadUint64(&holder.readWrite) == 0 {
				goto retry // sometimes, short circuit
			}
			runtime.Gosched() // time to time
			continue
		}

		if rb.initializer != nil {
			rb.initializer.CloneIn(item, &holder.value)
		} else {
			holder.value = item
		}
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 2, 1) {
			err = ErrRaced // runtime.Gosched() // never happens
		}

		log.VDebugf("[W] tail %v => %v, head: %v | ENQUEUED value = %v | [0]=%v, [1]=%v",
			tail, nt, head, toString(holder.value), toString(rb.data[0].value), toString(rb.data[1].value))
		return
	}
}

func (rb *ringBuf[T]) Get() (item T, err error) { return rb.Dequeue() }

func (rb *ringBuf[T]) Dequeue() (item T, err error) {
	var tail, head, nh uint32
	var holder *rbItem[T]
	for {
		// var quad uint64
		// quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
		// head = (uint32)(quad & MaxUint32_64)
		// tail = (uint32)(quad >> 32)
		head = atomic.LoadUint32(&rb.head)
		tail = atomic.LoadUint32(&rb.tail)

		isEmpty := head == tail
		if isEmpty {
			if head == MaxUint32 {
				err = ErrQueueNotReady
				return
			}
			err = ErrQueueEmpty
			return
		}

		holder = &rb.data[head]

		nh = (head + 1) & rb.capModMask
		atomic.CompareAndSwapUint32(&rb.head, head, nh)
	retry:
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 1, 3) {
			if atomic.LoadUint64(&holder.readWrite) == 1 {
				goto retry // sometimes, short circuit
			}
			runtime.Gosched() // time to time
			continue
		}

		if rb.initializer != nil {
			item = rb.initializer.CloneOut(&holder.value)
		} else {
			item = holder.value
			// holder.value = zero
		}
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 3, 0) {
			err = ErrRaced // runtime.Gosched() // never happens
		}

		log.VDebugf("[ringbuf][GET] cap=%v, qty=%v, tail=%v, head=%v, new head=%v, item=%v", rb.Cap(), rb.qty(head, tail), tail, head, nh, toString(item))

		// if item == nil {
		// 	err = errors.New("[ringbuf][GET] cap: %v, qty: %v, head: %v, tail: %v, new head: %v", rb.cap, rb.qty(head, tail), head, tail, nh)
		// }

		// if !rb.debugMode {
		//	log.VWarnf("[ringbuf][GET] ", zap.Uint32("cap", rb.cap), zap.Uint32("qty", rb.qty(head, tail)), zap.Uint32("tail", tail), zap.Uint32("head", head), zap.Uint32("new head", nh))
		// }
		// log.Fatal("[ringbuf][GET] [ERR] unexpected nil element value FOUND!")

		// } else {
		//	// log.Printf("<< %v DEQUEUED, %v => %v, tail: %v", item, head, nh, tail)
		return
	}
}

func toString(i any) (sz string) {
	if s, ok := i.(string); ok {
		sz = s
	} else if s, ok := i.([]byte); ok {
		sz = string(s)
		// } else if s, ok := i.(*rbItem[T]); ok {
		// 	sz = toString(s.value)
	} else if s, ok := i.(*struct {
		RemoteAddr *net.UDPAddr
		Data       []byte
	}); ok {
		sz = string(s.Data)
	} else {
		sz = fmt.Sprintf("%v", i)
	}
	return
}
