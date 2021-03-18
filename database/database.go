package database

import (
	"database/sql"
	"errors"
	"log"
	"strings"

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
//
// Voted	  exists only locally, not saved in database! \ Created from
// Upvotes    exists only locally, not saved in database! / votes table
// Match      exists only locally, not saved in database!
//              (used by GetQuotesFromString to quantify how well this quote fits the string)
type QuoteT struct {
	QuoteID   int32
	TeacherID int32
	Context   string
	Text      string
	Unixtime  int64
	Voted     bool
	Upvotes   int32
	Match     float32
}

// UnverifiedQuoteT stores one unverified quote
// UserID       the unique ID of the submitting user
// QuoteID      the unique ID of the unverified quote
// TeacherID    the unique ID of the corresponding teacher
// TeacherName  the name of the teacher if no TeacherID is given (e.g. new teacher)
// Context      the context of the quote
// Text         the text of the quote itself
// Unixtime     the time of submission; optional
type UnverifiedQuoteT struct {
	UserID      int32
	QuoteID     int32
	TeacherID   int32
	TeacherName string
	Context     string
	Text        string
	Unixtime    int64
}

// TeacherT stores one teacher
// TeacherID  the unique identifier of the teacher
// Name       the teacher's name
// Title      the teacher's title
// Note       optional notes, e.g. subjects
type TeacherT struct {
	TeacherID int32
	Name      string
	Title     string
	Note      string
}

// UserT stores one user
// UserID    the unique identifier of the user
// Name      the user's name
// Password  the user's password
// Admin     flag if the user has admin priviliges 
type UserT struct {
	UserID   int32
	Name     string
	Password string
	Admin    bool
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
// needs to be called before any other function of database.go.
//
// Notice: Connect doesn't initialize any tables or the cache, hence Initialize should be called
// right afterwards.
//
// Possible returned error types: generic / DBError
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
		return DBError{ "Connect: connecting to database failed", err }
	}

	return nil
}

// Initialize creates all the required tables in database, if they don't already exist
// and initializes the cache from the database.
//
// Therefore it must be called before any other function of database.go despite Connect, which
// needs to been have called for Initialize to work.
//
// Possible returned error types: generic / DBError
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
		return DBError{ "Initialize: pinging database failed", err }
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
		return DBError{ "Initialize: creating teachers table failed", err }
	}

	// Create quotes table in database if it doesn't exist
	// for more information see QuoteT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS quotes (
		QuoteID serial PRIMARY KEY,
		TeacherID integer REFERENCES teachers (TeacherID) ON DELETE CASCADE, 
		Context varchar,
		Text varchar,
		Unixtime bigint)`)
	if err != nil {
		database.Close()
		return DBError{ "Initialize: creating quotes table failed", err }
	}

	// Create unverifiedQuotes table in database if it doesn't exist
	// for more information see UnverifiedQuoteT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS unverifiedQuotes (
		UserID integer REFERENCES users (UserID) ON DELETE CASCADE,
		QuoteID serial PRIMARY KEY,
		TeacherID integer REFERENCES teachers (TeacherID) ON DELETE CASCADE, 
		TeacherName varchar, 
		Context varchar,
		Text varchar,
		Unixtime bigint)`)
	if err != nil {
		database.Close()
		return DBError{ "Initialize: creating unverifiedQuotes table failed", err }
	}

	// Create users table in database if it doesn't exist
	// for more information see UserT declaration
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS users (
		UserID serial PRIMARY KEY,
		Name varchar,
		Password varchar,
		Admin boolean)`)
	if err != nil {
		database.Close()
		return DBError{ "Initialize: creating users table failed", err }
	}

	// Create votes table in database if it doesn't exist
	_, err = database.Exec(
		`CREATE TABLE IF NOT EXISTS votes (
		Hash bigint PRIMARY KEY,
		UserID integer REFERENCES users (UserID) ON DELETE CASCADE,
		QuoteID integer REFERENCES quotes (QuoteID) ON DELETE CASCADE)`)
	if err != nil {
		database.Close()
		return DBError{ "Initialize: creating votes table failed", err }
	}

	unsafeLoadCache()

	return nil
}

// CloseAndClearCache closes database and cache.
//
// Possible returned error type: generic
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
// This function is to be used in a testing environment.
//
// Possible returned error types: generic / DBError
func ExecuteQuery(query string) error {
	if database == nil {
		return errors.New("ExecuteQuery: not connected to database")
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	_, err := database.Exec(query)

	if err != nil {
		return DBError{ "ExecuteQuery: Exec failed", err }
	}
	return nil
}

/* -------------------------------------------------------------------------- */
/*                          EXPORTED QUOTES FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// GetAllQuotes returns a slice containing all quotes.
// The weight variable will be zero.
//
// Possible returned error type: generic
func GetAllQuotes() ([]QuoteT, error) {
	if database == nil {
		return nil, errors.New("GetQuotes: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get quotes from cache
	return unsafeGetAllQuotesFromCache(), nil
}

// GetNQuotesFrom returns a slice containing n quotes
// starting from index from. May return fewer than n quotes.
// The weight variable will be zero
func GetNQuotesFrom(n, from int) ([]QuoteT, error) {
	if database == nil {
		return nil, errors.New("GetQuotes: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get quotes from cache
	return unsafeGetNQuotesFromFromCache(n, from), nil
}

// GetQuotesAmount returns how many quotes there are
func GetQuotesAmount() int {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	return unsafeGetQuotesAmountFromCache()
}

// GetQuotesByString returns a slice containing all quotes.
// The weight variable will indicate how well the given text matches the corresponding quote.
// Possible returned error type: generic
func GetQuotesByString(text string) ([]QuoteT, error) {
	if database == nil {
		return nil, errors.New("GetQuotesByString: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get weighted quotes from cache
	return unsafeGetQuotesByStringFromCache(text), nil
}

// CreateQuote creates a new quote.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError
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
		return DBError{ "CreateQuote: pinging database failed", err }
	}

	// add quote to database
	err = database.QueryRow(
		`INSERT INTO quotes (TeacherID, Context, Text, Unixtime) VALUES ($1, $2, $3, $4) RETURNING QuoteID`,
		q.TeacherID, q.Context, q.Text, q.Unixtime).Scan(&q.QuoteID)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return InvalidTeacherIDError{ "CreateQuote: no teacher with given TeacherID" }
		}
		return DBError{ "CreateQuote: inserting quote into database failed", err }
	}

	// add quote to cache
	unsafeAddQuoteToCache(q)

	return nil
}

// UpdateQuote updates an existing quote by given QuoteID.
// Voted, Upvotes and Unixtime fields will be ignored.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError / InvalidQuoteIDError
func UpdateQuote(q QuoteT) error {
	if database == nil {
		return errors.New("UpdateQuote: not connected to database")
	}

	var err error

	if q.QuoteID == 0 {
		return InvalidQuoteIDError{ "UpdateQuote: QuoteID is zero" }
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "UpdateQuote: pinging database failed: ", err }
	}

	// try to find corresponding entry in database and overwrite it
	var res sql.Result
	res, err = database.Exec(
		`UPDATE quotes SET TeacherID=$2, Context=$3, Text=$4 WHERE QuoteID=$1`,
		q.QuoteID, q.TeacherID, q.Context, q.Text)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return InvalidTeacherIDError{ "UpdateQuote: no teacher with given TeacherID" }
		}
		return DBError{ "UpdateQuote: updating quote in database failed", err }
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidQuoteIDError{ "UpdateQuote: no matching database row found" }
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeOverwriteQuoteInCache(q)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Print("DATABASE: UpdateQuote: unsafeOverwriteQuoteInCache returned: " + err.Error())
		log.Print("DATABASE: Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

// DeleteQuote deletes the quote corresponding to the given ID from the database and the quotes slice.
// It will also modifiy the words map.
//
// Possible returned error types: generic / DBError / InvalidQuoteIDError
func DeleteQuote(ID int32) error {
	if database == nil {
		return errors.New("DeleteQuote: not connected to database")
	}

	var err error

	if ID == 0 {
		return InvalidQuoteIDError{ "DeleteQuote: QuoteID is zero" }
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "DeleteQuote: pinging database failed", err }
	}

	// try to find corresponding entry in database and delete it
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM quotes WHERE QuoteID=$1`, ID)
	if err != nil {
		return DBError{ "DeleteQuote: deleting quote from database failed", err }
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidQuoteIDError{ "UpdateQuote: no matching database row found" }
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeDeleteQuoteFromCache(ID)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Print("DATABASE: DeleteQuote: unsafeDeleteQuoteFromCache returned: " + err.Error())
		log.Print("DATABASE: Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

/* -------------------------------------------------------------------------- */
/*                         EXPORTED TEACHERS FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

// GetTeachers returns a slice containing all teachers.
// The returned slice is not sorted.
//
// Possible returned error type: generic
func GetTeachers() ([]TeacherT, error) {
	if database == nil {
		return nil, errors.New("GetTeachers: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// get teachers from cache
	return unsafeGetTeachersFromCache(), nil
}

// GetTeacherByID returns the teacher corresponding to the given ID.
//
// Possible returned error types: generic / InvalidTeacherIDError
func GetTeacherByID(ID int32) (TeacherT, error) {
	if database == nil {
		return TeacherT{}, errors.New("GetTeacherByID: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	teacher, ok := unsafeGetTeacherByIDFromCache(ID)

	if !ok {
		// Teacher not found
		return TeacherT{}, InvalidTeacherIDError{ "GetTeacherByID: no matching teacher found" }
	}

	return teacher, nil
}


// CreateTeacher creates a new teacher.
//
// Possible returned error types: generic / DBError
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
		return DBError{ "CreateTeacher: pinging database failed", err }
	}

	// add teacher to database
	err = database.QueryRow(
		`INSERT INTO teachers (Name, Title, Note) VALUES ($1, $2, $3) RETURNING TeacherID`,
		t.Name, t.Title, t.Note).Scan(&t.TeacherID)
	if err != nil {
		return DBError{ "CreateTeacher: inserting teacher into database failed", err }
	}

	// add teacher to cache
	unsafeAddTeacherToCache(t)

	return nil
}

// UpdateTeacher updates a teacher by given TeacherID.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError
func UpdateTeacher(t TeacherT) error {
	if database == nil {
		return errors.New("UpdateTeacher: not connected to database")
	}

	var err error

	if t.TeacherID == 0 {
		return InvalidTeacherIDError{ "UpdateTeacher: TeacherID is zero" }
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "UpdateTeacher: pinging database failed: ", err }
	}

	// try to find corresponding entry in database and overwrite it
	var res sql.Result
	res, err = database.Exec(
		`UPDATE teachers SET Name=$2, Title=$3, Note=$4 WHERE TeacherID=$1`,
		t.TeacherID, t.Name, t.Title, t.Note)
	if err != nil {
		return DBError{ "UpdateTeacher: updating teacher in database failed", err }
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidTeacherIDError{ "UpdateTeacher: no matching database row found" }
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeOverwriteTeacherInCache(t)
	if err != nil {
		// if this code is executed
		// database was updated successfully but quote cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateTeacher() should not fail,
		// so the cache will be reloaded

		log.Print("DATABASE: UpdateTeacher: unsafeOverwriteTeacherInCache returned: " + err.Error())
		log.Print("DATABASE: Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

// DeleteTeacher deletes the teacher corresponding to the given ID from the database and the teachers slice.
// It will delete all corresponding quotes.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError
func DeleteTeacher(ID int32) error {
	if database == nil {
		return errors.New("DeleteTeacher: not connected to database")
	}

	var err error

	if ID == 0 {
		return InvalidTeacherIDError{ "DeleteTeacher: TeacherID is zero" }
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// Verify connection to database
	err = database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "DeleteTeacher: pinging database failed: ", err }
	}

	// try to find corresponding entry in database and delete it
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM teachers WHERE TeacherID=$1`, ID)
	if err != nil {
		return DBError{ "DeleteTeacher: deleting teacher from database failed", err }
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidTeacherIDError{ "DeleteTeacher: no matching database row found" }
	}

	// try to find corresponding entry in cache and overwrite it
	err = unsafeDeleteTeacherFromCache(ID)
	if err != nil {
		// if this code is executed
		// database was updated successfully but teacher cannot be found in cache
		// thus cache and database are out of sync
		// because the database is the only source of truth, UpdateQuote() should not fail,
		// so the cache will be reloaded

		log.Print("DATABASE: DeleteTeacher: unsafeDeleteTeacherFromCache returned: " + err.Error())
		log.Print("DATABASE: Cache is out of sync with database, trying to reload")
		go Initialize()
	}

	return nil
}

/* -------------------------------------------------------------------------- */
/*                    EXPORTED UNVERIFIED QUOTES FUNCTIONS                    */
/* -------------------------------------------------------------------------- */

// GetUnverifiedQuotes returns a slice containing all unverified quotes.
//
// Possible returned error types: generic / DBError
func GetUnverifiedQuotes() ([]UnverifiedQuoteT, error) {
	if database == nil {
		return nil, errors.New("GetUnverifiedQuotes: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return nil, DBError{ "GetUnverifiedQuotes: pinging database failed: ", err }
	}

	// get all unverifiedQuotes from database
	rows, err := database.Query(`SELECT
		UserID,
		QuoteID,
		TeacherID, 
		TeacherName, 
		Context,
		Text,
		Unixtime FROM unverifiedQuotes`)
	if err != nil {
		return nil, DBError{ "GetUnverifiedQuotes: loading unverifiedQuotes from database failed", err }
	}

	var quotes []UnverifiedQuoteT

	// Iterate over all unverifiedQuotes from database
	for rows.Next() {
		// Get unverifiedQuotes data

		var q UnverifiedQuoteT
		var TeacherID sql.NullInt32

		err := rows.Scan(&q.UserID, &q.QuoteID, &TeacherID, &q.TeacherName, &q.Context, &q.Text, &q.Unixtime)
		if err != nil {
			return nil, DBError{ "GetUnverifiedQuotes: parsing unverifiedQuotes failed", err }
		}

		// TeacherID can be nill, see CreateUnverifiedQuote and UpdateUnverifiedQuote
		if TeacherID.Valid {
			q.TeacherID = TeacherID.Int32
		}

		// Add unverifiedQuote to return slice
		quotes = append(quotes, q)
	}

	return quotes, nil
}

// GetUnverifiedQuoteByID returns a single unverified quote corresponding to the given ID.
//
// Possible returned error types: generic / DBError / InvalidQuoteIDError
func GetUnverifiedQuoteByID(ID int32) (UnverifiedQuoteT, error) {
	if database == nil {
		return UnverifiedQuoteT{}, errors.New("GetUnverifiedQuoteByID: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return UnverifiedQuoteT{}, DBError{ "GetUnverifiedQuoteByID: pinging database failed: ", err }
	}

	// Query database
	rows, err := database.Query(`SELECT
		UserID,
		TeacherID, 
		TeacherName, 
		Context,
		Text,
		Unixtime FROM unverifiedQuotes WHERE QuoteID=$1`, ID)
	if err != nil {
		return UnverifiedQuoteT{}, DBError{ "GetUnverifiedQuoteByID: loading unverifiedQuote from database failed", err }
	}

	if rows.Next() == false {
		// QuoteID not found
		return UnverifiedQuoteT{}, InvalidQuoteIDError{ "GetUnverifiedQuoteByID: no matching database row found" }
	}

	var q UnverifiedQuoteT
	var TeacherID sql.NullInt32

	err = rows.Scan(&q.UserID, &TeacherID, &q.TeacherName, &q.Context, &q.Text, &q.Unixtime)
	if err != nil {
		return UnverifiedQuoteT{}, DBError{ "GetUnverifiedQuoteByID: parsing unverifiedQuotes failed", err }
	}

	q.QuoteID = ID

	// TeacherID can be nill, see CreateUnverifiedQuote and UpdateUnverifiedQuote
	if TeacherID.Valid {
		q.TeacherID = TeacherID.Int32
	}

	return q, nil
}

// CreateUnverifiedQuote stores an unverified quote.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError
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
		return DBError{ "CreateUnverifiedQuote: pinging database failed", err }
	}

	// add quote to database - by ID or by name
	if q.TeacherID != 0 {
		_, err = database.Exec(
			`INSERT INTO unverifiedQuotes (UserID, TeacherID, TeacherName, Context, Text, Unixtime) VALUES ($1, $2, $3, $4, $5, $6)`,
			q.UserID, q.TeacherID, q.TeacherName, q.Context, q.Text, q.Unixtime)
		if err != nil {
			if strings.Contains(err.Error(), "violates foreign key constraint") {
				return InvalidTeacherIDError{ "CreateUnverifiedQuote: no teacher with given TeacherID" }
			}
			return DBError{ "CreateUnverifiedQuote: inserting quote into database failed", err }
		}
	} else {
		_, err = database.Exec(
			`INSERT INTO unverifiedQuotes (UserID, TeacherID, TeacherName, Context, Text, Unixtime) VALUES ($1, $2, $3, $4, $5, $6)`,
			q.UserID, nil, q.TeacherName, q.Context, q.Text, q.Unixtime)
		if err != nil {
			return DBError{ "CreateUnverifiedQuote: inserting quote into database failed", err }
		}
	}

	return nil
}

// UpdateUnverifiedQuote updates an unverified quote.
// UserID and Unixtime field will be ignored.
//
// Possible returned error types: generic / DBError / InvalidTeacherIDError / InvalidQuoteIDError
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
		return DBError{ "UpdateUnverifiedQuote: pinging database failed", err }
	}

	// try to find corresponding entry database and overwrite it
	var res sql.Result
	if q.TeacherID != 0 {
		res, err = database.Exec(
			`UPDATE unverifiedQuotes SET TeacherID=$2, TeacherName=$3, Context=$4, Text=$5 WHERE  QuoteID=$1`,
			q.QuoteID, q.TeacherID, q.TeacherName, q.Context, q.Text)
		if err != nil {
			if strings.Contains(err.Error(), "violates foreign key constraint") {
				return InvalidTeacherIDError{ "UpdateUnverifiedQuote: no teacher with given TeacherID" }
			}
			return DBError{ "UpdateUnverifiedQuote: updating unverifiedQuote in database failed", err }
		}
	} else {
		res, err = database.Exec(
			`UPDATE unverifiedQuotes SET TeacherID=$2, TeacherName=$3, Context=$4, Text=$5 WHERE  QuoteID=$1`,
			q.QuoteID, nil, q.TeacherName, q.Context, q.Text)
		if err != nil {
			return DBError{ "UpdateUnverifiedQuote: updating unverifiedQuote in database failed", err }
		}
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidQuoteIDError{ "UpdateUnverifiedQuote: no matching database row found" }
	}

	return nil
}

// DeleteUnverifiedQuote deletes an unverified quote.
//
// Possible returned error types: generic / DBError / InvalidQuoteIDError
func DeleteUnverifiedQuote(ID int32) error {
	if database == nil {
		return errors.New("DeleteUnverifiedQuote: not connected to database")
	}

	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "DeleteUnverifiedQuote: pinging database failed", err }
	}

	// try to find corresponding entry in database and delete it
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM unverifiedQuotes WHERE  QuoteID=$1`, ID)
	if err != nil {
		return DBError{ "DeleteUnverifiedQuote: deleting unverifiedQuote from database failed", err }
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return InvalidQuoteIDError{ "DeleteUnverifiedQuote: no matching database row found" }
	}

	return nil
}


/* -------------------------------------------------------------------------- */
/*                          EXPORTED USERS FUNCTIONS                          */
/* -------------------------------------------------------------------------- */

// IsUser checks if a user with the given username and password exists
// If the user exists a UserID != 0 is returned
//
// Possible returned error types: -
func IsUser(name string, password string) int32 {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()
	
	return unsafeGetUserFromCache(name, password).UserID
}


// IsAdmin checks if a user with the given username and password exists and
// if this user has admin priviliges
// If the user exists a UserID != 0 is returned
//
// Possible returned error types: -
func IsAdmin(name string, password string) int32 {
	globalMutex.MinorLock()
	defer globalMutex.MinorUnlock()

	user := unsafeGetUserFromCache(name, password)
	if user.Admin {
		return user.UserID
	}
	return 0
}


/* -------------------------------------------------------------------------- */
/*                          EXPORTED VOTING FUNCTIONS                         */
/* -------------------------------------------------------------------------- */

// AddVote adds a vote from one user for one quote to the database
func AddVote(u int32, q int32) error {
	if u < 1 {
		// u must be greater than zero to be a valid UserID
		return errors.New("AddVote: invalid UserID, must be greater than zero")
	}
	
	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "AddVote: pinging database failed", err }
	}

	// add vote to database
	var res sql.Result
	res, err = database.Exec(
		`INSERT INTO votes (Hash, UserID, QuoteID) VALUES ($1, $2, $3) 
		 ON CONFLICT (Hash) DO NOTHING `, voteHash(u, q), u, q)
	if err != nil {
		return DBError{ "AddVote: inserting vote into database failed", err }
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		// already voted - that's not a problem
		return nil
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// add vote to cache
	unsafeAddVoteToCache(u, q)
	
	return nil
}

// DeleteVote deletes a vote from one user for one quote from the database
func DeleteVote(u int32, q int32) error {
	if u < 1 {
		// u must be greater than zero to be a valid UserID
		return errors.New("DeleteVote: invalid UserID, must be greater than zero")
	}
	
	// Verify connection to database
	err := database.Ping()
	if err != nil {
		database.Close()
		return DBError{ "DeleteVote: pinging database failed", err }
	}

	// add vote to database
	var res sql.Result
	res, err = database.Exec(
		`DELETE FROM votes WHERE Hash = $1`, voteHash(u, q))
	if err != nil {
		return DBError{ "DeleteVote: deleting vote from database failed", err }
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		// not voted - that's not a problem
		return nil
	}

	globalMutex.MajorLock()
	defer globalMutex.MajorUnlock()

	// add vote to cache
	unsafeDeleteVoteFromCache(u, q)
	
	return nil
}


/* -------------------------------------------------------------------------- */
/*                         UNEXPORTED HELPER FUNCTIONS                        */
/* -------------------------------------------------------------------------- */

func voteHash(u int32, q int32) int64 {
	return int64(u)<<32 | int64(q)
}
