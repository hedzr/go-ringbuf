package mpmc

import "sync/atomic"

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
	for i := 0; i < (int)(rb.cap); i++ {
		rb.data[i].readWrite = 0 // bit 0: readable, bit 1: writable
	}
	// atomic.StoreUint64((*uint64)(unsafe.Pointer(&rb.head)), 0)
	atomic.StoreUint32(&rb.head, 0)
	atomic.StoreUint32(&rb.tail, 0)
}

func (rb *ringBuf[T]) Debug(enabled bool) (lastState bool) {
	// lastState = rb.debugMode
	// rb.debugMode = enabled
	return
}

// roundUpToPower2 takes a uint32 positive integer and
// rounds it up to the next power of 2.
func roundUpToPower2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}
