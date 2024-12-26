package term

import "io"

const bufferSize = 10 * 1024

// Buffer is a simple in-memory buffer that can be used as an io.Reader or io.Writer.
// It's like bytes.Buffer, but it reads and writes to a channel instead of a byte slice.
// So the read and write operations can block until data is available.
// One of the NewBuffer* functions should be used to create a new buffer.
// The Close method should be called to notify readers that no more data will be written.
type Buffer struct {
	ch  chan string
	str string
	pos int
}

// Read reads data from the channel and returns it in p. It will block until data
// is available or the channel is closed.
func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.pos >= len(b.str) {
		str, ok := <-b.ch
		if !ok {
			return 0, io.EOF
		}
		b.str = str
		b.pos = 0
	}

	// Copy data from current string to p
	n = copy(p, b.str[b.pos:])
	b.pos += n
	return n, nil
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	return b.WriteString(string(p))
}

func (b *Buffer) Close() error {
	close(b.ch)
	return nil
}

func (b *Buffer) WriteString(s string) (n int, err error) {
	b.ch <- s
	return len(s), nil
}

func (b *Buffer) String() string {
	if b == nil {
		return "<nil>"
	}
	bytes, err := io.ReadAll(b)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func NewBuffer() *Buffer {
	return &Buffer{
		ch: make(chan string, bufferSize),
	}
}

func NewBufferString(s string) *Buffer {
	b := NewBuffer()
	b.ch <- s
	return b
}

func NewBufferChan(ch chan string) *Buffer {
	return &Buffer{
		ch: ch,
	}
}

func NewBufferSize(size int) *Buffer {
	return &Buffer{
		ch: make(chan string, size),
	}
}
