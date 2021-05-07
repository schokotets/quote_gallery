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
// Default order: descending (most popular first)
var quoteIndexByPop []uint32

// quoteSliceByCon mirrors the cache.quoteSlice
// quotes are sorted by quote controversy (QuoteT.Stats.Con)
// Default order: descending (most controversal first)
var quoteIndexByCon []uint32

// quoteSliceByTime mirrors the cache.quoteSlice
// quotes are sorted by quote unixtime (QuoteT.Unixtime)
// Default order: descending (latest first)
var quoteIndexByTime []uint32

var cacheIndexingMux Mutex = Mutex{unlocked, 0, false}

var refreshNecessary bool = false
var t *time.Timer = nil

type indexFunction func(quoteSlice []QuoteT, n, from int)

// IndexHandler contains an indexing function and its sorting name
type IndexHandler struct {
	Function indexFunction
	Name string
}

// DefaultIndexHandlerName is used by the frontend
var DefaultIndexHandlerName = "timeDesc"

// IndexHandlerOrder is used by the frontend
var IndexHandlerOrder = [6]string{"timeDesc", "timeAsce", "popDesc", "popAsce", "conDesc", "conAsce"}

// IndexHandlers maps indexing names to corresponding functions and sorting names
var IndexHandlers = map[string]IndexHandler {
	"timeDesc": { timeDesc, "Zeit (neueste zuerst)" },
	"timeAsce": { timeAsce, "Zeit (älteste zuerst)" },
	"popDesc":  { popDesc,  "Beliebteste (beste zuerst)" },
	"popAsce":  { popAsce,  "Beliebteste (schlechteste zuerst)" },
	"conDesc":  { conDesc,  "Kontroversität (kontroverseste zuerst)" },
	"conAsce":  { conAsce,  "Kontroversität (kontroverseste zuletzt)" },
}

/* -------------------------------------------------------------------------- */
/*                     UNEXPORTED CACHE_INDEXING FUNCTIONS                    */
/* -------------------------------------------------------------------------- */

// Starts cache indexing, should only be called once
func startAutoCacheIndexing() {
	cacheIndexingMux.MajorLock()
	defer cacheIndexingMux.MajorUnlock()

	if t == nil {
		t = time.AfterFunc(timing, handler)
	}
}

func stopAutoCacheIndexing() {
	cacheIndexingMux.MajorLock()
	defer cacheIndexingMux.MajorUnlock()

	if t != nil {
		t.Stop()
		t = nil
	}
}

// unsafeForceCacheIndexGen is unsafe because it reads from cache without checking the global cache mutex
// this function can always be called to enforce an immediate CacheIndex generation,
// whether AutoCacheIndexing is active or not
func unsafeForceCacheIndexGen() {
	cacheIndexingMux.MajorLock()
	defer cacheIndexingMux.MajorUnlock()
	generateIndexes()
}

// requestCacheIndexGen is thread save
// The request will be processed with the next run of AutoCacheIndexing
func requestCacheIndexGen() error {
	refreshNecessary = true
	return nil
}

// unsafeGetQuotesFromIndexedCache
// n         number of quotes to get
// from	     starting index
// indexFn   one of the indexFunctions
func unsafeGetQuotesFromIndexedCache(n, from int, indexFn indexFunction) ([]QuoteT) {
	if indexFn == nil {
		return nil
	}

	cacheIndexingMux.MinorLock()
	defer cacheIndexingMux.MinorUnlock()

	// lengths of quoteIndexByTime, quoteIndexByPop
	// and quoteIndexByCon ARE ALLWAYS EQUAL
	if from >= len(quoteIndexByTime) {
		return nil
	}
	if from+n >= len(quoteIndexByTime) {
		n = len(quoteIndexByTime) - from
	}
	quoteSlice := make([]QuoteT, n)

	indexFn(quoteSlice, n, from)

	return quoteSlice
}

/* -------------------------------------------------------------------------- */
/*                                   KEEPOUT                                  */
/* -------------------------------------------------------------------------- */

// Never call this function manually, the cache may get messed up
func generateIndexes() {
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
}

// Never call this funtion manually, this may lead to a deadlock
func handler() {
	if refreshNecessary {
		// execute generateIndexes
		// restart timer, if AutoCacheIndexing is still active

		refreshNecessary = false

		globalMutex.MinorLock()
		defer globalMutex.MinorUnlock()

		cacheIndexingMux.MajorLock()
		defer cacheIndexingMux.MajorUnlock()

		generateIndexes()

		if t != nil {
			t = time.AfterFunc(timing, handler)
		}
	} else {
		// Just restart timer, if AutoCacheIndexing is still active

		cacheIndexingMux.MajorLock()
		defer cacheIndexingMux.MajorUnlock()

		if t != nil {
			t = time.AfterFunc(timing, handler)
		}
	}
}

/* -------------------------------------------------------------------------- */
/*                             INDEXING FUNCTIONS                             */
/* -------------------------------------------------------------------------- */

// fills the given quoteSlice with n quotes starting with index from
// sorted by descending time
func timeDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByTime[from + i] ]
	}
}

// fills the given quoteSlice with n quotes starting with index from
// sorted by ascending time
func timeAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByTime[from - i] ]
	}
}

// fills the given quoteSlice with n quotes starting with index from
// sorted by descending popularity
func popDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByPop[from + i] ]
	}
}

// fills the given quoteSlice with n quotes starting with index from
// sorted by ascending popularity
func popAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByPop[from - i] ]
	}
}

// fills the given quoteSlice with n quotes starting with index from
// sorted by descending description
func conDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByCon[from + i] ]
	}
}

// fills the given quoteSlice with n quotes starting with index from
// sorted by ascending description
func conAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByCon[from - i] ]
	}
}
