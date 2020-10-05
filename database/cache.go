package database

import (
	"errors"
	"log"
)

/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED CACHE FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// Fills cache from PostgreSQL database
// unsafe functions aren't concurrency safe
func unsafeLoadCache() error {
	var err error

	unsafeClearCache()

	log.Print("Filling cache from PostgreSQL database...")

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		return errors.New("unsafeLoadCache: pinging database failed" + err.Error())
	}

	/* --------------------------------- QUOTES --------------------------------- */

	// get all quotes from PostgreSQL database
	rows, err := postgresDatabase.Query(`SELECT 
		QuoteID,
		TeacherID, 
		Context,
		Text,
		Unixtime,
		Upvotes FROM quotes`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading quotes from database failed: " + err.Error())
	}

	// initialize wordsMap of cache
	cache.wordsMap = make(map[string]wordsMapT)

	// Iterrate over all quotes from PostgreSQL database
	for rows.Next() {
		// Get id and text of quote
		var q QuoteT
		rows.Scan(&q.QuoteID, &q.TeacherID, &q.Context, &q.Text, &q.Unixtime, &q.Upvotes)

		// add to local database
		// unsafe, because cache is already locked for writing
		err = unsafeAddQuoteToCache(q)
		if err != nil {
			return errors.New("unsafeLoadCache: adding quote to cache failed: " + err.Error())
		}
	}

	rows.Close()

	/* -------------------------------- TEACHERS -------------------------------- */

	// get all teachers from PostgreSQL database
	rows, err = postgresDatabase.Query(`SELECT
		TeacherID, 
		Name, 
		Title, 
		Note FROM teachers`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading teachers from database failed: " + err.Error())
	}

	// Iterate over all teachers from PostgreSQL database
	for rows.Next() {
		// Get teacher data (id, name, title, note)
		var t TeacherT
		rows.Scan(&t.TeacherID, &t.Name, &t.Title, &t.Note)

		// add to local database
		// unsafe, because cache is already locked for writing
		unsafeAddTeacherToCache(t)
	}

	rows.Close()

	log.Print("Filled cache successfully")
	return nil
}

func unsafeClearCache() {
	cache.quoteSlice = nil
	cache.teacherSlice = nil
	cache.wordsMap = nil
}

// Just adds quote to cache (quoteSlice and wordsMap) without checking q.QuoteID
// using addQuoteToCache without checking if q.QuoteID already exists may be fatal
func addQuoteToCache(q QuoteT) error {

	cache.mux.MajorLock()
	defer cache.mux.MajorUnlock()

	return unsafeAddQuoteToCache(q)
}

func unsafeAddQuoteToCache(q QuoteT) error {

	cache.quoteSlice = append(cache.quoteSlice, q)
	var enumID int32 = int32(len(cache.quoteSlice) - 1)

	if enumID < 0 {
		return errors.New("unsafeAddQuoteToCache: could not add quote to quoteSlice of cache")
	}

	// Iterrate over all words of quote
	for word, count := range getWordsFromString(q.Text) {
		wordsMapItem := cache.wordsMap[word]
		wordsMapItem.totalOccurences += count

		wordsMapItem.occurenceSlice = append(wordsMapItem.occurenceSlice, occurenceSliceT{enumID, count})

		cache.wordsMap[word] = wordsMapItem
	}

	return nil
}

// unsafe functions aren't concurrency safe
func addTeacherToCache(t TeacherT) {
	cache.mux.MajorLock()
	defer cache.mux.MajorUnlock()
	unsafeAddTeacherToCache(t)
}

// unsafe functions aren't concurrency safe
func unsafeAddTeacherToCache(t TeacherT) {
	cache.teacherSlice = append(cache.teacherSlice, t)
}

func overwriteTeacherInCache(t TeacherT) error {
	cache.mux.MajorLock()
	defer cache.mux.MajorUnlock()

	affected := false
	for i, v := range cache.teacherSlice {
		if v.TeacherID == t.TeacherID {
			cache.teacherSlice[i] = t
			affected = true
			break
		}
	}

	if affected == false {
		return errors.New("overwriteTeacherInCache: could not find specified entry for overwrite")
	}
	return nil
}

func overwriteQuoteInCache(q QuoteT) error {
	cache.mux.MajorLock()
	defer cache.mux.MajorUnlock()

	var enumID int32 = -1
	for i, v := range cache.quoteSlice {
		if v.QuoteID == q.QuoteID {
			cache.quoteSlice[i] = q
			enumID = int32(i)
		}
	}

	if enumID < 0 {
		return errors.New("overwriteQuoteInCache: could not find specified entry for overwrite")
	}

	wordsFromString := getWordsFromString(q.Text)

	for word, wordsMapItem := range cache.wordsMap {
		for i, v := range wordsMapItem.occurenceSlice {
			if v.enumID == enumID {
				wordCount := wordsFromString[word]
				wordsMapItem.totalOccurences -= wordsMapItem.occurenceSlice[i].count
				if wordCount > 0 {
					wordsMapItem.occurenceSlice[i].count = wordCount
					wordsMapItem.totalOccurences += wordCount
					wordsFromString[word] = 0
					cache.wordsMap[word] = wordsMapItem
				} else if wordsMapItem.totalOccurences == 0 {
					delete(cache.wordsMap, word)
				} else {
					iMax := len(wordsMapItem.occurenceSlice) - 1
					wordsMapItem.occurenceSlice[i] = wordsMapItem.occurenceSlice[iMax]
					wordsMapItem.occurenceSlice[iMax] = occurenceSliceT{0, 0}
					wordsMapItem.occurenceSlice = wordsMapItem.occurenceSlice[:iMax]
					cache.wordsMap[word] = wordsMapItem
				}
				break
			}
		}
	}

	for word, count := range wordsFromString {
		if count > 0 {
			wordsMapItem := cache.wordsMap[word]
			wordsMapItem.totalOccurences += count

			wordsMapItem.occurenceSlice = append(wordsMapItem.occurenceSlice, occurenceSliceT{enumID, count})

			cache.wordsMap[word] = wordsMapItem
		}
	}

	return nil
}

func getQuotesFromCache() *[]QuoteT {
	cache.mux.MinorLock()
	defer cache.mux.MinorUnlock()
	quoteSlice := cache.quoteSlice
	return &quoteSlice
}

func getTeachersFromCache() *[]TeacherT {
	cache.mux.MinorLock()
	defer cache.mux.MinorUnlock()
	teacherSlice := cache.teacherSlice

	return &teacherSlice
}

func getQuotesByStringFromCache(text string) *[]QuoteT {
	cache.mux.MinorLock()
	defer cache.mux.MinorUnlock()

	quoteSlice := cache.quoteSlice

	for word, count := range getWordsFromString(text) {
		wordsMapItem := cache.wordsMap[word]

		_ = count
		for _, v := range wordsMapItem.occurenceSlice {
			quoteSlice[v.enumID].Match += float32(v.count) / float32(wordsMapItem.totalOccurences)
		}
	}
	return &quoteSlice
}

// PrintWordsMap is a debugging function
func PrintWordsMap() {
	log.Print(cache.wordsMap)
}
