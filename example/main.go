package main

import (
    "fmt"
    "time"

    "github.com/goupdate/deadlock"
)

func main() {
    var mutex deadlock.RWMutex

    deadlock.SetGlobalLockTimeout(time.Second/2, func(dur time.Duration, file string, line int) {
	panic(fmt.Sprintf("Detected deadlock via lock timeout! Locked for %s at %s:%d\n", dur, file, line))
    })

    go func() {
        mutex.Lock()
        defer mutex.Unlock()
        time.Sleep(time.Second) // Simulate long operation
    }()
	
    time.Sleep(time.Second/10)

    mutex.Lock() // << alert happens here
}
