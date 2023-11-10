package log

import (
	"bytes"
	"io"
)

type LogMarshaler func(m *LogMessage) ([]byte, error)

type writer struct {
	w    io.Writer
	fn   LogMarshaler
	last *LogMessage
}

func NewWriter(w io.Writer, fn LogMarshaler) io.Writer {
	return &writer{w, fn, nil}
}

func (w *writer) WriteLogMessage(m *LogMessage) error {
	var msg []byte
	if v, err := w.fn(m); err != nil {
		return err
	} else {
		msg = v
	}
	_, err := w.w.Write(msg)
	return err
}

func (w *writer) Write(p []byte) (int, error) {
	m := new(LogMessage)
	if er := m.unmarshal(bytes.NewBuffer(p)); er != nil {
		// 解码失败
		return w.w.Write(p)
	}
	return len(p), w.WriteLogMessage(m)
}

func NewTextWriter(w io.Writer) io.Writer {
	return NewWriter(w, TextMarshaler)
}

func NewJsonWriter(w io.Writer) io.Writer {
	return NewWriter(w, JsonMarshaler)
}

func NewColorWriter(w io.Writer) io.Writer {
	return NewWriter(w, ColorMarshaler)
}

func MultiWriter(writers ...io.Writer) io.Writer {
	allWriters := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if mw, ok := w.(*multiWriter); ok {
			allWriters = append(allWriters, mw.writers...)
		} else {
			allWriters = append(allWriters, w)
		}
	}
	return &multiWriter{allWriters}
}

type multiWriter struct {
	writers []io.Writer
}

func (t *multiWriter) WriteLogMessage(m *LogMessage) (err error) {
	for _, w := range t.writers {
		if vv, ok := w.(LogMessageWriter); ok {
			err = vv.WriteLogMessage(m)
		} else {
			_, err = w.Write(m.marshal())
		}
		if err != nil {
			return err
		}
	}
	return err
}

func (t *multiWriter) Write(p []byte) (n int, err error) {
	m := new(LogMessage)
	if er := m.unmarshal(bytes.NewBuffer(p)); er != nil {
		for _, w := range t.writers {
			if _, err = w.Write(m.marshal()); err != nil {
				return 0, err
			}
		}
	} else {
		return len(p), t.WriteLogMessage(m)
	}
	return len(p), err
}
