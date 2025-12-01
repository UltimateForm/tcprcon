package logger

import (
	"io"
	"log"
	"os"
)

const (
	LevelEmergency = iota // 0: System is unusable
	LevelAlert            // 1: Action must be taken immediately
	LevelCritical         // 2: Critical conditions
	LevelError            // 3: Error conditions
	LevelWarning          // 4: Warning conditions
	LevelNotice           // 5: Normal but significant condition
	LevelInfo             // 6: Informational messages
	LevelDebug            // 7: Debug-level messages
)

type levelWriter struct {
	level uint8
	dst   io.Writer
}

func (src levelWriter) Write(bytes []byte) (int, error) {
	// fmt.Fprintf(src.dst, "L%v::", src.level)
	return src.dst.Write(bytes)
}

type LogWriter struct {
	Info     io.Writer
	Debug    io.Writer
	Error    io.Writer
	Critical io.Writer
	Warn     io.Writer
}

// TODO: consider deleting this, will probably never need
func New(level uint8) *LogWriter {
	var info, debug, errorWriter, critical, warn io.Writer = io.Discard, io.Discard, io.Discard, io.Discard, io.Discard
	if level >= LevelCritical {
		critical = levelWriter{level: LevelCritical, dst: os.Stderr}
	}
	if level >= LevelError {
		errorWriter = levelWriter{level: LevelError, dst: os.Stderr}
	}
	if level >= LevelWarning {
		warn = levelWriter{level: LevelWarning, dst: os.Stdout}
	}
	if level >= LevelInfo {
		info = levelWriter{level: LevelInfo, dst: os.Stdout}
	}
	if level >= LevelDebug {
		debug = levelWriter{level: LevelDebug, dst: os.Stdout}
	}

	return &LogWriter{
		Info:     info,
		Error:    errorWriter,
		Critical: critical,
		Warn:     warn,
		Debug:    debug,
	}
}

var (
	Writer   *LogWriter
	Info     *log.Logger
	Debug    *log.Logger
	Err      *log.Logger
	Warn     *log.Logger
	Critical *log.Logger
)

func init() {
	Writer = New(LevelWarning)
	Info = log.New(Writer.Info, "INF::", 0)
	Debug = log.New(Writer.Debug, "DBG::", 0)
	Err = log.New(Writer.Error, "ERR::", 0)
	Warn = log.New(Writer.Warn, "WRN::", 0)
	Critical = log.New(Writer.Critical, "CRT::", 0)
}
