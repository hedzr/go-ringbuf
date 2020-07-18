package fast

import (
	"gopkg.in/hedzr/errors.v2"
	"io"
	"runtime"
	"sync/atomic"
)

type (
	// Queue interface provides a set of standard queue operations
	Queue interface {
		Enqueue(item interface{}) (err error)
		Dequeue() (item interface{}, err error)
		// Cap returns the outer capacity of the ring buffer.
		Cap() uint32
		// Size returns the quantity of items in the ring buffer queue
		Size() uint32
		IsEmpty() (b bool)
		IsFull() (b bool)
		Reset()
	}

	// RingBuffer interface provides a set of standard ring buffer operations
	RingBuffer interface {
		io.Closer // for logger

		Queue

		Put(item interface{}) (err error)
		Get() (item interface{}, err error)

		// Quantity returns the quantity of items in the ring buffer queue
		Quantity() uint32

		// // Cap returns the outer capacity of the ring buffer.
		// Cap() uint32
		// IsEmpty() (b bool)
		// IsFull() (b bool)

		Debug(enabled bool) (lastState bool)

		ResetCounters()
	}

	// ringBuf implements a circular buffer. It is a fixed size,
	// and new writes will be blocked when queue is full.
	ringBuf struct {
		// isEmpty bool
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
		data       []rbItem
		debugMode  bool
		logger     Logger
		//logger     *zap.Logger
		// _         cpu.CacheLinePad
	}

	rbItem struct {
		readWrite uint64      // 0: writable, 1: readable, 2: write ok, 3: read ok
		value     interface{} // ptr
		_         [CacheLinePadSize - 8 - 8]byte
		// _         cpu.CacheLinePad
	}

	// Logger interface fo ringBuf
	Logger interface {
		Flush() error
		Info(fmt string, args ...interface{})
		Debug(fmt string, args ...interface{})
		Warn(fmt string, args ...interface{})
		Fatal(fmt string, args ...interface{})
	}

	ringer struct {
		cap uint32
		// _         [CacheLinePadSize-unsafe.Sizeof(uint32)]byte
	}
)

func (rb *ringBuf) Put(item interface{}) (err error) {
	err = rb.Enqueue(item)
	return
}

func (rb *ringBuf) Enqueue(item interface{}) (err error) {
	var tail, head, nt uint32
	var holder *rbItem
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

		holder.value = item
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 2, 1) {
			runtime.Gosched() // never happens
		}
		// if fast.debugMode {
		////log.Printf("[W] tail %v => %v, head: %v | ENQUEUED value = %v", tail, nt, head, item)
		// }
		return
	}
}

func (rb *ringBuf) Get() (item interface{}, err error) {
	item, err = rb.Dequeue()
	return
}

func (rb *ringBuf) Dequeue() (item interface{}, err error) {
	var tail, head, nh uint32
	var holder *rbItem
	for {
		//var quad uint64
		//quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
		//head = (uint32)(quad & MaxUint32_64)
		//tail = (uint32)(quad >> 32)
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

		item = holder.value
		holder.value = 0
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 3, 0) {
			runtime.Gosched() // never happens
		}

		// if fast.debugMode {
		// 	fast.logger.Debug("[ringbuf][GET] ", zap.Uint32("cap", fast.cap), zap.Uint32("qty", fast.qty(head, tail)), zap.Uint32("tail", tail), zap.Uint32("head", head), zap.Uint32("new head", nh))
		// }

		if item == nil {
			err = errors.New("[ringbuf][GET] cap: %v, qty: %v, head: %v, tail: %v, new head: %v", rb.cap, rb.qty(head, tail), head, tail, nh)

			//if !rb.debugMode {
			//	rb.logger.Warn("[ringbuf][GET] ", zap.Uint32("cap", rb.cap), zap.Uint32("qty", rb.qty(head, tail)), zap.Uint32("tail", tail), zap.Uint32("head", head), zap.Uint32("new head", nh))
			//}
			//rb.logger.Fatal("[ringbuf][GET] [ERR] unexpected nil element value FOUND!")

			//} else {
			//	// log.Printf("<< %v DEQUEUED, %v => %v, tail: %v", item, head, nh, tail)
		}
		return
	}
}
