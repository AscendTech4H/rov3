package main

import (
	"errors"
	"sync"
)

type xMutex struct {
	lck   sync.Mutex
	inuse bool
}

func (xm *xMutex) Lock() error {
	xm.lck.Lock()
	defer xm.lck.Unlock()
	if xm.inuse {
		return errors.New("resource currently in use")
	}
	xm.inuse = true
	return nil
}

func (xm *xMutex) Unlock() {
	xm.lck.Lock()
	defer xm.lck.Unlock()
	xm.inuse = false
}
