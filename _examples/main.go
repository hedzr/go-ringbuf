package main

import (
	"fmt"
	"log"

	"gopkg.in/hedzr/go-ringbuf.v1/fast"
)

func main() {
	var err error
	var rb = fast.New(80)
	err = rb.Enqueue(3)
	errChk(err)

	var item interface{}
	item, err = rb.Dequeue()
	errChk(err)
	fmt.Printf("dequeue ok: %v\n", item)
}

func errChk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
