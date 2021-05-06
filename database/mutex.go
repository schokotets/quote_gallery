package database

import (
	"runtime"
	"sync/atomic"
)

/* -------------------------------------------------------------------------- */
/*                                  CONSTANT                                  */
/* -------------------------------------------------------------------------- */

const (
	unlocked uint32 = 0
	locked   uint32 = 1
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

// Mutex struct
type Mutex struct {
	state             uint32
	minorThreadsCount uint32
	isMajor           bool
}

// SimpleMutex struct
type SimpleMutex struct {
	state uint32
}

/* -------------------------------------------------------------------------- */
/*                               MUTEX FUNCTIONS                              */
/* -------------------------------------------------------------------------- */

// Setup will initialize the Mutex struct to default values
func (m *Mutex) Setup() {
	m.state = unlocked
	m.minorThreadsCount = 0
	m.isMajor = false
}

// MinorLock only blocks if MajorLock is active or imminent
// several MinorLocks can exist in parallel
// e.g. if a routine only wants to read
func (m *Mutex) MinorLock() {
	doBlock := true
	for doBlock {
		for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
			runtime.Gosched()
		}

		if !m.isMajor {
			doBlock = false
			m.minorThreadsCount++
		}
		atomic.StoreUint32(&m.state, unlocked)

		if doBlock {
			runtime.Gosched()
		}
	}
}

// MinorUnlock must be called to reverse MinorLock
func (m *Mutex) MinorUnlock() {
	for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
		runtime.Gosched()
	}

	if m.minorThreadsCount != 0 {
		m.minorThreadsCount--
	}

	atomic.StoreUint32(&m.state, unlocked)
}

// MajorLock blocks until there are no active MinorLocks and no active MajorLock
// if MajorLock is active all other Minor- and MajorLocks will block
// e.g. if a rountine wants to read and to write
func (m *Mutex) MajorLock() {
	doBlock := true
	// are there no other major locks active?
	noOtherMajors := false

	for doBlock {
		for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
			runtime.Gosched()
		}

		if !m.isMajor {
			m.isMajor = true
			noOtherMajors = true
		}

		if m.minorThreadsCount == 0 && noOtherMajors {
			doBlock = false
		}

		atomic.StoreUint32(&m.state, unlocked)

		if doBlock {
			runtime.Gosched()
		}
	}
}

// MajorUnlock must be called to reverse MajorLock
func (m *Mutex) MajorUnlock() {
	for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
	}

	m.isMajor = false

	atomic.StoreUint32(&m.state, unlocked)
}

/* -------------------------------------------------------------------------- */
/*                            SIMPLEMUTEX FUNCTIONS                           */
/* -------------------------------------------------------------------------- */

func (m *SimpleMutex) Lock() {
	for !atomic.CompareAndSwapUint32(&m.state, unlocked, locked) {
		runtime.Gosched()
	}
}

func (m *SimpleMutex) Unlock() {
	atomic.CompareAndSwapUint32(&m.state, locked, unlocked)
}
