package recorder

import "io"

type WriterMonitor struct {
	Writer  io.Writer
	OnWrite func(in []byte)
}

func (m *WriterMonitor) Write(in []byte) (n int, err error) {
	n = len(in)

	m.OnWrite(in)
	if m.Writer != nil {
		n, err = m.Writer.Write(in)
	}

	return
}

type ReaderMonitor struct {
	Reader io.Reader
	OnRead func(in []byte)
}

func (m *ReaderMonitor) Read(in []byte) (n int, err error) {
	n = len(in)

	if m.Reader != nil {
		n, err = m.Reader.Read(in)
		m.OnRead(in)
	}

	return
}
