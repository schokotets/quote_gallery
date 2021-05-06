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

var isRequest bool = false
var t *time.Timer = nil

// var indexFunctions = [6]func(quoteSlice []QuoteT, n, from int) {
// 	byTimeDesc,	// 0: latest first
// 	byTimeAsce, // 1: oldest first
// 	byPopDesc,  // 2: most popular first
// 	byPopAsce, 	// 3: least popular first
// 	byConDesc,	// 4: most controversial first
// 	byConAsce,	// 5: least controversial first
// }

var indexFunctions = map[string]func(quoteSlice []QuoteT, n, from int) {
	"TimeDesc":	TimeDesc,
	"TimeAsce":	TimeAsce,
	"PopDesc":  PopDesc,
	"PopAsce":  PopAsce,
	"ConDesc":  ConDesc,
	"ConAsce":  ConAsce,
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

// unsafeForceCacheIndexGen is unsafe because it reads from cache without checking cache mutex
// this function can allways be called to enforce an immediate CacheIndex generation,
// whether AutoCacheIndexing is active or not
func unsafeForceCacheIndexGen() {
	cacheIndexingMux.MajorLock()
	defer cacheIndexingMux.MajorUnlock()
	generator()
}

// requestCacheIndexGen is thread save
// The request will be processed with the next run of AutoCacheIndexing
func requestCacheIndexGen() error {
	isRequest = true
	return nil
}

// unsafeGetQuotesFromIndexedCache
// n		 number of quotes to get
// from		 starting index
// indexType type of indexing, refer to indexFuntions
func unsafeGetQuotesFromIndexedCache(n, from int, indexType string) ([]QuoteT) {
	indexFuntion, ok := indexFunctions[indexType]
	if !ok {
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

	indexFuntion(quoteSlice, n, from)

	return quoteSlice
}

func isIndexType(indexType string) bool {
	_, ok := indexFunctions[indexType]
	return ok
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
}

// Never call this funtion manually, this may lead to a deadlock
func handler() {
	if isRequest {
		// execute generator
		// restart timer, if AutoCacheIndexing is still active

		isRequest = false

		globalMutex.MinorLock()
		defer globalMutex.MinorUnlock()

		cacheIndexingMux.MajorLock()
		defer cacheIndexingMux.MajorUnlock()

		generator()

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

// indexType: 0
func TimeDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByTime[from + i] ]
	}
}

// indexType: 1
func TimeAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByTime[from - i] ]
	}
}

// indexType: 2
func PopDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByPop[from + i] ]
	}
}

// indexType: 3
func PopAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByPop[from - i] ]
	}
}

// indexType: 4
func ConDesc(quoteSlice []QuoteT, n, from int) {
	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByCon[from + i] ]
	}
}

// indexType: 5
func ConAsce(quoteSlice []QuoteT, n, from int) {
	// calculate real starting index, because of reversed order
	from = len(quoteIndexByTime) - from - 1

	for i := 0; i < n; i++ {
		quoteSlice[i] = cache.quoteSlice[ quoteIndexByCon[from - i] ]
	}
}
