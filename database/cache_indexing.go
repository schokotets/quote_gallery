package database

import (
	"sort"
	"time"
)

/* -------------------------------------------------------------------------- */
/*                                  CONSTANTS                                 */
/* -------------------------------------------------------------------------- */

const timing = time.Minute * 3

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
var isAutoCacheIndexing bool = false
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
func unsafeGenerateCacheIndex() {
	cacheIndexingMux.Lock()
	defer cacheIndexingMux.Unlock()

	// Create new index slice if none exists
	if quoteIndexByPop == nil {
		quoteIndexByPop = make([]uint32, len(cache.quoteSlice))
	}

	if quoteIndexByCon == nil {
		quoteIndexByCon = make([]uint32, len(cache.quoteSlice))
	}

	if quoteIndexByTime == nil {
		quoteIndexByTime = make([]uint32, len(cache.quoteSlice))
	}

	// Modify existing slice to fit new quote amount
	for i := range cache.quoteSlice {
		if i >= len(quoteIndexByPop) {
			// the index slices will allways have equal lengths
			quoteIndexByPop = append(quoteIndexByPop, uint32(i))
			quoteIndexByCon = append(quoteIndexByCon, uint32(i))
			quoteIndexByTime = append(quoteIndexByTime, uint32(i))
		} else {
			quoteIndexByPop[i] = uint32(i)
			quoteIndexByCon[i] = uint32(i)
			quoteIndexByTime[i] = uint32(i)
		}
	}
	quoteIndexByPop = quoteIndexByPop[0:len(cache.quoteSlice)]
	quoteIndexByCon = quoteIndexByCon[0:len(cache.quoteSlice)]
	quoteIndexByTime = quoteIndexByTime[0:len(cache.quoteSlice)]

	sort.Slice(quoteIndexByPop, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByPop[i]].Stats.Pop >
			   cache.quoteSlice[quoteIndexByPop[j]].Stats.Pop
	})

	sort.Slice(quoteIndexByCon, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByCon[i]].Stats.Con >
			   cache.quoteSlice[quoteIndexByCon[j]].Stats.Con
	})

	sort.Slice(quoteIndexByTime, func(i, j int) bool {
		return cache.quoteSlice[quoteIndexByTime[i]].Unixtime >
			   cache.quoteSlice[quoteIndexByTime[j]].Unixtime
	})

	if t != nil {
		t = time.AfterFunc(timing, handler)
	}
}


/* -------------------------------------------------------------------------- */
/*                                   KEEPOUT                                  */
/* -------------------------------------------------------------------------- */

// Never call this funtion manually, this may lead to a deadlock
func handler() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
	unsafeGenerateCacheIndex()
}


