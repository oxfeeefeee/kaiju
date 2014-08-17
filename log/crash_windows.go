// +build windows

package log

import (
    "os"
    "syscall"
    )

func setStdHandle(ph *syscall.Proc, stdhandle int32, handle syscall.Handle) error {
    r0, _, e1 := syscall.Syscall(ph.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
    if r0 == 0 {
        if e1 != 0 {
            return error(e1)
        }
        return syscall.EINVAL
    }
    return nil
}

func setCrashLogFile(f *os.File) {
    kernel32 := syscall.MustLoadDLL("kernel32.dll")
    ph := kernel32.MustFindProc("SetStdHandle")
    err = winSetStdHandle(ph, syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
    if err != nil {
        Panicf("Error setting up winSetStdHandle: %s", err)
    }
}