package log

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

const TimeFormat = "2006-01-02 15:04:05.000"

var colorEnable = true

func TextMarshaler(m *LogMessage) ([]byte, error) {
	var msg string
	msg += m.Time.Format(TimeFormat) + " "

	if len(string(m.Group)) > 0 {
		msg += "[" + string(m.Group) + "] "
	}

	if len(m.File) > 0 {
		f := filepath.Base(m.File)
		msg += fmt.Sprintf("%s:%d ", f, m.Line)
	}

	msg += m.Message
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
