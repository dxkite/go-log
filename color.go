package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

var colorEnable = true

type colorWriter struct {
	writer
}

func NewColorWriter() io.Writer {
	return &colorWriter{writer{os.Stdout, TextMarshaler}}
}

func (w *colorWriter) WriteLogMessage(m *LogMessage) error {
	var msg []byte
	if v, err := w.fn(m); err != nil {
		return err
	} else {
		msg = v
	}
	_, err := w.ColorWrite(m.Level, msg)
	return err
}

func (w *colorWriter) ColorWrite(level LogLevel, msg []byte) (int, error) {
	if !colorEnable {
		return w.w.Write(msg)
	}
	var tpl = "%s"
	switch level {
	case Lerror:
		tpl = "\x1b[31;1m%s\x1b[0m"
	case Lwarn:
		tpl = "\x1b[33;1m%s\x1b[0m"
	case Linfo:
		tpl = "\x1b[36;1m%s\x1b[0m"
	case Ldebug:
	}
	n, err := w.w.Write([]byte(fmt.Sprintf(tpl, string(msg))))
	return n, err
}

func (w *colorWriter) Write(p []byte) (int, error) {
	m := new(LogMessage)
	if er := m.unmarshal(bytes.NewBuffer(p)); er != nil {
		// 解码失败
		return w.ColorWrite(Ldebug, p)
	}
	return len(p), w.WriteLogMessage(m)
}
