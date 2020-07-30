package fast

// Initializeable data item supports lighter-weight clone operations.
type Initializeable interface {
	PreAlloc(index int) (newBlock interface{})
	CloneIn(srcBlock, targetBlock interface{})
	CloneOut(srcBlock interface{}) (targetBlock interface{})
}
