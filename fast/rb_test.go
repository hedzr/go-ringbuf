package fast

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const N = 100
const NLtd = 16

func TestRingBuf_Put_OneByOne(t *testing.T) {
	var err error
	rb := New(NLtd, WithDebugMode(true))
	size := rb.Cap() - 1
	// t.Logf("Ring Buffer created, capacity = %v, real size: %v", size+1, size)
	defer rb.Close()

	for i := uint32(0); i < size; i++ {
		err = rb.Enqueue(i)
		if err != nil {
			t.Fatalf("faild on i=%v. err: %+v", i, err)
		} else {
			// t.Logf("%5d. '%v' put, quantity = %v.", i, i, fast.Quantity())
		}
	}

	for i := uint32(size); i < uint32(size)+6; i++ {
		err = rb.Enqueue(i)
		if err != ErrQueueFull {
			t.Fatalf("> %3d. expect ErrQueueFull but no error raised. err: %+v", i, err)
		}
	}

	var it interface{}
	for i := 0; ; i++ {

		it, err = rb.Dequeue()
		if err != nil {
			if err == ErrQueueEmpty {
				break
			}
			t.Fatalf("faild on i=%v. err: %+v. item: %v", i, err, it)
		} else {
			// t.Logf("< %3d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
		}
	}
}

//
// go test ./... -race -run '^TestRingBuf_Put_Randomized$'
// go test ./ringbuf/fast -race -run '^TestRingBuf_Put_R.*'
//
func TestRingBuf_Put_Randomized(t *testing.T) {
	var maxN = NLtd * 1000
	putRandomized(t, maxN, NLtd, func(r RingBuffer) {
		// r.Debug(true)
	})
}

//
// go test ./ringbuf/fast -race -bench=. -run=none
// go test ./ringbuf/fast -race -bench '.*Put128' -run=none
//
// go test ./ringbuf/fast -race -bench 'BenchmarkRingBuf' -run=none
//
// go test ./ringbuf/fast -race -bench=. -run=none -benchtime=3s
//
// go test ./ringbuf/fast -race -bench=. -benchmem -cpuprofile profile.out
// go test ./ringbuf/fast -race -bench=. -benchmem -memprofile memprofile.out -cpuprofile profile.out
// go tool pprof profile.out
//
// https://my.oschina.net/solate/blog/3034188
//
func BenchmarkRingBuf_Put16384(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 10000)
}

func BenchmarkRingBuf_Put1024(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 1000)
}

func BenchmarkRingBuf_Put128(b *testing.B) {
	b.ResetTimer()
	putRandomized(b, b.N, 100)
}

func putRandomized(t testing.TB, maxN int, queueSize uint32, opts ...func(r RingBuffer)) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var d1Closed int32
	d1, d2 := make(chan struct{}), make(chan struct{})

	_, noDebug := t.(*testing.B)
	rb := New(queueSize, WithDebugMode(!noDebug))
	for _, cb := range opts {
		cb(rb)
	}
	noDebug = !rb.Debug(!noDebug)
	t.Logf("Ring Buffer created, size = %v. maxN = %v, dbg: %v", rb.Cap(), maxN, !noDebug)
	defer rb.Close()

	go enqueueRoutine(t, rb, noDebug, &d1Closed, d1, maxN)
	go dequeueRoutine(t, rb, noDebug, &d1Closed, d2)

	<-d1
	<-d2

	if x, ok := rb.(Dbg); ok {
		t.Logf("Waits: get: %v, put: %v", x.GetGetWaits(), x.GetPutWaits())
	}
}

func enqueueRoutine(t testing.TB, rb RingBuffer, noDebug bool, d1Closed *int32, d1 chan struct{}, maxN int) {

	var err error
	for i, retry := 0, 1; i < maxN; i++ {
	retryPut:
		err = rb.Enqueue(i)
		if err != nil {
			if err == ErrQueueFull {
				// block till queue not full
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if retry > 1000 {
					if !noDebug && retry > 1000 {
						fmt.Printf("[PUT] %5d (retry %v). quantity = %v. FULL! block till queue not full.\n", i, retry, rb.Quantity())
					}
					break
				}
				goto retryPut
			}

			close(d1)
			atomic.StoreInt32(d1Closed, 1)
			t.Fatalf("[PUT] failed on i=%v. err: %+v.", i, err)
		}

		// fmt.Printf("[PUT] %5d. '%v' put, quantity = %v.\n", i, i, fast.Quantity())
		// t.Logf("[PUT] %5d. '%v' put, quantity = %v.", i, i, fast.Quantity())
		// time.Sleep(50 * time.Millisecond)
		retry = 1
	}
	close(d1)
	atomic.StoreInt32(d1Closed, 1)
	// t.Log("[PUT] END")
}

func dequeueRoutine(t testing.TB, rb RingBuffer, noDebug bool, d1Closed *int32, d2 chan struct{}) {

	d := time.Duration(rand.Intn(1000)+1000) * time.Millisecond
	time.Sleep(d)
	fmt.Printf("qty: %v\n", rb.Quantity())

	var err error
	var it interface{}
	var fetched []int
	// t.Logf("[GET] d1Closed: %v, err: %v", d1Closed, err)
	for i, retry := 0, 1; err != ErrQueueEmpty || err == nil; i++ {
	retryGet:
		it, err = rb.Dequeue()
		if err != nil {
			if err == ErrQueueEmpty {
				// block till queue not empty
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if !noDebug && retry > 2000 {
					fmt.Printf("[GET] %5d (retry %v). quantity = %v, fetched = %v. EMPTY! block till queue not empty.\n", i, retry, rb.Quantity(), len(fetched))
				}
				goto retryGet
			}
			t.Fatalf("[GET] failed on i=%v. err: %+v.", i, err)
		}

		retry = 1
		fetched = append(fetched, it.(int))
		// t.Logf("[GET] %5d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
		// time.Sleep(50 * time.Millisecond)

		if atomic.LoadInt32(d1Closed) == 1 {
			break
		}
	}
	close(d2)
	// t.Log("[GET] END")

	// checking the fetched
	last := fetched[0]
	for i := 1; i < len(fetched); i++ {
		ix := fetched[i]
		if ix != last+1 {
			t.Fatalf("[GET] the fetched sequence is wrong, expecting %v but got %v.", last+1, ix)
		}
		last = ix
	}

}

// multiple producers
func TestRingBuf_MPut(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	mtQueuePutGet(t, 1000, 1024*1024)
}

// multiple producers (long testing)
func TestRingBuf_MPutLong(t *testing.T) {
	if !testing.Short() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		mtQueuePutGet(t, 10000, 1024*1024)
	}
}

func mtQueuePutGet(t *testing.T, cnt int, qSize uint32) {
	// cnt := 10000
	sum, loops := 0, runtime.NumCPU()*4
	// fmt.Printf("--- TEST: cnt=%v, loops=%v.\n", cnt, loops)
	// start := time.Now()
	var putD, getD, lossTime time.Duration
	var putRetry, getRetry int64
	for i := 0; i <= loops; i++ {
		sum += i * cnt
		put, get, lt, pr, gr := mTestQueuePutGet(t, i, cnt, qSize)
		putD += put
		getD += get
		getRetry += gr
		putRetry += pr
		lossTime += lt
		// fmt.Printf("--- TEST %d/%d: putD = %v, getD = %v.\n", i, loops, putD, getD)
	}
	// end := time.Now()
	// use := end.Sub(start.Add(lossTime))
	use := putD + getD
	// fmt.Printf("--- TEST: use = %v.\n", use)
	op := use / time.Duration(sum)
	t.Logf("Grp: %d, Times: %d, use: %v, %v/op", loops, sum, use, op)
	t.Logf("Put: %d, use: %v, %v/op | retry times: %d", sum, putD, putD/time.Duration(sum), putRetry)
	t.Logf("Get: %d, use: %v, %v/op | retry times: %d", sum, getD, getD/time.Duration(sum), getRetry)
}

func mTestQueuePutGet(t testing.TB, grp, cnt int, qSize uint32) (put, get, lossTime time.Duration, putRetries, getRetries int64) {
	var wgPut, wgGet sync.WaitGroup
	var putI, getI int64
	var pc, gc int32
	var id int32
	var ltPutMax, ltGetMax time.Duration

	// fmt.Printf("    grp: %v, cnt: %v, qSize: %v\n", grp, cnt, qSize)
	rb := New(qSize) // , WithDebugMode(true))
	wgPut.Add(grp)
	wgGet.Add(grp)
	for i := 0; i < grp; i++ {
		// t.Logf("- [W] grp: %v, cnt: %v", grp, cnt)
		go putterRoutine(i, t, rb, &wgPut, &ltPutMax, &id, &pc, &putI, &putRetries, cnt)
	}

	// fmt.Printf("[R] -------------- grp: %v...\n", grp)
	for i := 0; i < grp; i++ {
		// t.Logf("- [R] i: %v, grp: %v, cnt: %v", i, grp, cnt)
		go func(g int) {
			defer wgGet.Done()
			d := time.Duration(rand.Intn(1500)+1500) * time.Millisecond
			time.Sleep(d)
			m := atomic.LoadInt64((*int64)(&ltGetMax))
			if int64(d) > m {
				m = int64(d) - m
				atomic.AddInt64((*int64)(&ltGetMax), m)
			}

			var err error
			start := time.Now()
			for j, retry := 0, int64(1); j < cnt; j++ {
			retryGet:
				_, err = rb.Dequeue()
				if err != nil {
					if err == ErrQueueEmpty {
						// block till queue not empty
						retry++
						if retry > 2000 {
							fmt.Printf("[R][ERR] retry failed. grp=%v, j=%v, cnt=%v.\n", g, j, cnt)
							break
						}
						time.Sleep(time.Duration(retry) * time.Microsecond)
						goto retryGet
					}
					t.Fatalf("[GET] failed on i=%v. err: %+v.", g, err)
				}
				atomic.AddInt32(&gc, 1)
				atomic.AddInt64(&getRetries, retry-1)
				retry = 1
			}
			// fmt.Printf("[R] i/grp: %v. done.\n", g)
			end := time.Now()
			getTime := end.Sub(start)
			atomic.AddInt64(&getI, int64(getTime))
		}(i)
	}

	wgGet.Wait()
	wgPut.Wait()

	lossTime += ltPutMax
	lossTime += ltGetMax

	if pc != gc {
		t.Errorf("put count (%v) != get count (%v)", pc, gc)
	}
	if q := rb.Quantity(); q != 0 {
		t.Errorf("Grp:%v, Quantity Error: [%v] <>[%v]", grp, q, 0)
	}

	put, get = time.Duration(putI), time.Duration(getI)

	// fmt.Printf("    [R] -------------- grp: %v END (put:%v, get: %v).\n", grp, put, get)
	return
}

func putterRoutine(g int, t testing.TB, rb RingBuffer, wgPut *sync.WaitGroup, ltPutMax *time.Duration, id, pc *int32, putI, putRetries *int64, cnt int) {
	defer wgPut.Done()

	d := time.Duration(rand.Intn(1000)+10) * time.Millisecond
	time.Sleep(d)
	m := atomic.LoadInt64((*int64)(ltPutMax))
	if int64(d) > m {
		m = int64(d) - m
		atomic.AddInt64((*int64)(ltPutMax), m)
	}

	var err error
	start := time.Now()
	for j, retry := 0, int64(1); j < cnt; j++ {
		val := fmt.Sprintf("Node.%d.%d.%d", g, j, atomic.AddInt32(id, 1))
	retryPut:
		err = rb.Enqueue(&val)
		if err != nil {
			if err == ErrQueueFull {
				// block till queue not full
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if retry > 1000 {
					fmt.Printf("[W][ERR] retry failed. val=%q.\n", val)
					break
				}
				goto retryPut
			}
			t.Fatalf("[PUT] failed on i=%v. err: %+v.", g, err)
		}
		atomic.AddInt32(pc, 1)
		atomic.AddInt64(putRetries, retry-1)
		retry = 1
	}
	// fmt.Printf("[W] i/grp: %v. done.\n", g)
	end := time.Now()
	putTime := end.Sub(start)
	atomic.AddInt64(putI, int64(putTime))
}

//
// go test ./ringbuf/fast -v -race -run 'TestQueuePutGetLong'
// go test ./ringbuf/fast -v -race -run '^TestQueuePutGetLong$'
// go test ./ringbuf/fast -v -race -run '^TestQueuePutGetLong.*'
//
// Benchmark testing manually for comparing (long testing)
func TestQueuePutGetLong(t *testing.T) {
	if !testing.Short() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		tQueuePutGet(t, 10000, 1024*1024)
	}
}

// // go test ./ringbuf/fast -v -race -run 'TestQueuePutGet'
// func TestQueuePutGetV(t *testing.T) {
// 	runtime.GOMAXPROCS(runtime.NumCPU())
// 	tQueuePutGet(t, 10000, 1024*1024)
// }

func tQueuePutGet(t *testing.T, cnt int, qSize uint32) {
	// cnt := 10000
	sum, loops := 0, runtime.NumCPU()*4
	// fmt.Printf("--- TEST: cnt=%v, loops=%v.\n", cnt, loops)
	start := time.Now()
	var putD, getD time.Duration
	for i := 0; i <= loops; i++ {
		sum += i * cnt
		put, get := testQueuePutGet(t, i, cnt, qSize)
		putD += put
		getD += get
		// fmt.Printf("--- TEST %d/%d: putD = %v, getD = %v.\n", i, loops, putD, getD)
	}
	end := time.Now()
	use := end.Sub(start)
	// fmt.Printf("--- TEST: use = %v.\n", use)
	op := use / time.Duration(sum)
	t.Logf("Grp: %d, Times: %d, use: %v, %v/op", loops, sum, use, op)
	t.Logf("Put: %d, use: %v, %v/op", sum, putD, putD/time.Duration(sum))
	t.Logf("Get: %d, use: %v, %v/op", sum, getD, getD/time.Duration(sum))
}

//
// go test ./ringbuf/fast -race -bench 'BenchmarkPutGet' -run=none -benchtime=10s -v
//
// go test ./ringbuf/fast -race -bench='.*PutGet16384$' -run=none
//
func BenchmarkPutGet(b *testing.B) {
	b.ResetTimer()
	bQueuePutGet(b, b.N, runtime.NumCPU()*4)
}

func bQueuePutGet(b testing.TB, cnt, loops int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sum := 0
	// fmt.Printf("--- TEST: cnt=%v, loops=%v.\n", cnt, loops)
	start := time.Now()
	var putD, getD time.Duration
	for i := 0; i <= loops; i++ {
		sum += (i) * cnt
		put, get := testQueuePutGet(b, i, cnt, 1024*1024)
		putD += put
		getD += get
		// fmt.Printf("--- TEST: putD = %v, getD = %v.\n", putD, getD)
	}
	end := time.Now()
	// fmt.Printf("--- TEST: use = %v.\n", end.Sub(start))
	use := end.Sub(start)
	op := use / time.Duration(sum)
	b.Logf("Grp: %d, Times: %d, use: %v, %v/op | cnt = %v", runtime.NumCPU()*4, sum, use, op, cnt)
	b.Logf("Put: %d, use: %v, %v/op", sum, putD, putD/time.Duration(sum))
	b.Logf("Get: %d, use: %v, %v/op", sum, getD, getD/time.Duration(sum))
}

func testQueuePutGet(t testing.TB, grp, cnt int, qSize uint32) (put, get time.Duration) {
	var wg sync.WaitGroup
	var id int32
	rb := New(qSize) // , WithDebugMode(true))
	// fmt.Printf("    grp: %v, cnt: %v, qSize: %v\n", grp, cnt, qSize)
	wg.Add(grp)
	start := time.Now()
	for i := 0; i < grp; i++ {
		// t.Logf("- [W] grp: %v, cnt: %v", grp, cnt)
		go func(g int) {
			defer wg.Done()
			var err error
			for j, retry := 0, 1; j < cnt; j++ {
				val := fmt.Sprintf("Node.%d.%d.%d", g, j, atomic.AddInt32(&id, 1))
			retryPut:
				err = rb.Enqueue(&val)
				if err != nil {
					if err == ErrQueueFull {
						// block till queue not full
						time.Sleep(time.Duration(retry) * time.Microsecond)
						retry++
						if retry > 1000 {
							fmt.Printf("[W][ERR] retry failed. val=%q.\n", val)
							break
						}
						goto retryPut
					}
					t.Fatalf("[PUT] failed on i=%v. err: %+v.", g, err)
				}
				retry = 1
			}
			// fmt.Printf("[W] i/grp: %v. done.\n", g)
		}(i)
	}
	wg.Wait()
	end := time.Now()
	put = end.Sub(start)

	// fmt.Printf("[R] -------------- grp: %v...\n", grp)
	var wg2 sync.WaitGroup
	wg2.Add(grp)
	start = time.Now()
	for i := 0; i < grp; i++ {
		// t.Logf("- [R] i: %v, grp: %v, cnt: %v", i, grp, cnt)
		go func(g int) {
			defer wg2.Done()
			var err error
			for j, retry := 0, 1; j < cnt; j++ {
			retryGet:
				_, err = rb.Get()
				if err != nil {
					if err == ErrQueueEmpty {
						// block till queue not empty
						// if !noDebug {
						// t.Logf("[GET] %5d. quantity = %v. EMPTY! block till queue not empty", i, fast.Quantity())
						// }
						time.Sleep(time.Duration(retry) * time.Microsecond)
						retry++
						if retry > 1000 {
							fmt.Printf("[R][ERR] retry failed. grp=%v, j=%v.\n", g, j)
							break
						}
						goto retryGet
					}
					t.Fatalf("[GET] failed on i=%v. err: %+v.", g, err)
				}
				retry = 1
			}
			// fmt.Printf("[R] i/grp: %v. done.\n", g)
		}(i)
	}
	wg2.Wait()
	end = time.Now()
	get = end.Sub(start)
	if q := rb.Quantity(); q != 0 {
		t.Errorf("Grp:%v, Quantity Error: [%v] <>[%v]", grp, q, 0)
	}

	// fmt.Printf("    [R] -------------- grp: %v END (put:%v, get: %v).\n", grp, put, get)
	return
}

func BenchmarkHello(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("hello")
	}
}
