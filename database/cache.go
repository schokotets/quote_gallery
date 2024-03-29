package database

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

// wordsMapT stores all the necessary search information for one word
// totalOccurences  number of occurences of this word in all quotes
// occurenceSlice   stores the number of occurences of this word for every quote
type wordsMapT struct {
	totalOccurences int32
	occurenceSlice  []occurenceSliceT
}

// occurenceSliceT stores the number of occurences of one word for one quote
// enumID  cache internal index of the quote
// count   number of occurences
type occurenceSliceT struct {
	enumID int32
	count  int32
}

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// Created from database at (re)start
// cache is a cache of the database to speed up read operations
//
// unverified quotes will not be cached in the local database, because read operations
// will only be performed by the operator and thus be very rare
//
// important: the index of a quote in quoteSlice is called its enumID
// which is used to quickly identify a quote with the wordsMap
//
// voteSlice doesn't require the UserID-field of VoteT, because the UserID is already
// used as index of voteSlice. But for the sake of not defining a second vote-struct,
// the one defined in database.go is used
var cache struct {
	quoteSlice   []QuoteT
	teacherSlice []TeacherT
	wordsMap     map[string]wordsMapT
	userSlice    []UserT
	voteSlice    [][]VoteT
}

/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED CACHE FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// Fills cache from database
// unsafe functions aren't concurrency safe
func unsafeLoadCache() error {
	var err error

	unsafeClearCache()

	log.Print("Filling cache from database...")

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		return errors.New("unsafeLoadCache: pinging database failed" + err.Error())
	}

	/* --------------------------------- QUOTES --------------------------------- */

	// get all quotes from database
	rows, err := database.Query(`SELECT
		QuoteID,
		TeacherID,
		Context,
		Text,
		Unixtime FROM quotes`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading quotes from database failed: " + err.Error())
	}

	// initialize wordsMap of cache
	cache.wordsMap = make(map[string]wordsMapT)

	// Iterrate over all quotes from database
	for rows.Next() {
		// Get id and text of quote
		var q QuoteT
		err = rows.Scan(&q.QuoteID, &q.TeacherID, &q.Context, &q.Text, &q.Unixtime)
		if err != nil {
			return errors.New("unsafeLoadCache: parsing quotes failed: " + err.Error())
		}

		// add to local database
		// unsafe, because cache is already locked for writing
		err = unsafeAddQuoteToCache(q)
		if err != nil {
			return errors.New("unsafeLoadCache: adding quote to cache failed: " + err.Error())
		}
	}

	rows.Close()

	/* -------------------------------- TEACHERS -------------------------------- */

	// get all teachers from database
	rows, err = database.Query(`SELECT
		TeacherID,
		Name,
		Title,
		Note FROM teachers`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading teachers from database failed: " + err.Error())
	}

	// Iterate over all teachers from database
	for rows.Next() {
		// Get teacher data (id, name, title, note)
		var t TeacherT
		err = rows.Scan(&t.TeacherID, &t.Name, &t.Title, &t.Note)
		if err != nil {
			return errors.New("unsafeLoadCache: parsing teachers failed: " + err.Error())
		}

		// add to local database
		// unsafe, because cache is already locked for writing
		unsafeAddTeacherToCache(t)
	}

	rows.Close()

	/* ---------------------------------- USERS --------------------------------- */

	// get all users from database
	rows, err = database.Query(`SELECT
		UserID,
		Name,
		Password,
		Admin FROM users`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading users from database failed: " + err.Error())
	}

	// Iterrate over all users from database
	for rows.Next() {
		// Get user data (id, name, password admin)
		var u UserT
		err = rows.Scan(&u.UserID, &u.Name, &u.Password, &u.Admin)
		if err != nil {
			return errors.New("unsafeLoadCache: parsing users failed: " + err.Error())
		}

		// add to local database
		// unsafe, because cache is already locked for writing
		unsafeAddUserToCache(u)
		if err != nil {
			return errors.New("unsafeLoadCache: adding user to cache failed: " + err.Error())
		}
	}

	rows.Close()

	/* ---------------------------------- VOTES --------------------------------- */

	// get all votes from database
	rows, err = database.Query(`SELECT
		UserID,
		QuoteID,
		Rating FROM votes`)

	if err != nil {
		return errors.New("unsafeLoadCache: loading votes from database failed: " + err.Error())
	}

	// Iterrate over all votes from database
	for rows.Next() {
		// Get vote data (userid, quoteid)
		var vote VoteT

		err = rows.Scan(&vote.UserID, &vote.QuoteID, &vote.Val)
		if err != nil {
			return errors.New("unsafeLoadCache: parsing votes failed: " + err.Error())
		}

		// add to local database
		// unsafe, because cache is already locked for writing
		_, err = unsafeAddVoteToCache(vote)
		if err != nil {
			return errors.New("unsafeLoadCache: adding vote to cache failed: " + err.Error())
		}
	}

	rows.Close()

	log.Print("Filled cache successfully")

	unsafeForceCacheIndexGen()
	startAutoCacheIndexing()

	return nil
}

func unsafeClearCache() {
	cache.quoteSlice = nil
	cache.teacherSlice = nil
	cache.wordsMap = nil
	cache.userSlice = nil
}

// Just adds quote to cache (quoteSlice and wordsMap) without checking q.QuoteID
// using addQuoteToCache without checking if q.QuoteID already exists may be fatal
func unsafeAddQuoteToCache(q QuoteT) error {

	q.Match = 0

	cache.quoteSlice = append(cache.quoteSlice, q)
	enumID := int32(len(cache.quoteSlice) - 1)

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
func unsafeAddTeacherToCache(t TeacherT) {
	cache.teacherSlice = append(cache.teacherSlice, t)
}

// unsafe functions aren't concurrency safe
func unsafeAddUserToCache(u UserT) {
	cache.userSlice = append(cache.userSlice, u)
}

// unsafe functions aren't concurrency safe
func unsafeAddVoteToCache(vote VoteT) (QuoteT, error) {
	if vote.UserID < 1 {
		// u must be greater than zero to be a valid UserID
		return QuoteT{}, errors.New("unsafeAddVoteToCache: invalid UserID, must be greater than zero")
	}

	if vote.Val < VoteMin || vote.Val > VoteMax {
		return QuoteT{}, fmt.Errorf("unsafeAddVoteToCache: invalid Rating, must be in range %d-%d", VoteMin, VoteMax)
	}

	for len(cache.voteSlice) < int(vote.UserID) {
		cache.voteSlice = append(cache.voteSlice, []VoteT{})
	}

	for i, oldVote := range cache.voteSlice[vote.UserID-1] {
		if oldVote.QuoteID == vote.QuoteID {
			// already voted - that's not a problem
			// old vote gets overwritten

			for j, quote := range cache.quoteSlice {
				if quote.QuoteID == vote.QuoteID {
					cache.quoteSlice[j].Stats.Data[oldVote.Val-1]--
					cache.quoteSlice[j].Stats.Data[vote.Val-1]++
					cache.voteSlice[vote.UserID-1][i] = vote

					calculateQuoteStats(&cache.quoteSlice[j])

					return cache.quoteSlice[j], nil
				}
			}

			// something went wrong, quote doesn't exist (anymore)
			return QuoteT{}, fmt.Errorf("unsafeAddVoteToCache: quote with QuoteID %d doesn't exist (anymore)", vote.QuoteID)

		}
	}

	for i, quote := range cache.quoteSlice {
		if quote.QuoteID == vote.QuoteID {
			cache.quoteSlice[i].Stats.Data[vote.Val-1]++
			cache.voteSlice[vote.UserID-1] = append(cache.voteSlice[vote.UserID-1], vote)

			calculateQuoteStats(&cache.quoteSlice[i])

			return cache.quoteSlice[i], nil
		}
	}

	// something went wrong, quote doesn't exist (anymore)
	return QuoteT{}, fmt.Errorf("unsafeAddVoteToCache: quote with QuoteID %d doesn't exist (anymore)", vote.QuoteID)
}

func unsafeOverwriteTeacherInCache(t TeacherT) error {

	affected := false
	for i, v := range cache.teacherSlice {
		if v.TeacherID == t.TeacherID {
			cache.teacherSlice[i] = t
			affected = true
			break
		}
	}

	if affected == false {
		return errors.New("unsafeOverwriteTeacherInCache: could not find specified entry for overwrite")
	}
	return nil
}

// any fields beside QuoteID, TeacherID, Context, Text will be ignored
func unsafeOverwriteQuoteInCache(q QuoteT) error {

	var enumID int32 = -1
	for i, v := range cache.quoteSlice {
		if v.QuoteID == q.QuoteID {
			cache.quoteSlice[i].Text = q.Text
			cache.quoteSlice[i].Context = q.Context
			cache.quoteSlice[i].TeacherID = q.TeacherID
			enumID = int32(i)
		}
	}

	if enumID < 0 {
		return errors.New("unsafeOverwriteQuoteInCache: could not find specified entry for overwrite")
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

func unsafeDeleteTeacherFromCache(ID int32) error {
	quotes := unsafeGetAllQuotesFromCache()
	for _, q := range quotes {
		if q.TeacherID == ID {
			log.Print(q)
			err := unsafeDeleteQuoteFromCache(q.QuoteID)
			if err != nil {
				return errors.New("unsafeDeleteTeacherFromCache: could not delete quote from cache: " + err.Error())
			}
		} else {
			log.Print(q.TeacherID)
		}
	}

	notDeleted := true
	for i, t := range cache.teacherSlice {
		if t.TeacherID == ID {
			iMax := int32(len(cache.teacherSlice) - 1)
			cache.teacherSlice[i] = cache.teacherSlice[iMax]
			cache.teacherSlice[iMax] = TeacherT{}
			cache.teacherSlice = cache.teacherSlice[:iMax]
			notDeleted = false
			break
		}
	}

	if notDeleted {
		return errors.New("unsafeDeleteTeacherFromCache: could not find entry to delete")
	}

	return nil

}

func unsafeDeleteQuoteFromCache(ID int32) error {
	var enumIDRemove int32 = -1
	var enumIDReplace int32 = -1

	for i, v := range cache.quoteSlice {
		if v.QuoteID == ID {
			enumIDReplace = int32(len(cache.quoteSlice) - 1)
			cache.quoteSlice[i] = cache.quoteSlice[enumIDReplace]
			cache.quoteSlice[enumIDReplace] = QuoteT{}
			cache.quoteSlice = cache.quoteSlice[:enumIDReplace]
			enumIDRemove = int32(i)
		}
	}

	if enumIDRemove < 0 {
		return errors.New("unsafeDeleteQuoteFromCache: could not find specified entry to delete")
	}

	for word, wordsMapItem := range cache.wordsMap {
		iMax := len(wordsMapItem.occurenceSlice) - 1
		if iMax < 0 {
			// occurenceSlice is empty, i.e. there are no occurences of word
			// this should not happen, but if it does, word can be removed from wordsMapItem
			delete(cache.wordsMap, word)
			continue
		}

		// Iterate through occurenceSlice
		indexRemove := -1
		indexReplace := -1
		for i, v := range wordsMapItem.occurenceSlice {
			if v.enumID == enumIDRemove {
				// found entry, which should be removed
				indexRemove = i
			}

			if v.enumID == enumIDReplace {
				// found entry, whose enumID needs to be replaced
				indexReplace = i
			}
		}

		if indexReplace >= 0 {
			wordsMapItem.occurenceSlice[indexReplace].enumID = enumIDRemove
		}

		if indexRemove >= 0 {
			wordsMapItem.totalOccurences -= wordsMapItem.occurenceSlice[indexRemove].count
			wordsMapItem.occurenceSlice[indexRemove] = wordsMapItem.occurenceSlice[iMax]
			wordsMapItem.occurenceSlice[iMax] = occurenceSliceT{5, 5}
			wordsMapItem.occurenceSlice = wordsMapItem.occurenceSlice[:iMax]
		}

		cache.wordsMap[word] = wordsMapItem

		if len(wordsMapItem.occurenceSlice) == 0 {
			delete(cache.wordsMap, word)
		}
	}

	return nil
}

// Returns maximum amount of n quotes from cache starting from index from.
// Returns nil if starting index is too big.
func unsafeGetNQuotesFromFromCache(n, from int) []QuoteT {
	if from >= len(cache.quoteSlice) {
		return nil
	}
	if from+n >= len(cache.quoteSlice) {
		n = len(cache.quoteSlice) - from
	}
	quoteSlice := make([]QuoteT, n)
	copy(quoteSlice, cache.quoteSlice[from:from+n])
	return quoteSlice
}

func unsafeGetAllQuotesFromCache() []QuoteT {
	quoteSlice := make([]QuoteT, len(cache.quoteSlice))
	copy(quoteSlice, cache.quoteSlice)
	return quoteSlice
}

func unsafeGetQuotesAmountFromCache() int {
	return len(cache.quoteSlice)
}

func unsafeGetTeachersFromCache() []TeacherT {
	teacherSlice := make([]TeacherT, len(cache.teacherSlice))
	copy(teacherSlice, cache.teacherSlice)
	return teacherSlice
}

func unsafeGetTeacherByIDFromCache(ID int32) (TeacherT, bool) {
	for _, teacher := range cache.teacherSlice {
		if teacher.TeacherID == ID {
			return teacher, true
		}
	}

	// TeacherID = 0 indicates no matching teacher has been found
	return TeacherT{}, false
}

func unsafeGetQuotesByStringFromCache(text string) []QuoteT {

	quoteSlice := make([]QuoteT, len(cache.quoteSlice))
	copy(quoteSlice, cache.quoteSlice)

	for word, count := range getWordsFromString(text) {
		wordsMapItem := cache.wordsMap[word]

		_ = count
		for _, v := range wordsMapItem.occurenceSlice {
			quoteSlice[v.enumID].Match += float32(v.count) / float32(wordsMapItem.totalOccurences)
		}
	}
	return quoteSlice
}

func unsafeGetUserFromCache(name string, password string) UserT {
	for _, user := range cache.userSlice {
		if strings.EqualFold(name, user.Name) && password == user.Password {
			return user
		}
	}

	// UserID = 0 indicates no matching user has been found
	return UserT{}
}

func unsafeAddUserDataToQuote(q *QuoteT, userid int32) error {
	if userid < 1 || len(cache.voteSlice) < int(userid) {
		// u must be greater than zero to be a valid UserID
		// u is used as index in cache.voteSlice, hence cannot be greater than the length of the slice
		return errors.New("unsafeAddUserDataToQuote: invalid UserID")
	}

	for _,vote := range cache.voteSlice[userid-1] {
		if vote.QuoteID == q.QuoteID {
			q.MyVote = vote.Val
			return nil
		}
	}

	q.MyVote = VoteNone

	return nil
}

/* -------------------------------------------------------------------------- */
/*                              HELPER FUNCTIONS                              */
/* -------------------------------------------------------------------------- */

func calculateQuoteStats(quote *QuoteT) {
	// Favourite
	num := int32(0)
	sum := int32(0)
	for rating, amount := range quote.Stats.Data {
		num += amount
		sum += amount * (int32(rating+1) - VoteDefault)
	}

	if num == 0 {
		quote.Stats.Pop = 0
		quote.Stats.Con = 0
		quote.Stats.Num = 0
		return
	}

	quote.Stats.Pop = float32(sum) / float32( len(cache.userSlice) ) + VoteDefault
	quote.Stats.Num = num

	// Controversial ... ?
	mean := float32(sum) / float32(num) + VoteDefault
	div := float32(0)
	for rating, amount := range quote.Stats.Data {
		div += float32(amount) * ( ( float32(rating+1) - mean ) * ( float32(rating+1) - mean ) )
	}

	quote.Stats.Con = div / float32(num)
}

/* -------------------------------------------------------------------------- */
/*                                  DEBUGGING                                 */
/* -------------------------------------------------------------------------- */

// PrintWordsMap is a debugging function
func PrintWordsMap() {
	log.Print(cache.wordsMap)
}

// PrintUserSlice is a debugging function
func PrintUserSlice() {
	log.Print(cache.userSlice)
}

// PrintVoteSlice is a debugging function
func PrintVoteSlice() {
	log.Print(cache.voteSlice)
}

// PrintQuoteSlice is a debugging function
func PrintQuoteSlice() {
	log.Print(cache.quoteSlice)
}
