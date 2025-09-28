package mpmc

import (
	"runtime"
	"sync/atomic"

	"github.com/hedzr/go-ringbuf/v2/mpmc/state"
)

type orbuf[T any] struct {
	ringBuf[T]
}

func (rb *orbuf[T]) Put(item T) (err error) { return rb.Enqueue(item) } //nolint:revive

func (rb *orbuf[T]) EnqueueM(item T) (overwrites uint32, err error) { //nolint:revive
	var tail, head, nt, nh uint32
	var holder *rbItem[T]
	for {
		head = atomic.LoadUint32(&rb.head)
		tail = atomic.LoadUint32(&rb.tail)
		nt = (tail + 1) & rb.capModMask

		isEmpty := head == tail
		if isEmpty && head == MaxUint32 {
			err = ErrQueueNotReady
			return
		}

		isFull := nt == head
		if isFull {
			nh = (head + 1) & rb.capModMask
			atomic.CompareAndSwapUint32(&rb.head, head, nh)
			overwrites++
		}

		holder = &rb.data[tail]
		atomic.CompareAndSwapUint32(&rb.tail, tail, nt)
	retry:
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 0, 2) { //nolint:gomnd
			if !atomic.CompareAndSwapUint64(&holder.readWrite, 1, 2) { //nolint:gomnd
				if atomic.LoadUint64(&holder.readWrite) == 0 {
					goto retry // sometimes, short circuit
				}
				runtime.Gosched() // time to time
				continue
			}
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

func (rb *orbuf[T]) EnqueueMRich(item T) (size, overwrites uint32, err error) { //nolint:revive
	var tail, head, nt, nh uint32
	var holder *rbItem[T]
	for {
		head = atomic.LoadUint32(&rb.head)
		tail = atomic.LoadUint32(&rb.tail)
		nt = (tail + 1) & rb.capModMask

		isEmpty := head == tail
		if isEmpty && head == MaxUint32 {
			err = ErrQueueNotReady
			return
		}

		isFull := nt == head
		if isFull {
			nh = (head + 1) & rb.capModMask
			atomic.CompareAndSwapUint32(&rb.head, head, nh)
			overwrites++
			size = rb.qty(head, tail) + 1
		}

		holder = &rb.data[tail]
		atomic.CompareAndSwapUint32(&rb.tail, tail, nt)
	retry:
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 0, 2) { //nolint:gomnd
			if !atomic.CompareAndSwapUint64(&holder.readWrite, 1, 2) { //nolint:gomnd
				if atomic.LoadUint64(&holder.readWrite) == 0 {
					goto retry // sometimes, short circuit
				}
				runtime.Gosched() // time to time
				continue
			}
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

		size = rb.qty(head, tail) + 1
		return
	}
}

func (rb *orbuf[T]) Enqueue(item T) (err error) { //nolint:revive
	var tail, head, nt, nh uint32
	var holder *rbItem[T]
	for {
		head = atomic.LoadUint32(&rb.head)
		tail = atomic.LoadUint32(&rb.tail)
		nt = (tail + 1) & rb.capModMask

		isEmpty := head == tail
		if isEmpty && head == MaxUint32 {
			err = ErrQueueNotReady
			return
		}

		isFull := nt == head
		if isFull {
			nh = (head + 1) & rb.capModMask
			atomic.CompareAndSwapUint32(&rb.head, head, nh)
		}

		holder = &rb.data[tail]
		atomic.CompareAndSwapUint32(&rb.tail, tail, nt)
	retry:
		if !atomic.CompareAndSwapUint64(&holder.readWrite, 0, 2) { //nolint:gomnd
			if !atomic.CompareAndSwapUint64(&holder.readWrite, 1, 2) { //nolint:gomnd
				if atomic.LoadUint64(&holder.readWrite) == 0 {
					goto retry // sometimes, short circuit
				}
				runtime.Gosched() // time to time
				continue
			}
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

func (rb *orbuf[T]) Get() (item T, err error) { return rb.Dequeue() } //nolint:revive

func (rb *orbuf[T]) Dequeue() (item T, err error) { //nolint:revive
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

		return
	}
}
