package deadlock_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/goupdate/deadlock"
)

func TestLockUnlock(t *testing.T) {
	var mu deadlock.RWMutex

	mu.Lock()
	mu.Unlock()

	mu.RLock()
	mu.RUnlock()
}

func TestDeadlockDetection(t *testing.T) {
	var mu1, mu2, mu3 deadlock.RWMutex

	pan := make(chan interface{})

	deadlock.SetGlobalLockTimeout(time.Second, func(dur time.Duration, file string, line int) {
		go func() {
			str := fmt.Sprintf("rwmutex hangs: %s at %s:%d", dur.String(), file, line)
			fmt.Printf("detected: %s\n", str)
			pan <- str
		}()
	})

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		mu1.Lock()
		time.Sleep(500 * time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock()
		time.Sleep(500 * time.Millisecond)
		mu3.Lock()
		mu3.Unlock()
		mu2.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock()
		time.Sleep(500 * time.Millisecond)
		mu1.Lock() // << panics here
		mu1.Unlock()
		mu3.Unlock()
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
	var mu1, mu2, mu3 deadlock.RWMutex

	pan := make(chan interface{})
	done := false

	deadlock.SetGlobalLockTimeout(time.Second*2, func(dur time.Duration, file string, line int) {
		if !done {
			t.Fatal("should not be triggered")
		}
	})

	mu1.SetLockTimeout(time.Second, func(dur time.Duration, file string, line int) {
		done = true
		go func() {
			str := fmt.Sprintf("rwmutex hangs: %s at %s:%d", dur.String(), file, line)
			fmt.Printf("detected: %s\n", str)
			pan <- str
		}()
	})

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		mu1.Lock()
		time.Sleep(500 * time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock()
		time.Sleep(500 * time.Millisecond)
		mu3.Lock()
		mu3.Unlock()
		mu2.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock()
		time.Sleep(500 * time.Millisecond)
		mu1.Lock() // << panics here
		mu1.Unlock()
		mu3.Unlock()
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
	var mu1, mu2, mu3 deadlock.RWMutex

	pan := make(chan interface{})

	deadlock.SetGlobalLockTimeout(time.Second, func(dur time.Duration, file string, line int) {
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
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		mu1.Lock()
		time.Sleep(500 * time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu2.Lock()
		time.Sleep(500 * time.Millisecond)
		mu3.Lock()
		mu3.Unlock()
		mu2.Unlock()
	}()

	go func() {
		defer func() {
			if r := recover(); r == nil {
				pan <- r
			}
		}()

		time.Sleep(250 * time.Millisecond)
		mu3.Lock()
		time.Sleep(500 * time.Millisecond)
		mu1.Lock() // << panics here
		mu1.Unlock()
		mu3.Unlock()
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
	var mu deadlock.RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock()
	mu.Lock()
}

func TestRecursiveLocks(t *testing.T) {
	var mu deadlock.RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock()
	mu.Lock()

	mu.Unlock()
	mu.Unlock()
}

func TestComboLocks1(t *testing.T) {
	var mu deadlock.RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.Lock()
	mu.RLock()

	mu.RUnlock()
	mu.Unlock()
}

func TestComboLocks2(t *testing.T) {
	var mu deadlock.RWMutex

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected a deadlock panic, but got none")
		} else {
			fmt.Println("Detected system deadlock:", r)
		}
	}()

	mu.RLock()
	mu.Lock()

	mu.Unlock()
	mu.RUnlock()
}

func TestConcurrentLocks(t *testing.T) {
	var mu deadlock.RWMutex
	var wg sync.WaitGroup
	var counter int

	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if counter != 100 {
		t.Fatalf("expected counter to be 100, but got %d", counter)
	}
}

func TestConcurrentRLocks(t *testing.T) {
	var mu deadlock.RWMutex
	var wg sync.WaitGroup
	var counter int

	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if counter != 100 {
		t.Fatalf("expected counter to be 100, but got %d", counter)
	}
}

func TestDeadlockDetectionOnRLock(t *testing.T) {
	var mu1, mu2 deadlock.RWMutex
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		mu1.RLock()
		time.Sleep(500 * time.Millisecond)
		mu2.RLock()
		mu2.RUnlock()
		mu1.RUnlock()
	}()

	go func() {
		defer wg.Done()
		time.Sleep(250 * time.Millisecond)
		mu2.RLock()
		time.Sleep(500 * time.Millisecond)
		mu1.RLock()
		mu1.RUnlock()
		mu2.RUnlock()
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
	var mu deadlock.RWMutex
	var wg sync.WaitGroup
	var counter int

	wg.Add(200)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if counter != 200 {
		t.Fatalf("expected counter to be 200, but got %d", counter)
	}
}
