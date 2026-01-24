package mpmc

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestForIssue9(t *testing.T) { //nolint:revive
	buf := NewOverlappedRingBuffer[int](1000)

	var wg sync.WaitGroup
	writesPerGoroutine := 100
	numGoroutines := 2

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < writesPerGoroutine; i++ {
				buf.EnqueueM(base + i)
			}
		}(g * writesPerGoroutine)
	}

	wg.Wait()

	// Count records
	count := 0
	for !buf.IsEmpty() {
		_, err := buf.Dequeue()
		if err == nil {
			count++
		}
	}

	expected := numGoroutines * writesPerGoroutine
	fmt.Printf("Expected: %d, Got: %d\n", expected, count)
	if count != expected {
		t.Fatal("BUG: Data loss detected!")
	}
}

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

func TestOverlappedRingBuf_EnqueueM_OneByOne(t *testing.T) { //nolint:revive
	var err error
	var c, cap uint32
	rb := NewOverlappedRingBuffer(NLtd, WithDebugMode[uint32](true))
	size := rb.Cap() - 1
	// t.Logf("Ring Buffer created, capacity = %v, real size: %v", size+1, size)
	defer rb.Close()

	for i := uint32(0); i < size; i++ {
		cap, c, err = rb.EnqueueMRich(i)
		if err != nil || c > 0 {
			t.Fatalf("faild on i=%v, overwrites=%v. err: %+v", i, c, err)
			// } else {
			// t.Logf("%5d. '%v' put, quantity = %v.", i, i, fast.Quantity())
		}
		if cap != i+1 {
			t.Fatalf("faild on i=%v, cap=%v.", i, cap)
		}
		t.Logf("  %3d. ringbuf elements -> %v, cap=%v, overwrites=%v", i, rb, cap, c)
	}

	for i := size; i < size+size; i++ {
		c, err = rb.EnqueueM(i)
		if c != 1 {
			t.Fatalf("expect the returninh overwrites is 1 BUT GOT %v.", c)
		}
		if errors.Is(err, ErrQueueFull) {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
		t.Logf("  %3d. ringbuf elements -> %v, overwrites=%v", i, rb, c)

		if i == size {
			_, err = rb.Dequeue()
			if err != nil {
				t.Fatalf("> %3d. expect err==nil but some error raised. err: %+v", i, err)
			}
			cap, c, err = rb.EnqueueMRich(i)
			if err != nil {
				t.Fatalf("> %3d. expect err==nil but some error raised. err: %+v", i, err)
			}
			if c != 0 {
				t.Fatalf("expect the returning overwrites is 0 BUT GOT %v.", c)
			}
			if cap != size {
				t.Fatalf("expect the returning cap is %v BUT GOT %v.", size, c)
			}
			t.Logf("  %3s. ringbuf elements -> %v, overwrites=%v, cap=%v", "", rb, c, cap)
		}
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
