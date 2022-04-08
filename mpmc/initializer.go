package mpmc

// Initializeable data item supports lighter-weight clone operations.
type Initializeable[T any] interface {
	PreAlloc(index int) (newBlock T)
	CloneIn(srcBlock T, targetBlock *T)
	CloneOut(srcBlock *T) (targetBlock T)
}
