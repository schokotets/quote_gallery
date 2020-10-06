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
// QuoteID    the unique identificator of the quote
// TeacherID  the unique identifier of the corresponding teacher
// Context    the context of the quote
// Text       the text of the quote itself
// Unixtime   optional
// Upvotes    optional
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
// QuoteID      the unique identificator of the unverified quote
// TeacherID    the unique identifier of the corresponding teacher
// TeacherName  the name of the teacher if no TeacherID is given (e.g. new teacher)
// Context      the context of the quote
// Text         the text of the quote itself
// Unixtime     optional
// IPHash       optional
type UnverifiedQuoteT struct {
	QuoteID     uint32
	TeacherID   uint32
	TeacherName string
	Context     string
	Text        string
	Unixtime    uint64
	IPHash      uint64
}

// TeacherT stores one teacher
// TeacherID  the unique identifier of the teacher
// Name       the teacher's name
// Title      the teacher's title
// Note       optional notes, e.g. subjects
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
	enumID int32
	count  uint32
}

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// Handle to the PostgreSQL database, used as long time storage
var postgresDatabase *sql.DB

// globalMutex is to be used if a function of the database package must assure that every other
// function is blocked
var globalMutex Mutex = Mutex{0, 0, false}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED GENERAL FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// Setup initializes the database backend
// Initialize postgres database
// Create cache from postgresDatabase
//
// Must be called once at startup before any of the other database functions
// Can be called during operation to reload the cache
func Setup() error {
	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	var err error

	if postgresDatabase != nil {
		postgresDatabase.Close()
		postgresDatabase = nil
	}

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
		TeacherID integer REFERENCES teachers (TeacherID), 
		TeacherName varchar, 
		Context varchar,
		Text varchar,
		Unictime bigint,
		IPHash bigint)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("Setup: creating unverifiedQuotes table failed: " + err.Error())
	}

	unsafeLoadCache()

	return nil

}

// CloseAndClearCache closes postgreSQL database and cache
func CloseAndClearCache() {
	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	postgresDatabase.Close()
	unsafeClearCache()
}

// ExecuteQuery runs a query on the database and returns the error
func ExecuteQuery(query string) error {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	_, err := postgresDatabase.Exec(query)
	return err
}

/* -------------------------------------------------------------------------- */
/*                          EXPORTED QUOTES FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// GetQuotes returns a slice containing all quotes
// The weight variable will be zero
func GetQuotes() *[]QuoteT {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get quotes from cache
	return getQuotesFromCache()
}

// GetQuotesByString returns a slice containing all quotes
// The weight variable will indicate how well the given text matches the corresponding quote
func GetQuotesByString(text string) *[]QuoteT {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get weighted quotes from cache
	return getQuotesByStringFromCache(text)
}

// CreateQuote creates a new quote
func CreateQuote(q QuoteT) error {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	var err error

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("CreateQuote: pinging database failed: " + err.Error())
	}

	// add quote to postgresDatabase
	err = postgresDatabase.QueryRow(
		`INSERT INTO quotes (TeacherID, Context, Text, Unixtime, Upvotes) VALUES ($1, $2, $3, $4, $5) RETURNING QuoteID`,
		q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes).Scan(&q.QuoteID)
	if err != nil {
		return errors.New("CreateQuote: inserting quote into database failed: " + err.Error())
	}

	// add quote to cache
	addQuoteToCache(q)

	return nil
}

// UpdateQuote updates an existing quote by given QuoteID
func UpdateQuote(q QuoteT) error {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	var err error

	if q.QuoteID == 0 {
		return errors.New("UpdateQuote: QuoteID is zero")
	}

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("UpdateQuote: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry postgresDatabase and overwrite it
	var res sql.Result
	res, err = postgresDatabase.Exec(
		`UPDATE quotes SET TeacherID=$2, Context=$3, Text=$4, Unixtime=$5, Upvotes=$6 WHERE QuoteID=$1`,
		q.QuoteID, q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes)
	if err != nil {
		return errors.New("UpdateQuote: updating quote in database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("UpdateQuote: could not find specified database row for updating")
	}

	// try to find corresponding entry in cache and overwrite it
	err = overwriteQuoteInCache(q)
	if err != nil {
		// is this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because of the database being the only source of truth, UpdateQuote() should not fail
		// but the cache will be reloaded

		log.Panic("UpdateQuote: overwriteQuoteInCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Setup()
	}

	return nil
}

// DeleteQuote deletes the quote corresponding to the given ID from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote(ID int) {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED TEACHERS FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

// GetTeachers returns a slice containing all teachers
// The returned slice is not sorted
func GetTeachers() *[]TeacherT {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get teachers from cache
	return getTeachersFromCache()
}

// CreateTeacher creates a new teacher
func CreateTeacher(t TeacherT) error {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	var err error

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("CreateTeacher: pinging database failed: " + err.Error())
	}

	// add teacher to postgresDatabase
	err = postgresDatabase.QueryRow(
		`INSERT INTO teachers (Name, Title, Note) VALUES ($1, $2, $3) RETURNING TeacherID`,
		t.Name, t.Title, t.Note).Scan(&t.TeacherID)
	if err != nil {
		return errors.New("CreateTeacher: inserting teacher into database failed: " + err.Error())
	}

	// add teacher to cache
	addTeacherToCache(t)

	return nil
}

// UpdateTeacher updates a teacher by given TeacherID
func UpdateTeacher(t TeacherT) error {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	var err error

	if t.TeacherID == 0 {
		return errors.New("UpdateTeacher: TeacherID is zero")
	}

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("UpdateTeacher: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry postgresDatabase and overwrite it
	var res sql.Result
	res, err = postgresDatabase.Exec(
		`UPDATE teachers SET Name=$2, Title=$3, Note=$4 WHERE TeacherID=$1`,
		t.TeacherID, t.Name, t.Title, t.Note)
	if err != nil {
		return errors.New("UpdateTeacher: updating teacher in database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected != 0 {
		return errors.New("UpdateTeacher: could not find specified database row for updating")
	}

	// try to find corresponding entry in cache and overwrite it
	err = overwriteTeacherInCache(t)
	if err != nil {
		// is this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because of the database being the only source of truth, UpdateQuote() should not fail
		// but the cache will be reloaded

		log.Panic("UpdateQuote: overwriteQuoteInCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Setup()
	}

	return nil
}

// DeleteTeacher deletes the teacher corresponding to the given ID from the database and the teachers slice
// It will delete all corresponding quotes
func DeleteTeacher() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}

/* -------------------------------------------------------------------------- */
/*                    EXPORTED UNVERIFIED QUOTES FUNCTIONS                    */
/* -------------------------------------------------------------------------- */

// GetUnverifiedQuotes returns a slice containing all unverified quotes
func GetUnverifiedQuotes() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}

// CreateUnverifiedQuote stores an unverified quote
func CreateUnverifiedQuote() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}

// UpdateUnverifiedQuote updates an unverified quote
func UpdateUnverifiedQuote() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}

// DeleteUnverifiedQuote deletes an unverified quote
func DeleteUnverifiedQuote() {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
}
