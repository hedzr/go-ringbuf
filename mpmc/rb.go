package mpmc

import (
	"fmt"
	"net"
	"runtime"
	"sync/atomic"

	"github.com/hedzr/go-ringbuf/v2/mpmc/state"
)

// ringBuf implements a circular buffer. It is a fixed size,
// and new writes will be blocked when queue is full.
type ringBuf[T any] struct {
	cap        uint32
	capModMask uint32
	_          [CacheLinePadSize - 8]byte
	head       uint32
	_          [CacheLinePadSize - 4]byte //nolint:revive
	tail       uint32
	_          [CacheLinePadSize - 4]byte //nolint:revive
	putWaits   uint64
	_          [CacheLinePadSize - 8]byte //nolint:revive
	getWaits   uint64
	_          [CacheLinePadSize - 8]byte //nolint:revive
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

func (rb *ringBuf[T]) Put(item T) (err error) { return rb.Enqueue(item) } //nolint:revive

func (rb *ringBuf[T]) Enqueue(item T) (err error) { //nolint:revive
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
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 0, 2) { //nolint:gomnd
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
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 2, 1) { //nolint:gomnd
			err = ErrRaced // runtime.Gosched() // never happens
		}

		if state.VerboseEnabled {
			state.Verbose("[W] enqueued",
				"tail", tail, "new-tail", nt, "head", head, "value", toString(holder.value),
				"value(rb.data[0])", toString(rb.data[0].value),
				"value(rb.data[1])", toString(rb.data[1].value))
		}
		return
	}
}

func (rb *ringBuf[T]) Get() (item T, err error) { return rb.Dequeue() } //nolint:revive

func (rb *ringBuf[T]) Dequeue() (item T, err error) { //nolint:revive
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
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 1, 3) { //nolint:gomnd
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
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 3, 0) { //nolint:gomnd
			err = ErrRaced // runtime.Gosched() // never happens
		}

		if state.VerboseEnabled {
			state.Verbose("[ringbuf][GET] states are:",
				"cap", rb.Cap(), "qty", rb.qty(head, tail), "tail", tail, "head", head, "new-head", nh, "item", toString(item))
		}

		// if item == nil {
		// 	err = errors.New("[ringbuf][GET] cap: %v, qty: %v, head: %v, tail: %v, new head: %v", rb.cap, rb.qty(head, tail), head, tail, nh)
		// }

		// if !rb.debugMode {
		//	log.VWarnf("[ringbuf][GET] ", zap.Uint32("cap", rb.cap), zap.Uint32("qty", rb.qty(head, tail)),
		//    zap.Uint32("tail", tail), zap.Uint32("head", head), zap.Uint32("new head", nh))
		// }
		// log.Fatal("[ringbuf][GET] [ERR] unexpected nil element value FOUND!")

		// } else {
		//	// log.Printf("<< %v DEQUEUED, %v => %v, tail: %v", item, head, nh, tail)
		return
	}
}

func toString(i any) (sz string) {
	switch s := i.(type) {
	case string:
		sz = s
	case []byte:
		sz = string(s)
	// case *rbItem[T]:
	// 	sz = toString(s.value)
	case *struct {
		RemoteAddr *net.UDPAddr
		Data       []byte
	}:
		sz = string(s.Data)
	default:
		sz = fmt.Sprintf("%v", i)
	}
	return sz
}
