package untitled

import (
	"testing"
)

func TestBuffer_Verbose(t *testing.T) {
	_, _ = NewBuffer(-1, WithOverflowHandler(nil))

	a, _ := NewBuffer(3, WithOverflowHandler(func(rb *RingBuffer, bufWriting []byte, sizeTarget int64) {
		t.Logf("overflow")
	}))
	t.Logf("cap: %v, len: %v. q: %v", a.Capacity(), a.Len(), a)
	for i := 0; i < 5; i++ {
		_, _ = a.Write([]byte{1})
	}
}
