package rb

import (
	"errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/cpu"
	"gopkg.in/natefinch/lumberjack.v2"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Opt func(buf *ringBuf)

func New(capacity uint32, opts ...Opt) RingBuffer {
	if !isInitialized() {
		return nil
	}

	size := roundToPower2(capacity)
	// logger := initLogger("all.log", "debug")
	logger := initLoggerConsole(zapcore.DebugLevel)
	rb := &ringBuf{
		data:       make([]rbItem, size),
		head:       0,
		tail:       0,
		cap:        size,
		capModMask: size - 1, // = 2^n - 1
		logger:     logger,
	}

	for _, opt := range opts {
		opt(rb)
	}

	if rb.debugMode {
		rb.logger.Debug("[ringbuf][INI] ", zap.Uint32("cap", rb.cap), zap.Uint32("capModMask", rb.capModMask))
	}

	for i := 0; i < (int)(size); i++ {
		rb.data[i].readWrite &= 0 // bit 0: readable, bit 1: writable
	}

	return rb
}

func WithDebugMode(debug bool) Opt {
	return func(buf *ringBuf) {
		buf.debugMode = debug
	}
}

// Dbg exposes some internal fields for debugging
type Dbg interface {
	GetGetWaits() uint64
	GetPutWaits() uint64
	Debug(enabled bool) (lastState bool)
	ResetCounters()
}

func (rb *ringBuf) GetGetWaits() uint64 {
	return atomic.LoadUint64(&rb.getWaits)
}

func (rb *ringBuf) GetPutWaits() uint64 {
	return atomic.LoadUint64(&rb.putWaits)
}

func (rb *ringBuf) ResetCounters() {
	atomic.StoreUint64(&rb.getWaits, 0)
	atomic.StoreUint64(&rb.putWaits, 0)
}

func (rb *ringBuf) Close() (err error) {
	if rb.logger != nil {
		err = rb.logger.Sync()
		rb.logger = nil
	}
	return
}

func (rb *ringBuf) qty(head, tail uint32) (quantity uint32) {
	if tail >= head {
		quantity = tail - head
	} else {
		quantity = rb.capModMask + (tail - head)
	}
	return quantity
}

func (rb *ringBuf) Quantity() uint32 {
	return rb.Size()
}

func (rb *ringBuf) Size() uint32 {
	var quantity uint32
	// head = atomic.LoadUint32(&rb.head)
	// tail = atomic.LoadUint32(&rb.tail)
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)

	if tail >= head {
		quantity = tail - head
	} else {
		quantity = rb.capModMask + (tail - head)
	}

	return quantity
}

func (rb *ringBuf) Cap() uint32 {
	return rb.cap
}

func (rb *ringBuf) IsEmpty() (b bool) {
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)
	// var tail, head uint32
	// head = atomic.LoadUint32(&rb.head)
	// tail = atomic.LoadUint32(&rb.tail)
	b = head == tail
	return
}

func (rb *ringBuf) IsFull() (b bool) {
	var tail, head uint32
	var quad uint64
	quad = atomic.LoadUint64((*uint64)(unsafe.Pointer(&rb.head)))
	head = (uint32)(quad & MaxUint32_64)
	tail = (uint32)(quad >> 32)
	// var tail, head uint32
	// head = atomic.LoadUint32(&rb.head)
	// tail = atomic.LoadUint32(&rb.tail)
	b = ((tail + 1) & rb.capModMask) == head
	return
}

func (rb *ringBuf) Debug(enabled bool) (lastState bool) {
	lastState = rb.debugMode
	rb.debugMode = enabled
	return
}

func roundToPower2(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func initLoggerConsole(l zapcore.Level) *zap.Logger {
	// alevel := zap.NewAtomicLevel()
	// http.HandleFunc("/handle/level", alevel.ServeHTTP)
	// logcfg.Level = alevel

	logcfg := zap.NewDevelopmentConfig()
	logcfg.Level = zap.NewAtomicLevelAt(l)
	logcfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	logger, _ := logcfg.Build()
	return logger
}

func initLogger(logpath string, loglevel string) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   logpath, // ⽇志⽂件路径
		MaxSize:    1024,    // megabytes
		MaxBackups: 3,       // 最多保留3个备份
		MaxAge:     7,       // days
		Compress:   true,    // 是否压缩 disabled by default
	}
	w := zapcore.AddSync(&hook)

	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)
	logger := zap.New(core)
	return logger
}

func init() {
	// fmt.Printf("CacheLinePadSize = %v\n", CacheLinePadSize)
	// fmt.Printf("ringBuf.Size = %v\n", unsafe.Sizeof(ringBuf{}))

	// logger, _ := zap.NewProduction()
	// logger, _ := zap.NewDevelopment()

	initializedOnce.Do(func() {

		ErrQueueFull = errors.New("queue full")
		ErrQueueEmpty = errors.New("queue empty")
		ErrRaced = errors.New("queue race")
		atomic.CompareAndSwapUint32(&initialized, 0, 1)

	})
}

func isInitialized() bool {
	return atomic.LoadUint32(&initialized) == 1
}

var initializedOnce sync.Once
var initialized uint32
var ErrQueueFull error
var ErrQueueEmpty error
var ErrRaced error

const CacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})

// const MaxUint = ^uint(0)
// const MinUint = 0
//
// const MaxUint16 = ^uint16(0)
// const MinUint16 = 0

const MaxUint32_64 = (uint64)(^uint32(0))

// const MaxUint32 = ^uint32(0)
// const MinUint32 = 0
//
// const MaxUint64 = ^uint64(0)
// const MinUint64 = 0
//
// const MaxInt = int(MaxUint >> 1)
// const MinInt = -MaxInt - 1
//
// const IntMAX = int(^uint(0) >> 1)
// const Int64MAX = int64(2) ^ 64 - 1
//
// const MaxInt16 = int(MaxUint16 >> 1)
// const MinInt16 = -MaxInt16 - 1
//
// const MaxInt32 = int(MaxUint32 >> 1)
// const MinInt32 = -MaxInt32 - 1
//
// const MaxInt64 = int(MaxUint64 >> 1)
// const MinInt64 = -MaxInt64 - 1
