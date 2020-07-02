# go-ringbuf

`go-ringbuf` provides a high-performance, lock-free circular queue (ring buffer) implementation in golang.


## Getting Start

### Import

```go
import "github.com/hedzr/ringbuf/fast"

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
```








## Contrib

Welcome


## LICENSE

MIT
