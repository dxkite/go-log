package log

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

type Group string

const Version1 = 1
const TimeBinary = /*version*/ 1 + /*sec*/ 8 + /*nsec*/ 4 + /*zone offset*/ 2

type LogMessage struct {
	Level   LogLevel  `json:"level"`
	Group   Group     `json:"group"`
	Time    time.Time `json:"time"`
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Message string    `json:"message"`
}

type LogMessageWriter interface {
	WriteLogMessage(msg *LogMessage) error
}

type LogLevel int

const LogDefaultMute = 3

const (
	Lerror LogLevel = iota
	Lwarn
	Linfo
	Ldebug
	LMaxLevel
)

var levelMap = map[LogLevel]string{
	Lerror: "ERROR",
	Lwarn:  "WARN",
	Linfo:  "INFO",
	Ldebug: "DEBUG",
}

func (l LogLevel) String() string {
	if v, ok := levelMap[l]; ok {
		return v
	}
	return fmt.Sprintf("Level-%2d", int(l))
}

type Logger struct {
	mu     sync.Mutex
	out    io.Writer
	caller bool
	level  LogLevel
	async  bool
	mute   int

	lastMsg *LogMessage
	lastCnt int
}

func New(w io.Writer, caller bool) *Logger {
	return &Logger{
		mu:     sync.Mutex{},
		out:    w,
		caller: caller,
		level:  LMaxLevel,
		mute:   LogDefaultMute,
	}
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
}

func (l *Logger) SetMute(m int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.mute = m
}

func (l *Logger) SetLevel(lv LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = lv
}

func (l *Logger) Writer() (w io.Writer) {
	return l.out
}

func (l *Logger) SetLogCaller(b bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.caller = b
}

func (l *Logger) SetAsync(b bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.async = b
}

func (l *Logger) split(args []interface{}) (Group, LogLevel, []interface{}) {
	var app Group
	level := Linfo
	for i := 0; i < 2; i++ {
		if len(args) > 1 {
			if vv, ok := args[0].(LogLevel); ok {
				level = vv
				args = args[1:]
			}
			if vv, ok := args[0].(Group); ok {
				app = vv
				args = args[1:]
			}
		}
	}
	return app, level, args
}

func (l *Logger) levelOutput(level LogLevel, args []interface{}) error {
	var app Group
	if len(args) > 1 {
		if vv, ok := args[0].(Group); ok {
			app = vv
			args = args[1:]
		}
	}
	return l.Output(3, app, level, fmt.Sprintln(args...))
}

func (l *Logger) Error(args ...interface{}) {
	w(l.levelOutput(Lerror, args))
}
func (l *Logger) Warn(args ...interface{}) {
	w(l.levelOutput(Lwarn, args))
}

func (l *Logger) Info(args ...interface{}) {
	w(l.levelOutput(Linfo, args))
}

func (l *Logger) Debug(args ...interface{}) {
	w(l.levelOutput(Ldebug, args))
}

func (l *Logger) Println(args ...interface{}) {
	var group Group
	var level LogLevel
	group, level, args = l.split(args)
	w(l.Output(2, group, level, fmt.Sprintln(args...)))
}

func (l *Logger) Fatalln(args ...interface{}) {
	w(l.levelOutput(Lerror, args))
	os.Exit(1)
}

func (l *Logger) Output(calldepth int, group Group, level LogLevel, s string) error {
	// 日志过滤
	if level > l.level {
		return nil
	}
	now := time.Now()
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.caller {
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	msg := &LogMessage{
		Level:   level,
		Group:   group,
		Time:    now,
		File:    file,
		Line:    line,
		Message: s,
	}

	if sameMessage(l.lastMsg, msg) {
		l.lastCnt++
	} else {
		l.lastCnt = 1
		l.lastMsg = msg
	}

	if l.mute > 0 && l.lastCnt >= l.mute {
		return nil
	}

	write := func() error {
		if vv, ok := l.out.(LogMessageWriter); ok {
			return vv.WriteLogMessage(msg)
		}
		_, err := l.out.Write(msg.marshal())
		return err
	}

	if l.async {
		go func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			_ = write()
		}()
		return nil
	}

	return write()
}

func (m *LogMessage) marshal() []byte {
	buf := []byte{Version1, byte(m.Level), byte(len(m.Group))}

	tm, _ := m.Time.MarshalBinary()
	buf = append(buf, tm...)
	bt := make([]byte, 4)

	binary.BigEndian.PutUint16(bt[:2], uint16(len(m.File)))
	buf = append(buf, bt[:2]...)

	binary.BigEndian.PutUint16(bt[:2], uint16(m.Line))
	buf = append(buf, bt[:2]...)

	binary.BigEndian.PutUint32(bt[:], uint32(len(m.Message)))
	buf = append(buf, bt[:]...)

	buf = append(buf, m.Group...)
	buf = append(buf, m.File...)
	buf = append(buf, m.Message...)
	return buf
}

func (m *LogMessage) unmarshal(r io.Reader) error {
	header := 3 + TimeBinary + 2 + 2 + 4
	off := 1
	buf := make([]byte, header)
	if _, er := io.ReadFull(r, buf); er != nil {
		return er
	}

	m.Level = LogLevel(buf[off])
	off += 1

	al := int(buf[off])
	off += 1

	m.Time = time.Time{}
	if er := m.Time.UnmarshalBinary(buf[off : off+TimeBinary]); er != nil {
		return er
	}
	off += TimeBinary

	fl := binary.BigEndian.Uint16(buf[off : off+2])
	off += 2

	line := binary.BigEndian.Uint16(buf[off : off+2])
	off += 2

	ml := binary.BigEndian.Uint32(buf[off : off+4])
	off += 4

	m.Line = int(line)
	txt := make([]byte, uint32(al)+uint32(fl)+ml)
	if _, er := io.ReadFull(r, txt); er != nil {
		return er
	}
	if al > 0 {
		m.Group = Group(txt[:al])
	}
	if fl > 0 {
		m.File = string(txt[al : uint16(al)+fl])
	}
	if ml > 0 {
		m.Message = string(txt[uint16(al)+fl:])
	}
	return nil
}

var std = New(NewTextWriter(os.Stdout), true)

func Error(args ...interface{}) {
	w(std.levelOutput(Lerror, args))
}

func Warn(args ...interface{}) {
	w(std.levelOutput(Lwarn, args))
}

func Info(args ...interface{}) {
	w(std.levelOutput(Linfo, args))
}

func Debug(args ...interface{}) {
	w(std.levelOutput(Ldebug, args))
}

func Fatalln(args ...interface{}) {
	w(std.levelOutput(Lerror, args))
	os.Exit(1)
}

func Println(args ...interface{}) {
	var app Group
	var level LogLevel
	app, level, args = std.split(args)
	_ = std.Output(2, app, level, fmt.Sprintln(args...))
}

func Output(calldepth int, app Group, level LogLevel, s string) error {
	return std.Output(calldepth+1, app, level, s)
}

func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

func SetLogCaller(b bool) {
	std.SetLogCaller(b)
}

func SetLevel(lv LogLevel) {
	std.SetLevel(lv)
}

func SetAsync(b bool) {
	std.SetAsync(b)
}

func Writer() (w io.Writer) {
	return std.Writer()
}

func w(err error) {
	if err != nil {
		fmt.Println("log.w error:", err)
	}
}

func sameMessage(a, b *LogMessage) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return false
	}
	if a.File != b.File {
		return false
	}
	if a.Line != b.Line {
		return false
	}
	if a.Message != b.Message {
		return false
	}
	return true
}
