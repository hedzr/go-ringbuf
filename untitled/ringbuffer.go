/*
 * Copyright © 2020 Hedzr Yeh.
 */

package untitled

import "errors"

// RingBuffer implements a circular buffer. It is a fixed size,
// and new writes overwrite older data, such that for a buffer
// of size N, for any amount of writes, only the last N bytes
// are retained.
type RingBuffer struct {
	data            []byte
	writeCursor     int64
	writtenCount    int64
	overflowHandler OverflowHandler
}

// RingBufferOption gives the options prototypes for NewBuffer()
type RingBufferOption func(*RingBuffer)

// NewBuffer creates a new buffer of a given size. The size
// must be greater than 0.
func NewBuffer(size int64, opts ...RingBufferOption) (*RingBuffer, error) {
	if size <= 0 {
		return nil, errors.New("size must be positive (>0)")
	}

	b := &RingBuffer{
		data: make([]byte, size),
	}

	for _, opt := range opts {
		opt(b)
	}

	return b, nil
}

// WithOverflowHandler can handle ring buffer overflow event
func WithOverflowHandler(oh OverflowHandler) RingBufferOption {
	return func(buffer *RingBuffer) {
		buffer.overflowHandler = oh
	}
}

// OverflowHandler is a user-defined callback, it will be triggered
// if writer is writing large data into ring buffer and making it
// full.
type OverflowHandler func(rb *RingBuffer, bufWriting []byte, sizeTarget int64)

// Write writes up to len(buf) bytes to the internal ring,
// overriding older data if necessary.
//
// Write 写入buf到ringbuffer内部
// 如果需要会覆盖旧数据(fifo)
//
// cases:
//   1. buf.len < rb.size && without rewind
//   2. buf.len < rb.size && rewind
//   3. buf.len >= rb.size: buf内容按照rb.size取模，剩余部分写入rb，同时触发overflow事件
func (b *RingBuffer) Write(buf []byte) (int, error) {

	// Account for total bytes written
	size, n := b.Size(), len(buf)
	n64 := int64(n)
	b.writtenCount += n64

	// If the buffer is larger than ours, then we only care
	// about the last size bytes anyways
	// 如果buf的大小超过容量限制,根据fifo原则
	// 我们只关注最近最新的部分数据
	if n64 > size {
		if b.overflowHandler != nil {
			b.overflowHandler(b, b.data, n64)
		}
		buf = buf[n64-size:]
	}

	remain := size - b.writeCursor
	copy(b.data[b.writeCursor:], buf)
	if n64 > remain {
		copy(b.data, buf[remain:])
	}

	b.writeCursor = (b.writeCursor + int64(len(buf))) % size
	return n, nil
}

// Size returns the size/capacity of the buffer
//
// 采用额外一字节作为头尾指针冲撞检测位置，因此rb的有效尺寸为 Size() - 1
func (b *RingBuffer) Size() int64 {
	return int64(len(b.data))
}

// Capacity returns the size/capacity of the buffer
func (b *RingBuffer) Capacity() int64 {
	return int64(len(b.data))
}

// Len returns the available length of the buffer
func (b *RingBuffer) Len() int64 {
	return int64(len(b.data)) - 1
}

// Bytes provides a slice of the bytes written. This
// slice should not be written to.
func (b *RingBuffer) Bytes() []byte {
	size := b.Size()
	switch {
	case b.writtenCount >= size && b.writeCursor == 0:
		return b.data

	case b.writtenCount > size:
		out := make([]byte, size)
		copy(out, b.data[b.writeCursor:])
		copy(out[size-b.writeCursor:], b.data[:b.writeCursor])
		return out

	default:
		return b.data[:b.writeCursor]
	}
}

// String returns the contents of the buffer as a string
func (b *RingBuffer) String() string {
	return string(b.Bytes())
}

// Reset resets the buffer so it has no content.
func (b *RingBuffer) Reset() {
	b.writeCursor = 0
	b.writtenCount = 0
}

// TotalWritten provides the total number of bytes written
func (b *RingBuffer) TotalWritten() int64 {
	return b.writtenCount
}
