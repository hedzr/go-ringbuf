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
	IsEmpty() (empty bool)
	IsFull() (full bool)
	Reset()
}

// RichOverlappedRingBuffer supplies a rich-measureable [EnqueueM] api.
type RichOverlappedRingBuffer[T any] interface {
	RingBuffer[T]

	// EnqueueM returns how many elements were overwritten by the
	// new incoming data, generally it would be 1 if overwriting
	// happened, or 0 for a normal insertion.
	// If error occurred, overwites field is undefined.
	EnqueueM(item T) (overwrites uint32, err error)
	// EnqueueMRich returns overwritten count of elements, and
	// current capacity of container.
	// If error occurred, both of these two fields are undefined.
	EnqueueMRich(item T) (size, overwrites uint32, err error)
}

// RingBuffer interface provides a set of standard ring buffer operations
type RingBuffer[T any] interface {
	Close()

	// Queue[T]

	Enqueue(item T) (err error)   // or [Put] as alternative
	Dequeue() (item T, err error) // or [Get] as alternative
	Cap() uint32                  // Cap returns the outer capacity of the ring buffer.
	CapReal() uint32              // CapReal returns the real (inner) capacity of the ring buffer.
	Size() uint32                 // Size returns the quantity of items in the ring buffer queue
	IsEmpty() (empty bool)
	IsFull() (full bool)
	Reset()

	Put(item T) (err error)
	Get() (item T, err error)

	Quantity() uint32 // Quantity returns the quantity of items in the ring buffer queue

	Debug(enabled bool) (lastState bool) // for internal debugging, see [Dbg] interface.
	ResetCounters()                      // for internal debugging, see [Dbg] interface.
}
