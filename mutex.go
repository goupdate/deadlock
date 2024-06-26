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

	waitingMutexes sync.Map // *RWMutex -> *lockInfo

	mu sync.Mutex
)

func init() {
	go monitor()
}

func ResetGlobalTimers() {
	reset := []interface{}{}
	waitingMutexes.Range(func(k, _ interface{}) bool {
		reset = append(reset, k)
		return true
	})
	for _, v := range reset {
		waitingMutexes.Delete(v)
	}
}

var callerDeep = 2

// Helper function to get the caller's file and line number
func getCaller() (string, int) {
	_, file, line, _ := runtime.Caller(callerDeep) // Adjusted to get the correct caller
	return file, line
}

func monitor() {
	for {
		mu.Lock()
		timeoutV := lockTimeout // global
		mu.Unlock()

		if timeoutV > time.Millisecond*50 {
			time.Sleep(timeoutV / 3)
		} else {
			time.Sleep(time.Millisecond * 300)
		}

		alerted := []interface{}{}
		now := time.Now()
		waitingMutexes.Range(func(k, v interface{}) bool {
			m := k.(*RWMutex)
			info := v.(*lockInfo)

			spent := now.Sub(info.lockTime)

			timeout := timeoutV
			if m.lockTimeout > 0 {
				timeout = m.lockTimeout // custom
			}

			if spent > timeout && timeout > 0 {
				mu.Lock()
				lockTimeoutHandlerV := lockTimeoutHandler
				mu.Unlock()
				if m.lockTimeoutHandler != nil {
					lockTimeoutHandlerV = m.lockTimeoutHandler
				}

				if lockTimeoutHandlerV != nil {
					// who locked?
					file, line, _ := m.LastLocker()

					lockTimeoutHandlerV(spent, file, line)
				}
				alerted = append(alerted, k)
			}
			return true
		})
		for _, k := range alerted {
			waitingMutexes.Delete(k)
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
	waitingMutexes.Store(m, info)

	m.RWMutex.Lock()
	m.wlockInfo.Store(goid, info)

	waitingMutexes.Delete(m)
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
	waitingMutexes.Store(m, info)

	m.RWMutex.RLock()
	m.rlockInfo.Store(goid, info)

	waitingMutexes.Delete(m)
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

func (m *RWMutex) GetLockTimeout() time.Duration {
	return m.lockTimeout
}

// SetGlobalLockTimeout sets the global lock timeout and handler
// if handler == nil or duration == 0, checking is turned off
func SetGlobalLockTimeout(duration time.Duration, handler func(dur time.Duration, file string, line int)) {
	mu.Lock()
	lockTimeout = duration
	lockTimeoutHandler = handler
	mu.Unlock()
}

func GetGlobalLockTimeout() time.Duration {
	mu.Lock()
	defer mu.Unlock()
	return lockTimeout
}
