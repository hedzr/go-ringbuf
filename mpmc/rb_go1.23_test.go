package mpmc

import (
	"testing"
)

// func (rb *ringBuf[T]) Iterator(yield func(v T) bool) bool {
// 	return
// }

func TestRingBuf_Only123(t *testing.T) { //nolint:revive
	size := roundUpToPower2(4)
	t.Logf("start, size = %v", size)
	q := &ringBuf[uint32]{
		data:       make([]rbItem[uint32], size),
		head:       0,
		tail:       0,
		cap:        size,
		capModMask: size - 1,
	}

	t.Logf("[go1.23] Test %v", q.Only123())
}
