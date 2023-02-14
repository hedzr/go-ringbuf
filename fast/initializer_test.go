package fast

import (
	"fmt"
	"testing"
)

type DI struct {
	intVal int
	ptrVal *DI
}

func (s *DI) PreAlloc(index int) (newBlock interface{}) {
	ni := &DI{intVal: index}
	ni.ptrVal = s
	return ni
}

func (s *DI) CloneIn(srcBlock, targetBlock interface{}) {
	if di, ok := targetBlock.(*DI); ok {
		ss := srcBlock.(*DI)
		di.intVal = ss.intVal
		di.ptrVal = ss.ptrVal
	}
}

func (s *DI) CloneOut(srcBlock interface{}) (targetBlock interface{}) {
	if ss, ok := srcBlock.(*DI); ok {
		di := &DI{}
		targetBlock = di

		// light-weight extracting here:
		di.intVal = (ss).intVal
		di.ptrVal = (ss).ptrVal
		(ss).intVal = -1 // while you want clear the data item block
	} else {
		fmt.Printf("invalid srcBlockPtr: %v\n", srcBlock)
	}
	return
}

func (s *DI) Inc() *DI {
	s.intVal++
	return s
}

func checkDIResult(t *testing.T, rb1 *ringBuf, got interface{}, expect int) {
	if g, ok := got.(*DI); ok && g.intVal == expect {
		t.Logf("got = %v / %v | expected: %v", got, g, expect)
	} else {
		t.Fatalf("got = %v / %v | expected: %v | WRONG!!", got, g, expect)
	}
}

func TestCustomItem(t *testing.T) {
	templ := &DI{1, nil}

	rb := New(4, WithItemInitializer(templ))
	rb1 := rb.(*ringBuf)

	var err error
	var it interface{}
	var dataItem *DI

	if _, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}

	dataItem = &DI{0, templ}

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, rb1, 1)

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, rb1, 2)

	err = rb.Enqueue(dataItem.Inc())
	checkerr(t, err)
	checkqty(t, rb1, 3)

	if err = rb.Enqueue(dataItem); err != ErrQueueFull { // 3
		t.Fatal("expect full event")
	}

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)
	checkDIResult(t, rb1, it, 1)

	err = rb.Enqueue(dataItem.Inc()) // 4
	// t.Log(rb1.qty(rb1.head, rb1.tail))  // wanted: 3
	checkerr(t, err)
	checkqty(t, rb1, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)
	checkDIResult(t, rb1, it, 2)

	err = rb.Enqueue(dataItem.Inc()) // 5
	checkerr(t, err)
	checkqty(t, rb1, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 2)
	checkDIResult(t, rb1, it, 3)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 1)
	checkDIResult(t, rb1, it, 4)

	it, err = rb.Dequeue()
	checkerr(t, err)
	checkqty(t, rb1, 0)
	checkDIResult(t, rb1, it, 5)

	if _, err = rb.Dequeue(); err != ErrQueueEmpty {
		t.Fatal("expect empty event")
	}
}
