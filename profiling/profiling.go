package profiling

import _ "net/http/pprof"
import (
    "log"
    "net/http"
    )

func RunProfiler() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
}