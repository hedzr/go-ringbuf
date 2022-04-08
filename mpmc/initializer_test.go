package mpmc

import (
	"testing"
)

type DI struct {
	intVal int
	ptrVal *DI
}

func (s *DI) PreAlloc(index int) (newBlock DI) {
	ni := DI{intVal: index}
	ni.ptrVal = s
	return ni
}

func (s *DI) CloneIn(srcBlock DI, targetBlock *DI) {
	targetBlock.intVal = srcBlock.intVal
	targetBlock.ptrVal = srcBlock.ptrVal
}

func (s *DI) CloneOut(srcBlock *DI) (targetBlock DI) {
	targetBlock.intVal = (srcBlock).intVal
	targetBlock.ptrVal = (srcBlock).ptrVal
	(srcBlock).intVal = -1 // while you want clear the data item block
	return
}

func (s *DI) Inc() DI {
	s.intVal++
	return *s
}

func checkDIResult(t *testing.T, desc string, rb1 *ringBuf[DI], got DI, expect int) {
	if got.intVal == expect {
		t.Logf("[%s] got = %v / %v | expected: %v", desc, got, got, expect)
	} else {
		t.Fatalf("[%s] got = %v | expected: %v | WRONG!!", desc, got, expect)
	}
}

func TestCustomItem(t *testing.T) {
	templ := &DI{1, nil}

	rb := newRingBuf[DI](4, WithItemInitializer[DI](templ))
	rb1 := rb // rb.(*ringBuf[DI])

	var err error
	var it DI
	var dataItem *DI

	if it, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}

	dataItem = &DI{0, templ}

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, "Enqueue(dataItem.Inc()) #1", rb1, 1)

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, "Enqueue(dataItem.Inc()) #2", rb1, 2)

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, "Enqueue(dataItem.Inc()) #3", rb1, 3)

	if err = rb.Enqueue(*dataItem); err != ErrQueueFull { // 3, full
		t.Fatal("expect full event")
	}

	it, err = rb.Dequeue() // 3 -> 2
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 2)
	checkDIResult(t, "Dequeue()", rb1, it, 1)

	err = rb.Enqueue(dataItem.Inc()) // 3
	// t.Log(rb1.qty(rb1.head, rb1.tail))  // wanted: 3
	checkerr(t, err)
	checkqty(t, "Enqueue(dataItem.Inc()) #4", rb1, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 2)
	checkDIResult(t, "Dequeue()", rb1, it, 2)

	err = rb.Enqueue(dataItem.Inc()) // 5
	checkerr(t, err)
	checkqty(t, "Enqueue(dataItem.Inc()) $5", rb1, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 2)
	checkDIResult(t, "Dequeue()", rb1, it, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 1)
	checkDIResult(t, "Dequeue()", rb1, it, 4)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, "Dequeue()", rb1, 0)
	checkDIResult(t, "Dequeue()", rb1, it, 5)

	if _, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}
}
