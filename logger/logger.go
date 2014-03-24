package logger

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	defaultLevel        = infoLevel
	defaultSectionLevel = errorLevel
)

const (
	quietLevel   = -1
	errorLevel   = 0
	infoLevel    = 1
	warningLevel = 2
	debugLevel   = 3
)

var (
	sections = map[string]*int{}
	std      = New(os.Stderr, "default", 0)
)

func init() {
	// Tell the flag package we are using -v but discard the value. We'll manually parse later.
	sections["default"] = flag.Int("v", defaultLevel, "verbose level")
}

type Logger struct {
	sync.Mutex
	out     io.Writer
	section string
	level   int
}

func New(out io.Writer, section string, level ...int) *Logger {
	if len(level) > 1 {
		panic("Can't instanciate a logger with more than one level")
	}
	lvl := 0
	if len(level) == 1 {
		lvl = level[0]
	}
	if section == "" {
		section = "default"
	}
	if out == nil {
		out = os.Stderr
	}
	l := &Logger{
		out:     out,
		section: section,
		level:   lvl,
	}
	if section != "default" {
		if _, exists := sections[section]; !exists && flag.Parsed() {
			panic("Can't add a section after parsing the flags")
		} else if !exists {
			sections[section] = flag.Int("v."+section, defaultSectionLevel, "verbose level for section"+section)
		}
	}
	return l
}

func (l *Logger) Output(level int, s string) error {
	sectionLevel := sections[l.section]
	if *sectionLevel >= level {
		l.Lock()
		_, err := l.out.Write([]byte(s))
		l.Unlock()
		return err
	}
	return nil
}

func (l *Logger) SetOutput(out io.Writer) {
	l.Lock()
	defer l.Unlock()

	l.out = out
}

func (l *Logger) Info(v ...interface{}) {
	l.Output(l.level+infoLevel, fmt.Sprint(v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Output(l.level+errorLevel, fmt.Sprint(v...))
}

func (l *Logger) Warning(v ...interface{}) {
	l.Output(l.level+warningLevel, fmt.Sprint(v...))
}

func (l *Logger) Debug(v ...interface{}) {
	l.Output(l.level+debugLevel, fmt.Sprint(v...))
}

func (l *Logger) Print(v ...interface{}) {
	l.Output(l.level, fmt.Sprint(v...))
}

func (l *Logger) Lprint(lvl int, v ...interface{}) {
	l.Output(l.level+lvl, fmt.Sprint(v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(l.level+infoLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(l.level+errorLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Output(l.level+warningLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(l.level+debugLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Output(l.level, fmt.Sprintf(format, v...))
}

func (l *Logger) Lprintf(lvl int, format string, v ...interface{}) {
	l.Output(l.level+lvl, fmt.Sprintf(format, v...))
}

func SetOutput(out io.Writer) {
	std.Lock()
	defer std.Unlock()

	std.out = out
}

func Info(v ...interface{}) {
	std.Output(std.level+infoLevel, fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	std.Output(std.level+errorLevel, fmt.Sprint(v...))
}

func Warning(v ...interface{}) {
	std.Output(std.level+warningLevel, fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	std.Output(std.level+debugLevel, fmt.Sprint(v...))
}

func Print(v ...interface{}) {
	std.Output(std.level, fmt.Sprint(v...))
}

func Lprint(lvl int, v ...interface{}) {
	std.Output(std.level+lvl, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(std.level+infoLevel, fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	std.Output(std.level+errorLevel, fmt.Sprintf(format, v...))
}

func Warningf(format string, v ...interface{}) {
	std.Output(std.level+warningLevel, fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	std.Output(std.level+debugLevel, fmt.Sprintf(format, v...))
}

func Printf(format string, v ...interface{}) {
	std.Output(std.level, fmt.Sprintf(format, v...))
}

func Lprintf(lvl int, format string, v ...interface{}) {
	std.Output(std.level+lvl, fmt.Sprintf(format, v...))
}
