package mpmc

import (
	"errors"
	"testing"
)

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

		for i := 0; i < int(size); i++ {
			ringBuffer.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
			if ringBuffer.initializer != nil {
				ringBuffer.data[i].value = ringBuffer.initializer.PreAlloc(i)
			}
		}
	}
	return
}

func TestResets(t *testing.T) {
	rb := newRingBuf(NLtd,
		WithDebugMode[int](true),
	)
	defer rb.Close()
	rb.ResetCounters()
	rb.Reset()
	_ = rb.Put(3)

	x := rb // rb.(*ringBuf[int])
	t.Logf("qty = %v, isEmpty = %v, isFull = %v", x.qty(x.head, x.tail), rb.IsEmpty(), rb.IsFull())

	_, _ = rb.Get()

	x.head = MaxUint32
	x.tail = MaxUint32
	_ = rb.Put(3)
	_, _ = rb.Get()
}

func checkerr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

func checkqty[T any](t *testing.T, desc string, rb1 *ringBuf[T], expect uint32) {
	if qty, inner := rb1.Quantity(), rb1.qty(rb1.head, rb1.tail); qty == inner && qty == expect { //nolint:gocritic
		t.Logf("[%s] qty = %v / %v | expected: %v", desc, qty, inner, expect)
	} else {
		t.Fatalf("FATAL ERROR: [%s] qty = %v / %v | expected: %v | WRONG!!", desc, qty, inner, expect)
	}
}

func checkresult[T any](t *testing.T, desc string, rb1 *ringBuf[T], got any, expect int) {
	if g, ok := got.(int); ok && g == expect {
		t.Logf("[%s] got = %v / %v | expected: %v", desc, got, g, expect)
	} else {
		t.Fatalf("[%s] got = %v / %v | expected: %v | WRONG!!", desc, got, g, expect)
	}
	_ = rb1
}

func TestRoundedQty(t *testing.T) {
	rb := newRingBuf[int](4)
	rb1 := rb // .(*ringBuf[int])

	var err error
	var it any

	if it, err = rb.Dequeue(); !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("expect empty event. it: %v", it)
	}

	err = rb.Enqueue(1)
	checkerr(t, err)
	checkqty(t, "Enqueue(1)", rb, 1)

	err = rb.Enqueue(2)
	checkerr(t, err)
	checkqty(t, "Enqueue(2)", rb, 2)

	err = rb.Enqueue(3)
	checkerr(t, err)
	checkqty(t, "Enqueue(3)", rb, 3)

	if err = rb.Enqueue(3); !errors.Is(err, ErrQueueFull) { // full, 3
		t.Fatal("expect full event")
	}

	it, err = rb.Dequeue() // 3 -> 2
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb, 2)
	checkresult(t, "Dequeue()", rb, it, 1)

	err = rb.Enqueue(4) // 3 -> 4
	// t.Log(rb1.qty(rb1.head, rb1.tail))  // wanted: 3
	checkerr(t, err)
	checkqty(t, "Enqueue(4)", rb, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb, 2)
	checkresult(t, "Dequeue()", rb, it, 2)

	err = rb.Enqueue(5)
	checkerr(t, err)
	checkqty(t, "Enqueue(5)", rb, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 2)
	checkresult(t, "Dequeue()", rb1, it, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb, 1)
	checkresult(t, "Dequeue()", rb, it, 4)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb, 0)
	checkresult(t, "Dequeue()", rb, it, 5)

	if _, err = rb.Dequeue(); !errors.Is(err, ErrQueueEmpty) {
		t.Fatal("expect empty event")
	}
}
