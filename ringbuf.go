package ringbuf

import "github.com/hedzr/go-ringbuf/v2/mpmc"

// New returns the RingBuffer object
func New[T any](capacity uint32, opts ...mpmc.Opt[T]) mpmc.RingBuffer[T] {
	return mpmc.New[T](capacity, opts...)
}
