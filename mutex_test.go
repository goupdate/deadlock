package deadlock

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLockUnlock(t *testing.T) {
	var mu RWMutex

	goid := GetGoroutineId()

	mu.Lock(goid)
	mu.Unlock(goid)

	mu.RLock(goid)
	mu.RUnlock(goid)
}

func TestDeadlockDetection(t *testing.T) {
	var mu1, mu2, mu3 RWMutex
	pan := make(chan interface{})

	SetGlobalLockTimeout(time.Second, func(dur time.Duration, file string, line int) {
		go func() {
			str := fmt.Sprintf("rwmutex hangs: %s at %s:%d", dur.String(), file, line)
			fmt.Printf("detected: %s\n", str)
			pan <- str
		}()
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		mu1.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu2.Lock(goid)
		mu2.Unlock(goid)
		mu1.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu3.Lock(goid)
		mu3.Unlock(goid)
		mu2.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu1.Lock(goid) // << panics here
		mu1.Unlock(goid)
		mu3.Unlock(goid)
	}()

	time.Sleep(time.Second * 2)

	select {
	case p := <-pan:
		fmt.Printf("Detected deadlock: %v\n", p)
	default:
		t.Fatalf("expected a deadlock panic, but got none")
	}
}

func TestDeadlockDetection2(t *testing.T) {
	var mu1, mu2, mu3 RWMutex

	pan := make(chan interface{})
	done := false

	SetGlobalLockTimeout(time.Second*5, func(dur time.Duration, file string, line int) {
		if !done {
			t.Fatal("should not be triggered")
		}
	})

	mu1.SetLockTimeout(time.Millisecond*400, func(dur time.Duration, file string, line int) {
		done = true
		go func() {
			str := fmt.Sprintf("rwmutex hangs: %s at %s:%d", dur.String(), file, line)
			fmt.Printf("detected: %s\n", str)
			pan <- str
		}()
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()
		goid := GetGoroutineId()

		mu1.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu2.Lock(goid)
		mu2.Unlock(goid)
		mu1.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu3.Lock(goid)
		mu3.Unlock(goid)
		mu2.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu1.Lock(goid) // << panics here
		mu1.Unlock(goid)
		mu3.Unlock(goid)
	}()

	time.Sleep(time.Second * 2)

	select {
	case p := <-pan:
		fmt.Printf("Detected deadlock: %v\n", p)
	default:
		t.Fatalf("expected a deadlock panic, but got none")
	}
}

func TestDeadlockDetection3(t *testing.T) {
	var mu1, mu2, mu3 RWMutex

	pan := make(chan interface{})

	SetGlobalLockTimeout(time.Second, func(dur time.Duration, file string, line int) {
		go func() {
			str := fmt.Sprintf("rwmutex hangs: %s at %s:%d", dur.String(), file, line)
			fmt.Printf("detected: %s\n", str)
			pan <- str
		}()
	})

	mu1.SetLockTimeout(time.Second*3, func(dur time.Duration, file string, line int) {
		t.Fatal("should not be detected")
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		mu1.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu2.Lock(goid)
		mu2.Unlock(goid)
		mu1.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu3.Lock(goid)
		mu3.Unlock(goid)
		mu2.Unlock(goid)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				pan <- r
			}
		}()

		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock(goid)
		time.Sleep(500 * time.Millisecond)
		mu1.Lock(goid) // << panics here
		mu1.Unlock(goid)
		mu3.Unlock(goid)
	}()

	time.Sleep(time.Second * 2)

	select {
	case p := <-pan:
		fmt.Printf("Detected deadlock: %v\n", p)
	default:
		t.Fatalf("expected a deadlock panic, but got none")
	}
}

func TestDoubleLock(t *testing.T) {
	var mu RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock(1)
	mu.Lock(1)
}

func TestRecursiveLocks(t *testing.T) {
	var mu RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock(1)
	mu.Lock(1)

	mu.Unlock(1)
	mu.Unlock(1)
}

func TestComboLocks1(t *testing.T) {
	var mu RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock(1)
	mu.RLock(1)

	mu.RUnlock(1)
	mu.Unlock(1)
}

func TestComboLocks2(t *testing.T) {
	var mu RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.RLock(1)
	mu.Lock(1)

	mu.Unlock(1)
	mu.RUnlock(1)
}

func TestConcurrentLocks(t *testing.T) {
	var mu RWMutex
	var wg sync.WaitGroup

	var counter int
	cntm := 0

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		cntm++
		go func() {
			goid := GetGoroutineId()

			defer wg.Done()
			mu.Lock(goid)
			counter++
			mu.Unlock(goid)
		}()
	}

	wg.Wait()

	if counter != cntm {
		t.Fatalf("expected counter to be 100, but got %d", counter)
	}
}

func TestConcurrentRLocks(t *testing.T) {
	var mu RWMutex
	var wg sync.WaitGroup
	var counter int32
	cntm := int32(0)

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		cntm++
		go func() {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("expected no deadlock panic, but got: %v", r)
				}
			}()

			goid := GetGoroutineId()

			mu.RLock(goid)
			atomic.AddInt32(&counter, 1)
			mu.RUnlock(goid)
		}()
	}

	wg.Wait()

	if counter != cntm {
		t.Fatalf("expected counter to be 100, but got %d", counter)
	}
}

func TestDeadlockDetectionOnRLock(t *testing.T) {
	var mu1, mu2 RWMutex
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		goid := GetGoroutineId()

		mu1.RLock(goid)
		time.Sleep(500 * time.Millisecond)
		mu2.RLock(goid)
		mu2.RUnlock(goid)
		mu1.RUnlock(goid)
	}()

	go func() {
		defer wg.Done()
		goid := GetGoroutineId()

		time.Sleep(250 * time.Millisecond)
		mu2.RLock(goid)
		time.Sleep(500 * time.Millisecond)
		mu1.RLock(goid)
		mu1.RUnlock(goid)
		mu2.RUnlock(goid)
	}()

	// Expect a panic due to deadlock
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Detected deadlock, not expected: %v", r)
		}
	}()

	wg.Wait()
}

func TestMixedLocks(t *testing.T) {
	var mu RWMutex
	var wg sync.WaitGroup
	var counter int
	cntm := 0

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		cntm++
		go func() {
			defer wg.Done()
			goid := GetGoroutineId()

			mu.Lock(goid)
			counter++
			mu.Unlock(goid)
		}()
	}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		cntm++
		go func() {
			defer wg.Done()
			goid := GetGoroutineId()

			mu.Lock(goid)
			counter++
			mu.Unlock(goid)
		}()
	}

	wg.Wait()

	if counter != cntm {
		t.Fatalf("expected counter to be 200, but got %d", counter)
	}
}
