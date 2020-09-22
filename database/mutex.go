package database

import (
	"runtime"
	"sync/atomic"
)

const (
	unlocked uint32 = 0
	locked   uint32 = 1
)

// Mutex struct
type Mutex struct {
	state               uint32
	readingThreadsCount uint32
	isWrite             bool
}

// Setup will initialize the Mutex struct to default values
func (m *Mutex) Setup() {
	m.state = unlocked
	m.readingThreadsCount = 0
	m.isWrite = false
}

// LockRead must be used if the corresponding rountine only reads from memory
func (m *Mutex) LockRead() {
	doNotRead := true
	for doNotRead {
		for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
			runtime.Gosched()
		}

		if m.isWrite == false {
			doNotRead = false
			m.readingThreadsCount++
		}
		atomic.StoreUint32(&m.state, unlocked)

		if doNotRead == true {
			runtime.Gosched()
		}
	}
}

// LockWrite must be used if the corresponding rountine also writes to memory
func (m *Mutex) LockWrite() {
	doNotWrite := true
	noOtherWrites := false

	for doNotWrite {
		for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
			runtime.Gosched()
		}

		if m.isWrite == false {
			m.isWrite = true
			noOtherWrites = true
		}

		if m.readingThreadsCount == 0 && noOtherWrites {
			doNotWrite = false
		}

		atomic.StoreUint32(&m.state, unlocked)

		if doNotWrite == true {
			runtime.Gosched()
		}
	}
}

// UnlockRead must be called to reverse LockRead
func (m *Mutex) UnlockRead() {
	for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
		runtime.Gosched()
	}

	if m.readingThreadsCount != 0 {
		m.readingThreadsCount--
	}

	atomic.StoreUint32(&m.state, unlocked)
}

// UnlockWrite must be called to reverse LockWrite
func (m *Mutex) UnlockWrite() {
	for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
	}

	m.isWrite = false

	atomic.StoreUint32(&m.state, unlocked)
}
