package main

import (
	"fmt"
	"github.com/hedzr/go-ringbuf/v2"
	"log"
)

func main() {
	testIntRB()
	testStringRB()
}

func testStringRB() {
	var err error
	var rb = ringbuf.New[string](80)
	err = rb.Enqueue("abcde")
	errChk(err)

	var item string
	item, err = rb.Dequeue()
	errChk(err)
	fmt.Printf("dequeue ok: %v\n", item)
}

func testIntRB() {
	var err error
	var rb = ringbuf.New[int](80)
	err = rb.Enqueue(3)
	errChk(err)

	var item int
	item, err = rb.Dequeue()
	errChk(err)
	fmt.Printf("dequeue ok: %v\n", item)
}

func errChk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
