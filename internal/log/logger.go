package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	err   *log.Logger
	out   *log.Logger
	file  *os.File
}

func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	fileWriter := io.Writer(f)
	stdWriter := io.Writer(os.Stderr)
	multi := io.MultiWriter(fileWriter, stdWriter)

	return &Logger{
		info: log.New(fileWriter, "I ", log.Ldate|log.Ltime),
		err:  log.New(multi, "E ", log.Ldate|log.Ltime),
		out:  log.New(os.Stderr, "", 0),
		file: f,
	}, nil
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

func (l *Logger) Out(format string, v ...interface{}) {
	l.out.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.err.Printf(format, v...)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.info.Printf("D "+format, v...)
}
