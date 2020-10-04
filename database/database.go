package database

import (
	"database/sql"
	"errors"
	"log"

	// loading postgresql driver
	_ "github.com/lib/pq"
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

// QuoteT stores one quote
// QuoteID   the unique identificator of the quote
// TeacherID the unique identifier of the corresponding teacher
// Context	  the context of the quote
// text       the text of the quote itself
// unixtime   optional
//
// match  	  only locally, not safed in PostgreSQL database!
//			  used by GetQuotesFromString to quantify how well this quote fits the string
type QuoteT struct {
	QuoteID   uint32
	TeacherID uint32
	Context   string
	Text      string
	Unixtime  uint64
	Upvotes   uint32
	Match     float32
}

// UnverifiedQuoteT stores one unverified quote
// QuoteID    the unique identificator of the unverified quote
// teacher 	   the name of the teacher
// Context	   the context of the quote
// text        the text of the quote itself
// unixtime    optional
// iphash	   optional
//
type UnverifiedQuoteT struct {
	QuoteID  uint32
	Teacher  string
	Context  string
	Text     string
	Unixtime uint64
	IPHash   uint64
}

// TeacherT stores one teacher
// TeacherID the unique identifier of the teacher
// name       the teacher's name
// title      the teacher's title
// note       optional notes, e.g. subjects
type TeacherT struct {
	TeacherID uint32
	Name      string
	Title     string
	Note      string
}

type wordsMapT struct {
	totalOccurences uint32
	occurenceSlice  []occurenceSliceT
}

type occurenceSliceT struct {
	enumid int32
	count  uint32
}

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// Handle to the PostgreSQL database, used as long time storage
var postgresDatabase *sql.DB

// Created from PostgreSQL database at (re)start
// localDatabase is a cache of the postgreSQL database to speed up read operations
//
// unverified quotes will not be cached in the local database, because read operations
// will only be performed by the operator and thus be very rare
//
// important: the unverifiedQuote table will not be mirrored
//
// important: the index of a quote in quoteSlice is called its enumid
// which is used to quickly identify a quote with the wordsMap
var localDatabase struct {
	quoteSlice   []QuoteT
	teacherSlice []TeacherT
	wordsMap     map[string]wordsMapT
	mux          Mutex
}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED GENERAL FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// Setup initializes the database backend
// Initialize postgres database
// Create localDatabase from postgresDatabase
//
// Must be called only once at startup before any of the other database functions
func Setup() error {
	var err error

	// Open PostgreSQL database
	postgresDatabase, err = sql.Open(
		"postgres",
		`user=postgres 
		password=1234 
		dbname=quote_gallery 
		sslmode=disable`)
	if err != nil {
		return errors.New("Setup: connecting to database failed: " + err.Error())
	}

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("Setup: pinging database failed: " + err.Error())
	}

	// Create teachers table in PostgreSQL database if it doesn't exist
	// for more information see TeachersT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS teachers (
		TeacherID serial PRIMARY KEY, 
		Name varchar, 
		Title varchar, 
		Note varchar)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("Setup: creating teachers table failed: " + err.Error())
	}

	// Create quotes table in PostgreSQL database if it doesn't exist
	// for more information see QuoteT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS quotes (
		QuoteID serial PRIMARY KEY,
		TeacherID integer REFERENCES teachers (TeacherID), 
		Context varchar,
		Text varchar,
		Unixtime bigint,
		Upvotes integer)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("Setup: creating quotes table failed: " + err.Error())
	}

	// Create unverifiedQuotes table in PostgreSQL database if it doesn't exist
	// for more information see UnverifiedQuoteT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS unverifiedQuotes (
		QuoteID serial PRIMARY KEY,
		Teacher varchar, 
		Context varchar,
		Text varchar,
		Unictime bigint,
		IPHash bigint)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("Setup: creating unverifiedQuotes table failed: " + err.Error())
	}

	localDatabase.mux.Setup()
	loadLocalDatabase()

	return nil

}

// Close database backend
func Close() {
	localDatabase.mux.LockWrite()
	localDatabase.quoteSlice = nil
	localDatabase.teacherSlice = nil
	localDatabase.wordsMap = nil
	localDatabase.mux.UnlockWrite()
	postgresDatabase.Close()
}

/* -------------------------------------------------------------------------- */
/*                          EXPORTED QUOTES FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// GetQuotes returns a slice containing all quotes
// The weight variable will be zero
func GetQuotes() *[]QuoteT {
	localDatabase.mux.LockRead()
	defer localDatabase.mux.UnlockRead()
	quoteSlice := localDatabase.quoteSlice

	return &quoteSlice
}

// GetQuotesByString returns a slice containing all quotes
// The weight variable will indicate how well the given text matches the corresponding quote
func GetQuotesByString(text string) *[]QuoteT {

	localDatabase.mux.LockRead()
	defer localDatabase.mux.UnlockRead()

	quoteSlice := localDatabase.quoteSlice

	for word, count := range getWordsFromString(text) {
		wordsMapItem := localDatabase.wordsMap[word]

		_ = count
		for _, v := range wordsMapItem.occurenceSlice {
			quoteSlice[v.enumid].Match += float32(v.count) / float32(wordsMapItem.totalOccurences)
		}
	}
	return &quoteSlice
}

// StoreQuote stores a new quote
// If the ID is not zero, StoreQuote will try to find the appropriate quote and overwrite it
// If the ID is zero a new quote will be created
func StoreQuote(q QuoteT) error {
	var err error

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("StoreQuote: pinging database failed: " + err.Error())
	}

	if q.QuoteID == 0 {
		// add quote to postgresDatabase
		err = postgresDatabase.QueryRow(
			`INSERT INTO quotes (TeacherID, Context, Text, Unixtime, Upvotes) VALUES ($1, $2, $3, $4, $5) RETURNING QuoteID`,
			q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes).Scan(&q.QuoteID)
		if err != nil {
			return errors.New("StoreQuote: inserting quote into database failed: " + err.Error())
		}

		// add quote to localDatabase
		addQuoteToLocalDatabase(q)
	} else {
		// try to find corresponding entry postgresDatabase and overwrite it
		var res sql.Result
		res, err = postgresDatabase.Exec(
			`UPDATE quotes SET TeacherID=$2, Context=$3, Text=$4, Unixtime=$5, Upvotes=$6 WHERE QuoteID=$1`,
			q.QuoteID, q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes)
		if err != nil {
			return errors.New("StoreQuote: updating quote in database failed: " + err.Error())
		}
		if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
			return errors.New("StoreQuote: could not find specified database row for updating")
		}

		// try to find corresponding entry in localDatabase and overwrite it
		err = overwriteQuoteInLocalDatabase(q)
		if err != nil {
			return errors.New("StoreQuote: updating quote in local cache failed: " + err.Error())
		}
	}

	return nil
}

// DeleteQuote deletes the quote corresponding to the given ID from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote(ID int) {

}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED TEACHERS FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

// GetTeachers returns a slice containing all teachers
// The returned slice is not sorted
func GetTeachers() *[]TeacherT {
	localDatabase.mux.LockRead()
	defer localDatabase.mux.UnlockRead()
	teacherSlice := localDatabase.teacherSlice

	return &teacherSlice
}

// StoreTeacher stores a new teacher
// If the ID is not zero, StoreTeacher will try to find the corresponding teacher and overwrite it
// If the ID is nil a new teacher will be created
func StoreTeacher(t TeacherT) error {
	var err error

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("StoreTeacher: pinging database failed: " + err.Error())
	}

	if t.TeacherID == 0 {
		// add teacher to postgresDatabase
		err = postgresDatabase.QueryRow(
			`INSERT INTO teachers (Name, Title, Note) VALUES ($1, $2, $3) RETURNING TeacherID`,
			t.Name, t.Title, t.Note).Scan(&t.TeacherID)
		if err != nil {
			return errors.New("StoreTeacher: inserting teacher into database failed: " + err.Error())
		}

		// add teacher to localDatabase
		addTeacherToLocalDatabase(t)
	} else {
		// try to find corresponding entry postgresDatabase and overwrite it
		var res sql.Result
		res, err = postgresDatabase.Exec(
			`UPDATE teachers SET Name=$2, Title=$3, Note=$4 WHERE TeacherID=$1`,
			t.TeacherID, t.Name, t.Title, t.Note)
		if err != nil {
			return errors.New("StoreTeacher: updating teacher in database failed: " + err.Error())
		}
		if rowsAffected, _ := res.RowsAffected(); rowsAffected != 0 {
			return errors.New("StoreTeacher: could not find specified database row for updating")
		}

		// try to find corresponding entry in localDatabase and overwrite it
		err = overwriteTeacherInLocalDatabase(t)
		if err != nil {
			return errors.New("StoreTeacher: updating quote in local cache failed: " + err.Error())
		}
	}

	log.Print(localDatabase.teacherSlice)

	return nil
}

// DeleteTeacher deletes the teacher corresponding to the given ID from the database and the teachers slice
// It will delete all corresponding quotes
func DeleteTeacher() {

}

/* -------------------------------------------------------------------------- */
/*                    EXPORTED UNVERIFIED QUOTES FUNCTIONS                    */
/* -------------------------------------------------------------------------- */

// GetUnverifiedQuotes returns a slice containing all quotes
func GetUnverifiedQuotes() {

}

// StoreUnverifiedQuote stores an unverified quote
func StoreUnverifiedQuote() {

}

// DeleteUnverifiedQuote deletes an unverified quote
func DeleteUnverifiedQuote() {

}

/* -------------------------------------------------------------------------- */
/*                              PRIVATE FUNCTIONS                             */
/* -------------------------------------------------------------------------- */

func loadLocalDatabase() error {
	var err error

	log.Print("Filling cache from PostgreSQL database...")

	// initialize characterLookup table
	setupCharacterLookup()

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		return errors.New("loadLocalDatabase: pinging database failed" + err.Error())
	}

	localDatabase.mux.LockWrite()
	defer localDatabase.mux.UnlockWrite()

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
		return errors.New("createLocalDatabase: loading quotes from databse failed: " + err.Error())
	}

	// initialize wordsMap of localDatabase
	localDatabase.wordsMap = make(map[string]wordsMapT)

	// Iterrate over all quotes from PostgreSQL database
	for rows.Next() {
		// Get id and text of quote
		var q QuoteT
		rows.Scan(&q.QuoteID, &q.TeacherID, &q.Context, &q.Text, &q.Unixtime, &q.Upvotes)

		// add to local database
		// unsafe, because localDatabase is already locked for writing
		err = unsafeAddQuoteToLocalDatabase(q)
		if err != nil {
			return errors.New("createLocalDatabase: adding quote to local database failed: " + err.Error())
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
		return errors.New("createLocalDatabase: loading teachers from databse failed: " + err.Error())
	}

	// Iterate over all teachers from PostgreSQL database
	for rows.Next() {
		// Get teacher data (id, name, title, note)
		var t TeacherT
		rows.Scan(&t.TeacherID, &t.Name, &t.Title, &t.Note)

		// add to local database
		// unsafe, because localDatabase is already locked for writing
		unsafeAddTeacherToLocalDatabase(t)
	}

	rows.Close()

	log.Print("Filled cache successfully")
	return nil
}

// Just adds quote to localDatabase (quoteSlice and wordsMap) without checking q.QuoteID
// using addQuoteToLocalDatabase without checking if q.QuoteID already exists may be fatal
func addQuoteToLocalDatabase(q QuoteT) error {
	localDatabase.mux.LockWrite()
	defer localDatabase.mux.UnlockWrite()

	return unsafeAddQuoteToLocalDatabase(q)
}

func unsafeAddQuoteToLocalDatabase(q QuoteT) error {
	localDatabase.quoteSlice = append(localDatabase.quoteSlice, q)
	enumid := int32(len(localDatabase.quoteSlice) - 1)

	if enumid < 0 {
		return errors.New("addQuoteToLocalDatabase: could not add quote to quoteSlice of localDatabase")
	}

	// Iterrate over all words of quote
	for word, count := range getWordsFromString(q.Text) {
		wordsMapItem := localDatabase.wordsMap[word]
		wordsMapItem.totalOccurences += count

		wordsMapItem.occurenceSlice = append(wordsMapItem.occurenceSlice, occurenceSliceT{enumid, count})

		localDatabase.wordsMap[word] = wordsMapItem
	}

	return nil
}

func addTeacherToLocalDatabase(t TeacherT) {
	localDatabase.mux.LockWrite()
	defer localDatabase.mux.UnlockWrite()
	localDatabase.teacherSlice = append(localDatabase.teacherSlice, t)
}

func unsafeAddTeacherToLocalDatabase(t TeacherT) {
	localDatabase.teacherSlice = append(localDatabase.teacherSlice, t)
}

func overwriteTeacherInLocalDatabase(t TeacherT) error {
	localDatabase.mux.LockWrite()
	defer localDatabase.mux.UnlockWrite()

	affected := false
	for i, v := range localDatabase.teacherSlice {
		if v.TeacherID == t.TeacherID {
			localDatabase.teacherSlice[i] = t
			affected = true
			break
		}
	}

	if affected == false {
		return errors.New("overwriteTeacherInLocalDatabase: could not find specified entry for overwrite")
	}
	return nil
}

func overwriteQuoteInLocalDatabase(q QuoteT) error {
	localDatabase.mux.LockWrite()
	defer localDatabase.mux.UnlockWrite()

	var enumid int32 = -1
	for i, v := range localDatabase.quoteSlice {
		if v.QuoteID == q.QuoteID {
			localDatabase.quoteSlice[i] = q
			enumid = int32(i)
		}
	}

	if enumid < 0 {
		return errors.New("overwriteTeacherInLocalDatabase: could not find specified entry for overwrite")
	}

	wordsFromString := getWordsFromString(q.Text)

	for word, wordsMapItem := range localDatabase.wordsMap {
		for i, v := range wordsMapItem.occurenceSlice {
			if v.enumid == enumid {
				wordCount := wordsFromString[word]
				wordsMapItem.totalOccurences -= wordsMapItem.occurenceSlice[i].count
				if wordCount > 0 {
					wordsMapItem.occurenceSlice[i].count = wordCount
					wordsMapItem.totalOccurences += wordCount
					wordsFromString[word] = 0
					localDatabase.wordsMap[word] = wordsMapItem
				} else if wordsMapItem.totalOccurences == 0 {
					delete(localDatabase.wordsMap, word)
				} else {
					iMax := len(wordsMapItem.occurenceSlice) - 1
					wordsMapItem.occurenceSlice[i] = wordsMapItem.occurenceSlice[iMax]
					wordsMapItem.occurenceSlice[iMax] = occurenceSliceT{0, 0}
					wordsMapItem.occurenceSlice = wordsMapItem.occurenceSlice[:iMax]
					localDatabase.wordsMap[word] = wordsMapItem
				}
				break
			}
		}
	}

	for word, count := range wordsFromString {
		if count > 0 {
			wordsMapItem := localDatabase.wordsMap[word]
			wordsMapItem.totalOccurences += count

			wordsMapItem.occurenceSlice = append(wordsMapItem.occurenceSlice, occurenceSliceT{enumid, count})

			localDatabase.wordsMap[word] = wordsMapItem
		}
	}

	return nil
}

// PrintWordsMap is a debugging function
func PrintWordsMap() {
	log.Print(localDatabase.wordsMap)
}
