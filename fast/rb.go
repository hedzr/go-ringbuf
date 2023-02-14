package fast

import (
	"fmt"
	"net"
	"runtime"
	"sync/atomic"

	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"
)

type (

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
		logger     log.Logger
		// logger     *zap.Logger
		// _         cpu.CacheLinePad
		initializer Initializeable
	}

	rbItem struct {
		readWrite uint64      // 0: writable, 1: readable, 2: write ok, 3: read ok
		value     interface{} // ptr
		_         [CacheLinePadSize - 8 - 8]byte
		// _         cpu.CacheLinePad
	}

	// ringer struct {
	//	cap uint32
	//	// _         [CacheLinePadSize-unsafe.Sizeof(uint32)]byte
	// }
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

		if rb.initializer != nil {
			rb.initializer.CloneIn(item, holder.value)
		} else {
			holder.value = item
		}
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 2, 1) {
			err = ErrRaced // runtime.Gosched() // never happens
		}
		if rb.debugMode && rb.logger != nil {
			rb.logger.Debugf("[W] tail %v => %v, head: %v | ENQUEUED value = %v | [0]=%v, [1]=%v",
				tail, nt, head, toString(holder.value), toString(rb.data[0].value), toString(rb.data[1].value))
		}
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
			item = rb.initializer.CloneOut(holder.value)
		} else {
			item = holder.value
			holder.value = 0
		}
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 3, 0) {
			err = ErrRaced // runtime.Gosched() // never happens
		}

		if rb.debugMode && rb.logger != nil {
			rb.logger.Debugf("[ringbuf][GET] cap=%v, qty=%v, tail=%v, head=%v, new head=%v, item=%v", rb.Cap(), rb.qty(head, tail), tail, head, nh, toString(item))
		}

		if item == nil {
			err = errors.New("[ringbuf][GET] cap: %v, qty: %v, head: %v, tail: %v, new head: %v", rb.cap, rb.qty(head, tail), head, tail, nh)
		}

		// if !rb.debugMode {
		//	rb.logger.Warn("[ringbuf][GET] ", zap.Uint32("cap", rb.cap), zap.Uint32("qty", rb.qty(head, tail)), zap.Uint32("tail", tail), zap.Uint32("head", head), zap.Uint32("new head", nh))
		// }
		// rb.logger.Fatal("[ringbuf][GET] [ERR] unexpected nil element value FOUND!")

		// } else {
		//	// log.Printf("<< %v DEQUEUED, %v => %v, tail: %v", item, head, nh, tail)
		return
	}
}

func toString(i interface{}) (sz string) {
	if s, ok := i.(string); ok {
		sz = s
	} else if s, ok := i.([]byte); ok {
		sz = string(s)
	} else if s, ok := i.(*rbItem); ok {
		sz = toString(s.value)
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
