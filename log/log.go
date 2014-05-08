package log

import (
    "log"
    "os"
    "fmt"
    "strings"
    )


var mainLogger, KioLogger, KioMsgLogger, KioPeerLogger *Logger
var KDBLogger *Logger

func init() {
    mainLogger = createLogger("main")
    KioLogger = createLogger("kio")
    KioMsgLogger = createLogger("kio.msg")
    KioPeerLogger = createLogger("kio.peer")
    KDBLogger = createLogger("kdb")
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