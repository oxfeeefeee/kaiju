package kaiju

import (
    "os"
    "time"
    "runtime"
    "math/rand"
    "github.com/oxfeeefeee/kaiju/profiling"
)

func Init() error {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    profiling.RunProfiler()

    rand.Seed(time.Now().UTC().UnixNano())

    err := ReadJsonConfigFile()
    if err != nil {
        MainLogger().Printf("Failed to ready config file: %s", err.Error())
        return err;
    }
    err = InitLog()
    if err != nil {
        MainLogger().Printf("Failed to init logging: %s", err.Error())
        return err;
    }

    // Print working directory
    wd, wderr := os.Getwd()
    if wderr == nil {
        MainLogger().Printf("Working directory: %s", wd)
    } else {
        MainLogger().Printf("Failed to print working directory: %s", wderr.Error())
        return wderr
    }
    return nil
}