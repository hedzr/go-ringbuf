/*
 * Copyright © 2020 Hedzr Yeh.
 */

package untitled_test

import (
	"testing"

	"gopkg.in/hedzr/go-ringbuf.v1/untitled"
)

// BenchmarkRingBuffer_Write tests
//
// go test -bench=. -run=^BenchmarkRingBuffer_Write$ -benchtime=4s ./...
func BenchmarkRingBuffer_Write(b *testing.B) {
	inputs := [][]byte{
		[]byte("hello world\n"),
		[]byte("this is a test\n"),
		[]byte("my cool input\n"),
	}
	var totalWritten int64 = 0

	// num := 10
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// fmt.Sprintf("%d", num)
		totalWritten += singleTest(inputs, b)
	}

	b.Logf("Total written bytes : %v", totalWritten)
}

func singleTest(inputs [][]byte, t *testing.B) (totalWritten int64) {
	// Write a bunch of data

	buf, err := untitled.NewBuffer(4)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for _, b := range inputs {
		n, err := buf.Write(b)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if n != len(b) {
			t.Fatalf("bad: %v", n)
		}
		totalWritten += int64(n)
	}

	// // Reset it
	// buf.Reset()
	//
	// // Write more data
	// input := []byte("hello")
	// n, err := buf.Write(input)
	// if err != nil {
	// 	t.Fatalf("err: %s", err)
	// }
	// if n != len(input) {
	// 	t.Fatalf("bad: %v", n)
	// }
	//
	// // Test the output
	// expect := []byte("ello")
	// if !bytes.Equal(buf.Bytes(), expect) {
	// 	t.Fatalf("bad: %v", string(buf.Bytes()))
	// }

	return
}
