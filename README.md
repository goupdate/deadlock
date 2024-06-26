## Deadlock Detection RWMutex

This Go package provides a custom implementation of `RWMutex` with built-in deadlock detection and timeout handling. It aims to enhance the standard `sync.RWMutex` by preventing deadlocks and providing useful debugging information when a deadlock situation is detected.

### Features

- **Deadlock Detection**: Automatically detects and prevents double locking scenarios within the same goroutine, including both `Lock` and `RLock` operations.
- **Timeout Handling**: Configurable lock timeout with customizable timeout handlers to notify or take action when a lock exceeds the specified duration.
- **Detailed Debugging Information**: Captures and reports the file and line number where the lock was last held, providing valuable insights for debugging.
- **Global and Instance-Specific Configuration**: Supports both global and per-instance lock timeout settings and handlers.

### Usage

1. **Import the Package**:
    ```go
    import "github.com/goupdate/deadlock"
    ```

2. **Initialize RWMutex**:
    ```go
    var mutex deadlock.RWMutex
    ```

3. **Lock and Unlock**:
    ```go
    go func() {
        mutex.Lock()
        defer mutex.Unlock()
        // Critical section
    }()
    ```

4. **Read Lock and Unlock**:
    ```go
    go func() {
        mutex.RLock()
        defer mutex.RUnlock()
        // Read-only section
    }()
    ```

5. **Set Global Lock Timeout and Handler**:
    ```go
    deadlock.SetGlobalLockTimeout(time.Second*5, func(dur time.Duration, file string, line int) {
        fmt.Printf("Global lock timeout: %s at %s:%d\n", dur, file, line)
    })
    ```

6. **Set Instance-Specific Lock Timeout and Handler**:
    ```go
    mutex.SetLockTimeout(time.Second*2, func(dur time.Duration, file string, line int) {
        fmt.Printf("Instance lock timeout: %s at %s:%d\n", dur, file, line)
    })
    ```

### Example

Here's a simple example demonstrating the usage of `RWMutex` with deadlock detection and timeout handling:

```go
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
```