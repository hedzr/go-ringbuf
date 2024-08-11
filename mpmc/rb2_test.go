package mpmc

import (
	"errors"
	"fmt"
	"testing"
)

func TestOverlappedRingBuf_Put_OneByOne(t *testing.T) { //nolint:revive
	var err error
	rb := NewOverlappedRingBuffer(NLtd, WithDebugMode[uint32](true))
	size := rb.Cap() - 1
	// t.Logf("Ring Buffer created, capacity = %v, real size: %v", size+1, size)
	defer rb.Close()

	for i := uint32(0); i < size; i++ {
		err = rb.Enqueue(i)
		if err != nil {
			t.Fatalf("faild on i=%v. err: %+v", i, err)
			// } else {
			// t.Logf("%5d. '%v' put, quantity = %v.", i, i, fast.Quantity())
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	for i := size; i < size+size; i++ {
		err = rb.Put(i)
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	var it any
	it, err = rb.Dequeue()
	_, _ = it, err
	t.Logf("  %3d. ringbuf elements -> %v", -1, rb)
	it, err = rb.Dequeue()
	_, _ = it, err
	t.Logf("  %3d. ringbuf elements -> %v", -1, rb)
	if fmt.Sprintf("%v", rb) != "[17,18,19,20,21,22,23,24,25,26,27,28,29,]/13" {
		t.Fatalf("faild: expecting elements are: [17,18,19,20,21,22,23,24,25,26,27,28,29,]/13")
	}

	for i := size * 2; i < size*3+2; i++ {
		err = rb.Put(i)
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}
	if fmt.Sprintf("%v", rb) != "[32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,]/15" {
		t.Fatalf("faild: expecting elements are: [32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,]/15")
	}

	for i := 0; ; i++ {
		it, err = rb.Dequeue()
		if err != nil {
			if errors.Is(err, ErrQueueEmpty) {
				break
			}
			t.Fatalf("faild on i=%v. err: %+v. item: %v", i, err, it)
			// } else {
			// t.Logf("< %3d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
		}
		if rb.Size() == 0 {
			// t.Log("empty ring buffer elements")
			if fmt.Sprintf("%v", rb) != "[]/0" {
				t.Fatalf("faild: expecting elements are: []/0")
			}
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	it, err = rb.Dequeue()
	if err == nil {
		t.Fatalf("faild: Dequeue on an empty ringbuf should return an ErrQueueEmpty object.")
	}
}

func TestOverlappedRingBuf_Random(t *testing.T) { //nolint:revive
	var err error
	var i uint32
	rb := NewOverlappedRingBuffer(NLtd, WithDebugMode[uint32](true))
	size := rb.Cap() - 1
	// t.Logf("Ring Buffer created, capacity = %v, real size: %v", size+1, size)
	defer rb.Close()

	for ; i < size; i++ {
		err = rb.Enqueue(i)
		if err != nil {
			t.Fatalf("faild on i=%v. err: %+v", i, err)
			// } else {
			// t.Logf("%5d. '%v' put, quantity = %v.", i, i, fast.Quantity())
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	for ; i < size+size; i++ {
		err = rb.Put(i)
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	t.Log("remove 2 elems.")

	var it any
	it, err = rb.Dequeue()
	_, _ = it, err
	t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	i++
	it, err = rb.Dequeue()
	_, _ = it, err
	t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	i++
	if fmt.Sprintf("%v", rb) != "[17,18,19,20,21,22,23,24,25,26,27,28,29,]/13" {
		t.Fatalf("faild: expecting elements are: [17,18,19,20,21,22,23,24,25,26,27,28,29,]/13")
	}

	t.Logf("removed 2 elems, and inserting some new...")

	for ; i < size*2+8; i++ {
		err = rb.Put(i)
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	it, err = rb.Dequeue()
	_, _ = it, err
	t.Logf("removed 1 elem: %v", it)
	t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	i++

	for ; i < size*3+3; i++ {
		err = rb.Put(i)
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}
	if fmt.Sprintf("%v", rb) != "[32,33,34,35,36,37,39,40,41,42,43,44,45,46,47,]/15" {
		t.Fatalf("faild: expecting elements are: [32,33,34,35,36,37,39,40,41,42,43,44,45,46,47,]/15")
	}

	// cleanup

	for i := 0; ; i++ {
		it, err = rb.Dequeue()
		if err != nil {
			if errors.Is(err, ErrQueueEmpty) {
				break
			}
			t.Fatalf("faild on i=%v. err: %+v. item: %v", i, err, it)
			// } else {
			// t.Logf("< %3d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
		}
		if rb.Size() == 0 {
			// t.Log("empty ring buffer elements")
			if fmt.Sprintf("%v", rb) != "[]/0" {
				t.Fatalf("faild: expecting elements are: []/0")
			}
		}
		t.Logf("  %3d. ringbuf elements -> %v", i, rb)
	}

	it, err = rb.Dequeue()
	if err == nil {
		t.Fatalf("faild: Dequeue on an empty ringbuf should return an ErrQueueEmpty object.")
	}
}
