package fast

import "io"

type (
	// Queue interface provides a set of standard queue operations
	Queue interface {
		Enqueue(item interface{}) (err error)
		Dequeue() (item interface{}, err error)
		// Cap returns the outer capacity of the ring buffer.
		Cap() uint32
		// CapReal returns the real (inner) capacity of the ring buffer.
		CapReal() uint32
		// Size returns the quantity of items in the ring buffer queue
		Size() uint32
		IsEmpty() (b bool)
		IsFull() (b bool)
		Reset()
	}

	// RingBuffer interface provides a set of standard ring buffer operations
	RingBuffer interface {
		io.Closer // for logger

		Queue

		Put(item interface{}) (err error)
		Get() (item interface{}, err error)

		// Quantity returns the quantity of items in the ring buffer queue
		Quantity() uint32

		Debug(enabled bool) (lastState bool)

		ResetCounters()
	}
)
