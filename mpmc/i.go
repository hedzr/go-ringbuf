package mpmc

// Queue interface provides a set of standard queue operations
type Queue[T any] interface {
	Enqueue(item T) (err error)
	Dequeue() (item T, err error)
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
type RingBuffer[T any] interface {
	Close()

	// Queue[T]

	Enqueue(item T) (err error)
	Dequeue() (item T, err error)
	// Cap returns the outer capacity of the ring buffer.
	Cap() uint32
	// CapReal returns the real (inner) capacity of the ring buffer.
	CapReal() uint32
	// Size returns the quantity of items in the ring buffer queue
	Size() uint32
	IsEmpty() (b bool)
	IsFull() (b bool)
	Reset()

	Put(item T) (err error)
	Get() (item T, err error)

	// Quantity returns the quantity of items in the ring buffer queue
	Quantity() uint32

	Debug(enabled bool) (lastState bool)

	ResetCounters()
}
