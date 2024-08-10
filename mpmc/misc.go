package mpmc

import (
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
)

// Dbg exposes some internal fields for debugging
type Dbg interface {
	GetGetWaits() uint64
	GetPutWaits() uint64
	Debug(enabled bool) (lastState bool)
	ResetCounters()
}

func (rb *ringBuf[T]) GetGetWaits() uint64 {
	return atomic.LoadUint64(&rb.getWaits)
}

func (rb *ringBuf[T]) GetPutWaits() uint64 {
	return atomic.LoadUint64(&rb.putWaits)
}

func (rb *ringBuf[T]) ResetCounters() {
	atomic.StoreUint64(&rb.getWaits, 0)
	atomic.StoreUint64(&rb.putWaits, 0)
}

func (rb *ringBuf[T]) Close() {
	// if rb.logger != nil {
	// 	// err = rb.logger.Flush()
	// 	rb.logger = nil
	// }
	// return
}

func (rb *ringBuf[T]) qty(head, tail uint32) (quantity uint32) {
	if tail >= head {
		quantity = tail - head
	} else {
		quantity = rb.cap + (tail - head)
	}
	return
}

func (rb *ringBuf[T]) Quantity() uint32 {
	return rb.Size()
}

func (rb *ringBuf[T]) Size() (quantity uint32) {
	var tail, head uint32
	head = atomic.LoadUint32(&rb.head)
	tail = atomic.LoadUint32(&rb.tail)
	return rb.qty(head, tail)
}

func (rb *ringBuf[T]) Cap() uint32 {
	return rb.cap
}

func (rb *ringBuf[T]) CapReal() uint32 {
	return rb.capModMask
}

func (rb *ringBuf[T]) IsEmpty() (b bool) {
	var tail, head uint32
	head = atomic.LoadUint32(&rb.head)
	tail = atomic.LoadUint32(&rb.tail)
	b = head == tail
	return
}

func (rb *ringBuf[T]) IsFull() (b bool) {
	var tail, head uint32
	head = atomic.LoadUint32(&rb.head)
	tail = atomic.LoadUint32(&rb.tail)
	b = ((tail + 1) & rb.capModMask) == head
	return
}

// Reset will clear the whole queue, but it might be unsafe in SMP runtime environment.
func (rb *ringBuf[T]) Reset() {
	// atomic.StoreUint64((*uint64)(unsafe.Pointer(&rb.head)), MaxUint64)
	atomic.StoreUint32(&rb.head, MaxUint32)
	atomic.StoreUint32(&rb.tail, MaxUint32)
	for i := 0; i < int(rb.cap); i++ {
		rb.data[i].readWrite = 0 // bit 0: readable, bit 1: writable
	}
	// atomic.StoreUint64((*uint64)(unsafe.Pointer(&rb.head)), 0)
	atomic.StoreUint32(&rb.head, 0)
	atomic.StoreUint32(&rb.tail, 0)
}

func (rb *ringBuf[T]) Debug(enabled bool) (lastState bool) {
	// lastState = rb.debugMode
	// rb.debugMode = enabled
	_ = enabled
	return
}

func (rb *ringBuf[T]) String() (ret string) {
	var sb strings.Builder
	_, _ = sb.WriteRune('[')

	head := atomic.LoadUint32(&rb.head)
	tail := atomic.LoadUint32(&rb.tail)
	if head < tail {
		for i := head; i < tail; i++ {
			it := &rb.data[i]
		retry:
			if st := atomic.LoadUint64(&it.readWrite); st > 1 {
				runtime.Gosched() // time to time
				goto retry
			}
			str := fmt.Sprintf("%v,", it.value)
			_, _ = sb.WriteString(str)
		}
	} else if head > tail {
		for i := head; i < rb.cap; i++ {
			it := &rb.data[i]
		retry1:
			if st := atomic.LoadUint64(&it.readWrite); st > 1 {
				runtime.Gosched() // time to time
				goto retry1
			}
			str := fmt.Sprintf("%v,", it.value)
			_, _ = sb.WriteString(str)
		}

		for i := uint32(0); i < tail; i++ {
			it := &rb.data[i]
		retry2:
			if st := atomic.LoadUint64(&it.readWrite); st > 1 {
				runtime.Gosched() // time to time
				goto retry2
			}
			str := fmt.Sprintf("%v,", it.value)
			_, _ = sb.WriteString(str)
		}
	}

	// _, _ = sb.WriteRune(']')
	// _, _ = sb.WriteRune('/')
	str := fmt.Sprintf("]/%v", rb.Size())
	_, _ = sb.WriteString(str)
	ret = sb.String()
	return
}

// roundUpToPower2 takes a uint32 positive integer and
// rounds it up to the next power of 2.
func roundUpToPower2(v uint32) uint32 {
	v--          //nolint:revive
	v |= v >> 1  //nolint:revive
	v |= v >> 2  //nolint:gomnd,revive
	v |= v >> 4  //nolint:gomnd,revive
	v |= v >> 8  //nolint:gomnd,revive
	v |= v >> 16 //nolint:gomnd,revive
	v++          //nolint:revive
	return v
}
