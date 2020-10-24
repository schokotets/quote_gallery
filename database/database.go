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
// QuoteID    the unique ID of the quote
// TeacherID  the unique ID of the corresponding teacher
// Context    the context of the quote
// Text       the text of the quote itself
// Unixtime   the time of submission; optional
// Upvotes    optional
//
// Match      exists only locally, not saved in database!
//              (used by GetQuotesFromString to quantify how well this quote fits the string)
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
// QuoteID      the unique ID of the unverified quote
// TeacherID    the unique ID of the corresponding teacher
// TeacherName  the name of the teacher if no TeacherID is given (e.g. new teacher)
// Context      the context of the quote
// Text         the text of the quote itself
// Unixtime     the time of submission; optional
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

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// Handle to the database, used as long time storage
var database *sql.DB

// globalMutex is to be used if a function of the database package must assure that every other
// function is blocked
var globalMutex Mutex = Mutex{0, 0, false}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED GENERAL FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// Connect establishes the connection to the PostgresSQL database and therefore
// needs to be called before any other function of database.go
//
// Notice: Connect doesn't initialize any tables or the cache, hence Initialize should be called
// right afterwards.
func Connect() error {
	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	var err error

	if database != nil {
		database.Close()
		database = nil
	}

	database, err = sql.Open(
		"postgres",
		`user=postgres 
		password=1234 
		dbname=quote_gallery 
		sslmode=disable`)
	if err != nil {
		return errors.New("Connect: connecting to database failed: " + err.Error())
	}

	return nil
}

// Initialize creates all the required tables in database, if they don't already exist
// and initializes the cache from the database.
//
// Therefore it must be called before any other function of database.go despite Connect, which
// needs to been have called for Initialize to work
func Initialize() error {
	if database == nil {
		return errors.New("Initialize: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return errors.New("Initialize: pinging database failed: " + err.Error())
	}

	// Create teachers table in database if it doesn't exist
	// for more information see TeachersT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS teachers (
		TeacherID serial PRIMARY KEY, 
		Name varchar, 
		Title varchar, 
		Note varchar)`)
	if err != nil {
		database.Close()
		return errors.New("Initialize: creating teachers table failed: " + err.Error())
	}

	// Create quotes table in database if it doesn't exist
	// for more information see QuoteT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS quotes (
		QuoteID serial PRIMARY KEY,
		TeacherID integer REFERENCES teachers (TeacherID) ON DELETE CASCADE, 
		Context varchar,
		Text varchar,
		Unixtime bigint,
		Upvotes integer)`)
	if err != nil {
		database.Close()
		return errors.New("Initialize: creating quotes table failed: " + err.Error())
	}

	// Create unverifiedQuotes table in database if it doesn't exist
	// for more information see UnverifiedQuoteT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS unverifiedQuotes (
		QuoteID serial PRIMARY KEY,
		TeacherID integer REFERENCES teachers (TeacherID) ON DELETE CASCADE, 
		TeacherName varchar, 
		Context varchar,
		Text varchar,
		Unixtime bigint,
		IPHash bigint)`)
	if err != nil {
		database.Close()
		return errors.New("Initialize: creating unverifiedQuotes table failed: " + err.Error())
	}

	unsafeLoadCache()

	return nil

}

// CloseAndClearCache closes database and cache
func CloseAndClearCache() error {
	if database == nil {
		return errors.New("CloseAndClearCache: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	database.Close()
	unsafeClearCache()

	return nil
}

// ExecuteQuery runs a query on the database and returns the error
// This function is to be used in a testing environment
func ExecuteQuery(query string) error {
	if database == nil {
		return errors.New("ExecuteQuery: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	_, err := database.Exec(query)
	return err
}

/* -------------------------------------------------------------------------- */
/*                          EXPORTED QUOTES FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// GetQuotes returns a slice containing all quotes
// The weight variable will be zero
func GetQuotes() (*[]QuoteT, error) {
	if database == nil {
		return nil, errors.New("GetQuotes: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get quotes from cache
	return unsafeGetQuotesFromCache(), nil
}

// GetQuotesByString returns a slice containing all quotes
// The weight variable will indicate how well the given text matches the corresponding quote
func GetQuotesByString(text string) (*[]QuoteT, error) {
	if database == nil {
		return nil, errors.New("GetQuotesByString: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get weighted quotes from cache
	return unsafeGetQuotesByStringFromCache(text), nil
}

// CreateQuote creates a new quote
func CreateQuote(q QuoteT) error {
	if database == nil {
		return errors.New("CreateQuote: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	var err error

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("CreateQuote: pinging database failed: " + err.Error())
	}

	// add quote to database
	err = database.QueryRow(
		`INSERT INTO quotes (TeacherID, Context, Text, Unixtime, Upvotes) VALUES ($1, $2, $3, $4, $5) RETURNING QuoteID`,
		q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes).Scan(&q.QuoteID)
	if err != nil {
		return errors.New("CreateQuote: inserting quote into database failed: " + err.Error())
	}

	// add quote to cache
	unsafeAddQuoteToCache(q)

	return nil
}

// UpdateQuote updates an existing quote by given QuoteID
func UpdateQuote(q QuoteT) error {
	if database == nil {
		return errors.New("UpdateQuote: not connected to database")
	}

	var err error

	if q.QuoteID == 0 {
		return errors.New("UpdateQuote: QuoteID is zero")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("UpdateQuote: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry in database and overwrite it
	var res sql.Result
	res, err = database.Exec(
		`UPDATE quotes SET TeacherID=$2, Context=$3, Text=$4, Unixtime=$5, Upvotes=$6 WHERE QuoteID=$1`,
		q.QuoteID, q.TeacherID, q.Context, q.Text, q.Unixtime, q.Upvotes)
	if err != nil {
		return errors.New("UpdateQuote: updating quote in database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("UpdateQuote: could not find specified database row for updating")
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeOverwriteQuoteInCache(q)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Panic("UpdateQuote: unsafeOverwriteQuoteInCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

// DeleteQuote deletes the quote corresponding to the given ID from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote(ID uint32) error {
	if database == nil {
		return errors.New("DeleteQuote: not connected to database")
	}

	var err error

	if ID == 0 {
		return errors.New("DeleteQuote: ID is zero")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("DeleteQuote: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry in database and delete it
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM quotes WHERE QuoteID=$1`, ID)
	if err != nil {
		return errors.New("DeleteQuote: deleting quote from database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("UpdateQuote: could not find specified database row for deleting")
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeDeleteQuoteFromCache(ID)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Panic("DeleteQuote: unsafeDeleteQuoteFromCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED TEACHERS FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

// GetTeachers returns a slice containing all teachers
// The returned slice is not sorted
func GetTeachers() (*[]TeacherT, error) {
	if database == nil {
		return nil, errors.New("GetTeachers: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get teachers from cache
	return unsafeGetTeachersFromCache(), nil
}

// CreateTeacher creates a new teacher
func CreateTeacher(t TeacherT) error {
	if database == nil {
		return errors.New("CreateTeacher: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	var err error

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("CreateTeacher: pinging database failed: " + err.Error())
	}

	// add teacher to database
	err = database.QueryRow(
		`INSERT INTO teachers (Name, Title, Note) VALUES ($1, $2, $3) RETURNING TeacherID`,
		t.Name, t.Title, t.Note).Scan(&t.TeacherID)
	if err != nil {
		return errors.New("CreateTeacher: inserting teacher into database failed: " + err.Error())
	}

	// add teacher to cache
	unsafeAddTeacherToCache(t)

	return nil
}

// UpdateTeacher updates a teacher by given TeacherID
func UpdateTeacher(t TeacherT) error {
	if database == nil {
		return errors.New("UpdateTeacher: not connected to database")
	}

	var err error

	if t.TeacherID == 0 {
		return errors.New("UpdateTeacher: TeacherID is zero")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("UpdateTeacher: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry in database and overwrite it
	var res sql.Result
	res, err = database.Exec(
		`UPDATE teachers SET Name=$2, Title=$3, Note=$4 WHERE TeacherID=$1`,
		t.TeacherID, t.Name, t.Title, t.Note)
	if err != nil {
		return errors.New("UpdateTeacher: updating teacher in database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("UpdateTeacher: could not find specified database row for updating")
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeOverwriteTeacherInCache(t)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateTeacher() should not fail,
		// so the cache will be reloaded

		log.Panic("UpdateTeacher: unsafeOverwriteTeacherInCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

// DeleteTeacher deletes the teacher corresponding to the given ID from the database and the teachers slice
// It will delete all corresponding quotes
func DeleteTeacher(ID uint32) error {
	if database == nil {
		return errors.New("DeleteTeacher: not connected to database")
	}

	var err error

	if ID == 0 {
		return errors.New("DeleteTeacher: ID is zero")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("DeleteTeacher: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry in database and delete it
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM teachers WHERE TeacherID=$1`, ID)
	if err != nil {
		return errors.New("DeleteTeacher: deleting teacher from database failed: " + err.Error())
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("DeleteTeacher: could not find specified database row for deleting")
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeDeleteTeacherFromCache(ID)
	if err != nil {
		// if this code is executed
		// database was updated successfully but teacher cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Panic("DeleteTeacher: unsafeDeleteTeacherFromCache returned: " + err.Error())
		log.Panic("Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

/* -------------------------------------------------------------------------- */
/*                    EXPORTED UNVERIFIED QUOTES FUNCTIONS                    */
/* -------------------------------------------------------------------------- */

// GetUnverifiedQuotes returns a slice containing all unverified quotes
func GetUnverifiedQuotes() (*[]UnverifiedQuoteT, error) {
	if database == nil {
		return nil, errors.New("GetUnverifiedQuotes: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get all unverifiedQuotes from database
	rows, err := database.Query(`SELECT
		QuoteID,
		TeacherID, 
		TeacherName, 
		Context,
		Text,
		Unixtime,
		IPHash FROM unverifiedQuotes`)
	if err != nil {
		return nil, errors.New("GetUnverifiedQuotes: loading unverifiedQuotes from database failed: " + err.Error())
	}

	var quotes []UnverifiedQuoteT

	// Iterate over all unverifiedQuotes from database
	for rows.Next() {
		// Get unverifiedQuotes data

		var q UnverifiedQuoteT
		var TeacherID sql.NullInt32

		err := rows.Scan(&q.QuoteID, &TeacherID, &q.TeacherName, &q.Context, &q.Text, &q.Unixtime, &q.IPHash)
		if err != nil {
			return nil, errors.New("GetUnverifiedQuotes: parsing unverifiedQuotes failed: " + err.Error())
		}

		// TeacherID can be nill, see CreateUnverifiedQuote and UpdateUnverifiedQuote
		if TeacherID.Valid {
			q.TeacherID = uint32(TeacherID.Int32)
		}

		// Add unverifiedQuote to return slice
		quotes = append(quotes, q)
	}

	return &quotes, nil
}

// CreateUnverifiedQuote stores an unverified quote
func CreateUnverifiedQuote(q UnverifiedQuoteT) error {
	if database == nil {
		return errors.New("CreateUnverifiedQuote: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	var err error

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return errors.New("CreateUnverifiedQuote: pinging database failed: " + err.Error())
	}

	// add quote to database
	if q.TeacherID != 0 {
		_, err = database.Exec(
			`INSERT INTO unverifiedQuotes (TeacherID, TeacherName, Context, Text, Unixtime, IPHash) VALUES ($1, $2, $3, $4, $5, $6)`,
			q.TeacherID, q.TeacherName, q.Context, q.Text, q.Unixtime, q.IPHash)
		if err != nil {
			return errors.New("CreateUnverifiedQuote: inserting quote into database failed: " + err.Error())
		}
	} else {
		_, err = database.Exec(
			`INSERT INTO unverifiedQuotes (TeacherID, TeacherName, Context, Text, Unixtime, IPHash) VALUES ($1, $2, $3, $4, $5, $6)`,
			nil, q.TeacherName, q.Context, q.Text, q.Unixtime, q.IPHash)
		if err != nil {
			return errors.New("CreateUnverifiedQuote: inserting quote into database failed: " + err.Error())
		}
	}

	return nil
}

// UpdateUnverifiedQuote updates an unverified quote
func UpdateUnverifiedQuote(q UnverifiedQuoteT) error {
	if database == nil {
		return errors.New("UpdateUnverifiedQuote: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return errors.New("UpdateUnverifiedQuote: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry database and overwrite it
	var res sql.Result

	if q.TeacherID != 0 {
		res, err = database.Exec(
			`UPDATE unverifiedQuotes SET TeacherID=$2, TeacherName=$3, Context=$4, Text=$5, Unixtime=$6, IPHash=$7 WHERE  QuoteID=$1`,
			q.QuoteID, q.TeacherID, q.TeacherName, q.Context, q.Text, q.Unixtime, q.IPHash)
		if err != nil {
			return errors.New("UpdateUnverifiedQuote: updating unverifiedQuote in database failed: " + err.Error())
		}
	} else {
		res, err = database.Exec(
			`UPDATE unverifiedQuotes SET TeacherID=$2, TeacherName=$3, Context=$4, Text=$5, Unixtime=$6, IPHash=$7 WHERE  QuoteID=$1`,
			q.QuoteID, nil, q.TeacherName, q.Context, q.Text, q.Unixtime, q.IPHash)
		if err != nil {
			return errors.New("UpdateUnverifiedQuote: updating unverifiedQuote in database failed: " + err.Error())
		}
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("UpdateUnverifiedQuote: could not find specified database row for updating")
	}

	return nil
}

// DeleteUnverifiedQuote deletes an unverified quote
func DeleteUnverifiedQuote(ID uint32) error {
	if database == nil {
		return errors.New("DeleteUnverifiedQuote: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return errors.New("DeleteUnverifiedQuote: pinging database failed: " + err.Error())
	}

	// try to find corresponding entry in database and delete it
	_, err = database.Exec(
		`DELETE FROM unverifiedQuotes WHERE  QuoteID=$1`, ID)
	if err != nil {
		return errors.New("DeleteUnverifiedQuote: deleting unverifiedQuote from database failed: " + err.Error())
	}

	return nil
}
