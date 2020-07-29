package fast

import (
	"github.com/hedzr/log"
	log2 "log"
	"testing"
)

func TestResets(t *testing.T) {
	logger := log.NewDummyLogger()

	log2.Printf("")

	rb := New(NLtd,
		WithDebugMode(true),
		WithLogger(logger),
	)
	defer rb.Close()
	rb.ResetCounters()
	rb.Reset()
	_ = rb.Put(3)
	x := rb.(*ringBuf)
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

func checkqty(t *testing.T, rb1 *ringBuf, expect uint32) {
	if rb1.Quantity() != rb1.qty(rb1.head, rb1.tail) || rb1.Quantity() != expect {
		t.Fatalf("qty = %v / %v | expected: %v | WRONG!!", rb1.Quantity(), rb1.qty(rb1.head, rb1.tail), expect)
	} else {
		t.Logf("qty = %v / %v | expected: %v", rb1.Quantity(), rb1.qty(rb1.head, rb1.tail), expect)
	}
}

func TestRoundedQty(t *testing.T) {
	rb := New(4)
	rb1 := rb.(*ringBuf)

	var err error

	if _, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}

	err = rb.Enqueue(1)
	checkerr(t, err)
	checkqty(t, rb1, 1)

	err = rb.Enqueue(2)
	checkerr(t, err)
	checkqty(t, rb1, 2)

	err = rb.Enqueue(3)
	checkerr(t, err)
	checkqty(t, rb1, 3)

	if err = rb.Enqueue(3); err != ErrQueueFull {
		t.Fatal("expect full event")
	}

	_, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)

	err = rb.Enqueue(4)
	// t.Log(rb1.qty(rb1.head, rb1.tail))  // wanted: 3
	checkerr(t, err)
	checkqty(t, rb1, 3)

	_, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)

	err = rb.Enqueue(5)
	checkerr(t, err)
	checkqty(t, rb1, 3)

	_, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)

	_, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 1)

	_, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 0)

	if _, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}
}
