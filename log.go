package kaiju

import (
    "log"
    "os"
    "fmt"
    "strings"
    )

var mainLogger *Logger

func init() {
    mainLogger = createLogger("main")
}

type Logger struct{
    *log.Logger
}

// Add a debug function for log.Logger
func (l *Logger) Debugf(format string, v ...interface{}) {
    f := strings.Join([]string{"<DEBUG>",format}, "")
    l.Logger.Output(2, fmt.Sprintf(f, v...))
}

func createLogger(category string) *Logger {
    l := log.New(os.Stdout, category, log.Ldate|log.Lmicroseconds|log.Lshortfile)
    return &Logger{l}
}

// Handy function
func MainLogger() *Logger {
    return mainLogger
}