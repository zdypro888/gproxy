package main

import (
	"errors"
	"fmt"
	"sync"
)

type locker struct {
	mutex   sync.Mutex
	Count   int
	Locking bool
}

func (lck *locker) Lock(lock bool) error {
	lck.mutex.Lock()
	defer lck.mutex.Unlock()
	if lck.Locking {
		return errors.New("tunnel is locking")
	}
	if lock {
		if lck.Count > 0 {
			return fmt.Errorf("tunnel has client: %d", lck.Count)
		}
		lck.Locking = true
	}
	lck.Count++
	return nil
}

func (lck *locker) Unlock() {
	lck.mutex.Lock()
	defer lck.mutex.Unlock()
	lck.Locking = false
	lck.Count--
}
