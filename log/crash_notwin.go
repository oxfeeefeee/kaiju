// +build !windows

package log

import (
    "os"
    "syscall"
    )

func setCrashLogFile(f *os.File) {
    syscall.Dup2(int(f.Fd()), 2)
}