package deadlock

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/petermattis/goid"
)

type lockInfo struct {
	sync.Mutex

	goid int64 // if locked or locked, stores goroutine's id (goid)
	file string
	line int

	lockTime time.Time
}

type RWMutex struct {
	sync.RWMutex

	wlockInfo sync.Map // who already Locked this?
	rlockInfo sync.Map // who already RLocked this? allow double Rlocks

	lockTimeout        time.Duration                                  // = *2 at every alert
	lockTimeoutHandler func(dur time.Duration, file string, line int) // custom handler
}

var (
	lockTimeout = (time.Second * 2)
	// calls if any mutex locked for at 2,4,8,16 seconds, etc.
	lockTimeoutHandler = func(dur time.Duration, file string, line int) {
		fmt.Fprintf(os.Stderr, "RWMutex in %s: %d is locked more %s\n", file, line, dur.Truncate(time.Second).String())
	}

	globalTimers sync.Map

	mu sync.Mutex
)

func ResetGlobalTimers() {
	reseted := []interface{}{}
	globalTimers.Range(func(k, v interface{}) bool {
		t := k.(*time.Timer)
		if !t.Stop() {
			<-t.C
		}
		reseted = append(reseted, t)
		return true
	})
	for _, v := range reseted {
		globalTimers.Delete(v)
	}
}

var callerDeep = 2

// Helper function to get the caller's file and line number
func getCaller() (string, int) {
	_, file, line, _ := runtime.Caller(callerDeep) // Adjusted to get the correct caller
	return file, line
}

func monitor(timer *time.Timer, info *lockInfo, m *RWMutex) {
	for range timer.C {
		spent := time.Since(info.lockTime)

		timeout := lockTimeout // global
		if m.lockTimeout > 0 {
			timeout = m.lockTimeout // custom
		}

		if spent > timeout {
			mu.Lock()
			lockTimeoutHandlerV := lockTimeoutHandler
			mu.Unlock()
			if m.lockTimeoutHandler != nil {
				lockTimeoutHandlerV = m.lockTimeoutHandler
			}

			// who locked?
			file, line, _ := m.LastLocker()

			lockTimeoutHandlerV(spent, file, line)

			timer.Stop()
		}
	}
}

// file and line of last Lock() position in code (if locked!)
func (m *RWMutex) LastLocker() (string, int, time.Duration) {
	var locker *lockInfo
	m.rlockInfo.Range(func(key, value interface{}) bool {
		info := value.(*lockInfo)
		locker = info
		return false
	})
	if locker != nil {
		return locker.file, locker.line, time.Since(locker.lockTime)
	}
	m.wlockInfo.Range(func(key, value interface{}) bool {
		info := value.(*lockInfo)
		locker = info
		return false
	})
	if locker != nil {
		return locker.file, locker.line, time.Since(locker.lockTime)
	}
	return "", 0, 0
}

// Lock method with deadlock detection and timeout handling
func (m *RWMutex) Lock() {
	goid := goid.Get()
	file, line := getCaller()

	// forbid Lock() after Lock() in same goroutine
	if wlock, ok := m.wlockInfo.Load(goid); ok && wlock != nil {
		info := wlock.(*lockInfo)
		if info.goid == goid {
			panic(fmt.Sprintf("Double lock detected: goroutine %d trying to acquire lock it already holds at %s:%d", goid, info.file, info.line))
		}
	}
	// forbid Lock() after RLock() in same goroutine
	if rlock, ok := m.rlockInfo.Load(goid); ok && rlock != nil {
		info := rlock.(*lockInfo)
		if info.goid == goid {
			panic(fmt.Sprintf("Double read lock detected: goroutine %d trying to acquire read lock it already holds at %s:%d", goid, info.file, info.line))
		}
	}

	info := &lockInfo{goid: goid, file: file, line: line, lockTime: time.Now()}
	timeout := m.lockTimeout
	if timeout == 0 {
		timeout = lockTimeout
	}
	var timer *time.Timer
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		go monitor(timer, info, m)
		globalTimers.Store(timer, 1)
	}

	m.RWMutex.Lock()
	m.wlockInfo.Store(goid, info)

	if timer != nil {
		if !timer.Stop() {
			<-timer.C
		}
		globalTimers.Delete(timer)
	}
}

// Unlock method
func (m *RWMutex) Unlock() {
	goid := goid.Get()

	m.RWMutex.Unlock()
	m.wlockInfo.Delete(goid)
}

// RLock method with deadlock detection and timeout handling
func (m *RWMutex) RLock() {
	goid := goid.Get()
	file, line := getCaller()

	// forbid RLock() after Lock() in same goroutine
	if wlock, ok := m.wlockInfo.Load(goid); ok && wlock != nil {
		info := wlock.(*lockInfo)
		if info.goid == goid {
			panic(fmt.Sprintf("Double lock detected: goroutine %d trying to acquire lock it already holds at %s:%d", goid, info.file, info.line))
		}
	}

	info := &lockInfo{goid: goid, file: file, line: line, lockTime: time.Now()}
	timeout := m.lockTimeout
	if timeout == 0 {
		timeout = lockTimeout
	}
	var timer *time.Timer
	if timeout > 0 {
		timer = time.NewTimer(timeout)
		go monitor(timer, info, m)
		globalTimers.Store(timer, 1)
	}

	m.RWMutex.RLock()
	m.rlockInfo.Store(goid, info)

	if timer != nil {
		if !timer.Stop() {
			<-timer.C
		}
		globalTimers.Delete(timer)
	}
}

// RUnlock method
func (m *RWMutex) RUnlock() {
	goid := goid.Get()

	m.RWMutex.RUnlock()
	m.rlockInfo.Delete(goid)
}

// SetLockTimeout sets the lock timeout and handler for an instance of RWMutex
// if handler == nil or duration == 0, checking is turned off
func (m *RWMutex) SetLockTimeout(duration time.Duration, handler func(dur time.Duration, file string, line int)) {
	m.lockTimeout = duration
	m.lockTimeoutHandler = handler
}

// SetGlobalLockTimeout sets the global lock timeout and handler
// if handler == nil or duration == 0, checking is turned off
func SetGlobalLockTimeout(duration time.Duration, handler func(dur time.Duration, file string, line int)) {
	mu.Lock()
	lockTimeout = duration
	lockTimeoutHandler = handler
	mu.Unlock()
}
