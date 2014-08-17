package kaiju

import (
    "os"
    "log"
    "time"
    "runtime"
    "math/rand"
    "path/filepath"
    klog "github.com/oxfeeefeee/kaiju/log"
)

func init() {
    err := readConfig()
    if err != nil {
        log.Panicln("Failed to read config file: ", err)
    }

    path := filepath.Join(ConfigFileDir(), cfg.LogFileName)
    klog.Init(path)

    runtime.GOMAXPROCS(runtime.NumCPU())
    rand.Seed(time.Now().UTC().UnixNano())

    // Print working directory
    wd, err := os.Getwd()
    if err == nil {
        klog.Infoln("Working directory:", wd)
    } else {
        klog.Panicln("Failed to print working directory:", err)
    }
}