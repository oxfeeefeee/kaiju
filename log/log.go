package log

import (
    "io"
    "log"
    "os"
    "fmt"
    )

var logger *Logger

type Logger struct {
    *log.Logger
}

func Init(path string) {
    f, err := os.OpenFile(path, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
        log.Panicf("error opening file: %s", err)
    }
    setCrashLogFile(f)
    multi := io.MultiWriter(f, os.Stdout)
    l := log.New(multi, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)
    logger = &Logger{l}
}

func Debug(v ...interface{}) {
    logger.SetPrefix("<DEBUG>")
    logger.Output(2, fmt.Sprint(v...))
}

func Debugf(format string, v ...interface{}) {
    logger.SetPrefix("<DEBUG>")
    logger.Output(2, fmt.Sprintf(format, v...))
}

func Debugln(v ...interface{}) {
    logger.SetPrefix("<DEBUG>")
    logger.Output(2, fmt.Sprintln(v...))
}

func Info(v ...interface{}) {
    logger.SetPrefix("<INFO>")
    logger.Output(2, fmt.Sprint(v...))
}

func Infof(format string, v ...interface{}) {
    logger.SetPrefix("<INFO>")
    logger.Output(2, fmt.Sprintf(format, v...))
}

func Infoln(v ...interface{}) {
    logger.SetPrefix("<INFO>")
    logger.Output(2, fmt.Sprintln(v...))
}

func Warning(v ...interface{}) {
    logger.SetPrefix("<WARNING>")
    logger.Output(2, fmt.Sprint(v...))
}

func Warningf(format string, v ...interface{}) {
    logger.SetPrefix("<WARNING>")
    logger.Output(2, fmt.Sprintf(format, v...))
}

func Warningln(v ...interface{}) {
    logger.SetPrefix("<WARNING>")
    logger.Output(2, fmt.Sprintln(v...))
}

func Error(v ...interface{}) {
    logger.SetPrefix("<ERROR>")
    logger.Output(2, fmt.Sprint(v...))
}

func Errorf(format string, v ...interface{}) {
    logger.SetPrefix("<ERROR>")
    logger.Output(2, fmt.Sprintf(format, v...))
}

func Errorln(v ...interface{}) {
    logger.SetPrefix("<ERROR>")
    logger.Output(2, fmt.Sprintln(v...))
}

func Panic(v ...interface{}) {
    logger.SetPrefix("<PANIC>")
    logger.Output(2, fmt.Sprint(v...))
}

func Panicf(format string, v ...interface{}) {
    logger.SetPrefix("<PANIC>")
    logger.Output(2, fmt.Sprintf(format, v...))
}

func Panicln(v ...interface{}) {
    logger.SetPrefix("<PANIC>")
    logger.Output(2, fmt.Sprintln(v...))
}