package log

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

const TimeFormat = "2006-01-02 15:04:05.000"

func TextMarshaler(m *LogMessage) ([]byte, error) {
	var msg string
	if len(m.File) > 0 {
		f := filepath.Base(m.File)
		msg = fmt.Sprintf("%s [%-5s] %s:%d %s", m.Time.Format(TimeFormat), m.Level, f, m.Line, m.Message)
	} else {
		msg = fmt.Sprintf("%s [%-5s] %s", m.Time.Format(TimeFormat), m.Level, m.Message)
	}
	return []byte(msg), nil
}

func JsonMarshaler(m *LogMessage) ([]byte, error) {
	if len(m.File) > 0 {
		m.File = filepath.Base(m.File)
	}
	if b, er := json.Marshal(m); er != nil {
		return nil, er
	} else {
		return []byte(string(b) + "\n"), nil
	}
}

func ColorMarshaler(m *LogMessage) ([]byte, error) {
	msg, err := TextMarshaler(m)
	if err != nil {
		return nil, err
	}
	if !colorEnable {
		return msg, nil
	}
	var tpl = "%s"
	switch m.Level {
	case Lerror:
		tpl = "\x1b[31;1m%s\x1b[0m"
	case Lwarn:
		tpl = "\x1b[33;1m%s\x1b[0m"
	case Linfo:
		tpl = "\x1b[36;1m%s\x1b[0m"
	case Ldebug:
	}
	return []byte(fmt.Sprintf(tpl, string(msg))), nil
}
