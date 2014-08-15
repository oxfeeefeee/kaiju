package kaiju

import (
    "io"
    "log"
    "os"
    "fmt"
    "strings"
    //"syscall"
    "runtime"
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
    
    if runtime.GOOS == "windows" { /*
        kernel32 := syscall.MustLoadDLL("kernel32.dll")
        ph := kernel32.MustFindProc("SetStdHandle")
        err = winSetStdHandle(ph, syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
        if err != nil {
            log.Fatalf("Error setting up winSetStdHandle: %v", err)
        } */
    }
    return nil
}

// Add a debug function for log.Logger
func (l *Logger) Debugf(format string, v ...interface{}) {
    f := strings.Join([]string{"<DEBUG>",format}, "")
    l.Output(2, fmt.Sprintf(f, v...))
}
/*
func winSetStdHandle(ph *syscall.Proc, stdhandle int32, handle syscall.Handle) error {
    r0, _, e1 := syscall.Syscall(ph.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
    if r0 == 0 {
        if e1 != 0 {
            return error(e1)
        }
        return syscall.EINVAL
    }
    return nil
}*/

// Handy function
func MainLogger() *Logger {
    return logger
}

