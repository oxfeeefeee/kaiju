package kaiju

import (
    "io"
    "log"
    "os"
    "fmt"
    "strings"
    "path/filepath"
    )

var logger *Logger

type Logger struct {
    *log.Logger
}

func InitLog() error {
    cfg := GetConfig()
    path := filepath.Join(GetConfigFileDir(), cfg.LogFileName)
    f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("error opening file: %v", err)
    }
    multi := io.MultiWriter(f, os.Stdout)
    l := log.New(multi, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
    logger = &Logger{l}
    return nil
}

// Add a debug function for log.Logger
func (l *Logger) Debugf(format string, v ...interface{}) {
    f := strings.Join([]string{"<DEBUG>",format}, "")
    l.Output(2, fmt.Sprintf(f, v...))
}

// Handy function
func MainLogger() *Logger {
    return logger
}