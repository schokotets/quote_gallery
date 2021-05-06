package database

import (
	"log"
	"sort"
	"time"
)

/* -------------------------------------------------------------------------- */
/*                                  CONSTANTS                                 */
/* -------------------------------------------------------------------------- */

const timing = time.Second * 3

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// quoteSliceByPop mirrors the cache.quoteSlice
// quotes are sorted by quote popularity (QuoteT.Stats.Pop)
var quoteIndexByPop []uint32

// quoteSliceByCon mirrors the cache.quoteSlice
// quotes are sorted by quote controversy (QuoteT.Stats.Con)
var quoteIndexByCon []uint32

// quoteSliceByTime mirrors the cache.quoteSlice
// quotes are sorted by quote unixtime (QuoteT.Unixtime)
var quoteIndexByTime []uint32

var cacheIndexingMux SimpleMutex = SimpleMutex{
	state: unlocked,
}

var isRequest bool = false
var t *time.Timer = nil


/* -------------------------------------------------------------------------- */
/*                      UNEXPORTED CACHE_MIRROR FUNCTIONS                     */
/* -------------------------------------------------------------------------- */

// Starts cache indexing, should only be called once
func startAutoCacheIndexing() {
	cacheIndexingMux.Lock()
	defer cacheIndexingMux.Unlock()

	if t == nil {
		t = time.AfterFunc(timing, handler)
	}
}

func stopAutoCacheIndexing() {
	cacheIndexingMux.Lock()
	defer cacheIndexingMux.Unlock()

	if t != nil {
		t.Stop()
		t = nil
	}
}

// unsafeGenerateCacheIndex is unsafe because it reads from cache without checking cache mutex
// this function can allways be called to enforce an immediate CacheIndex generation,
// whether AutoCacheIndexing is active or not
func unsafeForceCacheIndexGen() {
	cacheIndexingMux.Lock()
	defer cacheIndexingMux.Unlock()
	generator()
}

// requestCacheIndexGeneration is thread save
// The request will be processed with the next run of AutoCacheIndexing
func requestCacheIndexGen() error {
	isRequest = true
	return nil
}

/* -------------------------------------------------------------------------- */
/*                                   KEEPOUT                                  */
/* -------------------------------------------------------------------------- */

// Never call this function manually, the cache may get messed up
func generator() {
	diff := (len(cache.quoteSlice) - len(quoteIndexByTime))
	if diff > 0 {
		quoteIndexByTime = append(quoteIndexByTime, make([]uint32, diff)...)
		quoteIndexByPop  = append(quoteIndexByPop,  make([]uint32, diff)...)
		quoteIndexByCon  = append(quoteIndexByCon,  make([]uint32, diff)...)
	}

	quoteIndexByTime = quoteIndexByTime[0:len(cache.quoteSlice)]
	quoteIndexByPop  = quoteIndexByPop [0:len(cache.quoteSlice)]
	quoteIndexByCon  = quoteIndexByCon [0:len(cache.quoteSlice)]

	for i := range cache.quoteSlice {
		quoteIndexByTime[i] = uint32(i)
	}

	sort.Slice(quoteIndexByTime, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByTime[i]].Unixtime >
			cache.quoteSlice[quoteIndexByTime[j]].Unixtime
	})

	copy(quoteIndexByPop, quoteIndexByTime)
	copy(quoteIndexByCon, quoteIndexByTime)

	// By creating quoteIndexByPop and quoteIndexByCon from quoteIndexByTime and
	// using SliceStable, the quotes which cannot be compared by Popularity or Controversy
	// (i.e. have equal scores) are kept in chronological order
	sort.SliceStable(quoteIndexByPop, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByPop[i]].Stats.Pop >
			cache.quoteSlice[quoteIndexByPop[j]].Stats.Pop
	})

	sort.SliceStable(quoteIndexByCon, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByCon[i]].Stats.Con >
			cache.quoteSlice[quoteIndexByCon[j]].Stats.Con
	})

	log.Print("Index Generator:")
	log.Print(quoteIndexByTime)
	log.Print(quoteIndexByPop)
	log.Print(quoteIndexByCon)
}

// Never call this funtion manually, this may lead to a deadlock
func handler() {
	if isRequest {
		// execute generator
		// restart timer, if AutoCacheIndexing is still active

		isRequest = false

		globalMutex.MinorLock()
		defer globalMutex.MinorUnlock()

		cacheIndexingMux.Lock()
		defer cacheIndexingMux.Unlock()

		generator()

		if t != nil {
			t = time.AfterFunc(timing, handler)
		}
	} else {
		// Just restart timer, if AutoCacheIndexing is still active

		cacheIndexingMux.Lock()
		defer cacheIndexingMux.Unlock()

		if t != nil {
			t = time.AfterFunc(timing, handler)
		}
	}
}


