package fast

import (
	"golang.org/x/sys/cpu"
	"gopkg.in/hedzr/errors.v2"
	"sync"
	"sync/atomic"
	"unsafe"
)

func init() {
	// fmt.Printf("CacheLinePadSize = %v\n", CacheLinePadSize)
	// fmt.Printf("ringBuf.Size = %v\n", unsafe.Sizeof(ringBuf{}))

	// logger, _ := zap.NewProduction()
	// logger, _ := zap.NewDevelopment()

	initializedOnce.Do(func() {

		ErrQueueFull = errors.New("queue full")
		ErrQueueEmpty = errors.New("queue empty")
		ErrRaced = errors.New("queue race")
		ErrQueueNotReady = errors.New("queue not ready")
		atomic.CompareAndSwapUint32(&initialized, 0, 1)

	})
}

func isInitialized() bool {
	return atomic.LoadUint32(&initialized) == 1
}

//func initLoggerConsole(l zapcore.Level) *zap.Logger {
//	// alevel := zap.NewAtomicLevel()
//	// http.HandleFunc("/handle/level", alevel.ServeHTTP)
//	// logCfg.Level = alevel
//
//	logCfg := zap.NewDevelopmentConfig()
//	logCfg.Level = zap.NewAtomicLevelAt(l)
//	logCfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
//	logger, _ := logCfg.Build()
//	return logger
//}
//
//func initLogger(logPath string, loglevel string) *zap.Logger {
//	hook := lumberjack.Logger{
//		Filename:   logPath, // the logging file path
//		MaxSize:    1024,    // megabytes
//		MaxBackups: 3,       // 3 backups kept at most
//		MaxAge:     7,       // 7 days kept at most
//		Compress:   true,    // disabled by default
//	}
//	w := zapcore.AddSync(&hook)
//
//	var level zapcore.Level
//	switch loglevel {
//	case "debug":
//		level = zap.DebugLevel
//	case "info":
//		level = zap.InfoLevel
//	case "error":
//		level = zap.ErrorLevel
//	default:
//		level = zap.InfoLevel
//	}
//
//	encoderConfig := zap.NewProductionEncoderConfig()
//	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
//	core := zapcore.NewCore(
//		zapcore.NewConsoleEncoder(encoderConfig),
//		w,
//		level,
//	)
//	logger := zap.New(core)
//	return logger
//}

var initializedOnce sync.Once
var initialized uint32

// ErrQueueFull queue full when enqueueing
var ErrQueueFull error

// ErrQueueEmpty queue empty when dequeueing
var ErrQueueEmpty error

// ErrRaced the exception raised if data racing
var ErrRaced error

// ErrQueueNotReady queue not ready for enqueue or dequeue
var ErrQueueNotReady error

// CacheLinePadSize represents the CPU Cache Line Padding Size, compliant with the current running CPU Architect
const CacheLinePadSize = unsafe.Sizeof(cpu.CacheLinePad{})

// const MaxUint = ^uint(0)
// const MinUint = 0
//
// const MaxUint16 = ^uint16(0)
// const MinUint16 = 0

// MaxUint32_64 represents the maximal uint32 value
const MaxUint32_64 = (uint64)(^uint32(0))

// MaxUint64 represents the maximal uint64 value
const MaxUint64 = ^uint64(0)

// MaxUint32 represents the maximal uint32 value
const MaxUint32 = ^uint32(0)

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
