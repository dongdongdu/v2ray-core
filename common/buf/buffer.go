// Package buf provides a light-weight memory allocation mechanism.
package buf

import (
	"io"
)

// Supplier is a writer that writes contents into the given buffer.
type Supplier func([]byte) (int, error)

// Buffer is a recyclable allocation of a byte array. Buffer.Release() recycles
// the buffer into an internal buffer pool, in order to recreate a buffer more
// quickly.
type Buffer struct {
	v    []byte
	pool Pool

	start int
	end   int
}

// Release recycles the buffer into an internal buffer pool.
func (b *Buffer) Release() {
	if b == nil || b.v == nil {
		return
	}
	if b.pool != nil {
		b.pool.Free(b)
	}
	b.v = nil
	b.pool = nil
}

// Clear clears the content of the buffer, results an empty buffer with
// Len() = 0.
func (b *Buffer) Clear() {
	b.start = 0
	b.end = 0
}

// AppendBytes appends one or more bytes to the end of the buffer.
func (b *Buffer) AppendBytes(bytes ...byte) {
	b.Append(bytes)
}

// Append appends a byte array to the end of the buffer.
func (b *Buffer) Append(data []byte) {
	nBytes := copy(b.v[b.end:], data)
	b.end += nBytes
}

// AppendSupplier appends the content of a BytesWriter to the buffer.
func (b *Buffer) AppendSupplier(writer Supplier) error {
	nBytes, err := writer(b.v[b.end:])
	b.end += nBytes
	return err
}

// Byte returns the bytes at index.
func (b *Buffer) Byte(index int) byte {
	return b.v[b.start+index]
}

// SetByte sets the byte value at index.
func (b *Buffer) SetByte(index int, value byte) {
	b.v[b.start+index] = value
}

// Bytes returns the content bytes of this Buffer.
func (b *Buffer) Bytes() []byte {
	return b.v[b.start:b.end]
}

// Reset resets the content of the Buffer with a supplier.
func (b *Buffer) Reset(writer Supplier) error {
	b.start = 0
	nBytes, err := writer(b.v[b.start:])
	b.end = b.start + nBytes
	return err
}

// BytesRange returns a slice of this buffer with given from and to bounary.
func (b *Buffer) BytesRange(from, to int) []byte {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start+from : b.start+to]
}

// BytesFrom returns a slice of this Buffer starting from the given position.
func (b *Buffer) BytesFrom(from int) []byte {
	if from < 0 {
		from += b.Len()
	}
	return b.v[b.start+from : b.end]
}

// BytesTo returns a slice of this Buffer from start to the given position.
func (b *Buffer) BytesTo(to int) []byte {
	if to < 0 {
		to += b.Len()
	}
	return b.v[b.start : b.start+to]
}

// Slice cuts the buffer at the given position.
func (b *Buffer) Slice(from, to int) {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	if to < from {
		panic("Invalid slice")
	}
	b.end = b.start + to
	b.start += from
}

// SliceFrom cuts the buffer at the given position.
func (b *Buffer) SliceFrom(from int) {
	if from < 0 {
		from += b.Len()
	}
	b.start += from
}

// Len returns the length of the buffer content.
func (b *Buffer) Len() int {
	if b == nil {
		return 0
	}
	return b.end - b.start
}

// IsEmpty returns true if the buffer is empty.
func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

// IsFull returns true if the buffer has no more room to grow.
func (b *Buffer) IsFull() bool {
	return b.end == len(b.v)
}

// Write implements Write method in io.Writer.
func (b *Buffer) Write(data []byte) (int, error) {
	nBytes := copy(b.v[b.end:], data)
	b.end += nBytes
	return nBytes, nil
}

// Read implements io.Reader.Read().
func (b *Buffer) Read(data []byte) (int, error) {
	if b.Len() == 0 {
		return 0, io.EOF
	}
	nBytes := copy(data, b.v[b.start:b.end])
	if nBytes == b.Len() {
		b.Clear()
	} else {
		b.start += nBytes
	}
	return nBytes, nil
}

// String returns the string form of this Buffer.
func (b *Buffer) String() string {
	return string(b.Bytes())
}

// New creates a Buffer with 8K bytes of arbitrary content.
func New() *Buffer {
	return mediumPool.Allocate()
}

// NewSmall returns a buffer with 2K bytes capacity.
func NewSmall() *Buffer {
	return smallPool.Allocate()
}

// NewLocal creates and returns a buffer on current thread.
func NewLocal(size int) *Buffer {
	return &Buffer{
		v:    make([]byte, size),
		pool: nil,
	}
}
