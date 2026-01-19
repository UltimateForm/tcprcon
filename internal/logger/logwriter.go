package logger

import (
	"io"
	"log"
	"os"

	"github.com/UltimateForm/tcprcon/internal/ansi"
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

func newWithCustomDestinations(level uint8, writer io.Writer) *LogWriter {
	var info, debug, errorWriter, critical, warn io.Writer = io.Discard, io.Discard, io.Discard, io.Discard, io.Discard
	if level >= LevelCritical {
		critical = levelWriter{level: LevelCritical, dst: writer}
	}
	if level >= LevelError {
		errorWriter = levelWriter{level: LevelError, dst: writer}
	}
	if level >= LevelWarning {
		warn = levelWriter{level: LevelWarning, dst: writer}
	}
	if level >= LevelInfo {
		info = levelWriter{level: LevelInfo, dst: writer}
	}
	if level >= LevelDebug {
		debug = levelWriter{level: LevelDebug, dst: writer}
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
	writer   *LogWriter
	Info     *log.Logger
	Debug    *log.Logger
	Err      *log.Logger
	Warn     *log.Logger
	Critical *log.Logger
)

func setGlobalLoggers() {
	Info = log.New(writer.Info, ansi.Format("TCPRCON:INF::", ansi.DefaultColor), 0)
	Debug = log.New(writer.Debug, ansi.Format("TCPRCON:DBG::", ansi.Yellow), 0)
	Err = log.New(writer.Error, ansi.Format("TCPRCON:ERR::", ansi.Red), 0)
	Warn = log.New(writer.Warn, ansi.Format("TCPRCON:WRN::", ansi.Magenta), 0)
	Critical = log.New(writer.Critical, ansi.Format("TCPRCON:CRT::", ansi.Red), 0)
}

func SetupCustomDestination(level uint8, customWriter io.Writer) {
	writer = newWithCustomDestinations(level, customWriter)
	setGlobalLoggers()
}

func Setup(level uint8) {
	writer = New(level)
	setGlobalLoggers()
}

func init() {
	Setup(LevelWarning)
}
