package database

import (
	"database/sql"
	"errors"
	"log"
	"sync"

	// loading postgresql driver
	_ "github.com/lib/pq"
)

/* -------------------------------------------------------------------------- */
/*                                 DEFINITIONS                                */
/* -------------------------------------------------------------------------- */

// QuoteT stores one quote
// uidQuote   the unique identificator of the quote
// uidTeacher the unique identifier of the corresponding teacher
// context	  the context of the quote
// text       the text of the quote itself
// unixtime   optional
//
// match  	  only locally, not safed in PostgreSQL database!
//			  used by GetQuotesFromString to quantify how well this quote fits the string
type QuoteT struct {
	uidQuote   uint32
	uidTeacher uint32
	context    string
	text       string
	unixtime   uint64
	upvotes    uint32
	match      float32
}

// UnverifiedQuoteT stores one unverified quote
// uidQuote    the unique identificator of the unverified quote
// teacher 	   the name of the teacher
// context	   the context of the quote
// text        the text of the quote itself
// unixtime    optional
// iphash	   optional
//
type UnverifiedQuoteT struct {
	uidQuote uint32
	teacher  string
	context  string
	text     string
	unixtime uint64
	ipHash   uint64
}

// TeacherT stores one teacher
// uidTeacher the unique identifier of the teacher
// name       the teacher's name
// title      the teacher's title
// note       optional notes, e.g. subjects
type TeacherT struct {
	uidTeacher uint32
	name       string
	title      string
	note       string
}

type wordsMapT struct {
	totalOccurences uint32
	occurenceSlice  []occurenceSliceT
}

type occurenceSliceT struct {
	enumid uint32
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
	mux          sync.Mutex
}

/* -------------------------------------------------------------------------- */
/*                              GLOBAL FUNCTIONS                              */
/* -------------------------------------------------------------------------- */

// Setup initializes the database backend
// Initialize postgres database
// Create localDatabase from postgresDatabase
func Setup() error {
	var err error

	localDatabase.mux.Lock()
	defer localDatabase.mux.Unlock()

	// Open PostgreSQL database
	postgresDatabase, err = sql.Open(
		"postgres",
		`user=postgres 
		password=1234 
		dbname=quote_gallery 
		sslmode=disable`)
	if err != nil {
		return errors.New("At Setup: " + err.Error())
	}

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		postgresDatabase.Close()
		return errors.New("At Setup: " + err.Error())
	}

	// Create teachers table in PostgreSQL database if it doesn't exist
	// for more information see TeachersT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS teachers (
		uidTeacher serial PRIMARY KEY, 
		name varchar, 
		title varchar, 
		note varchar)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("At Setup: " + err.Error())
	}

	// Create quotes table in PostgreSQL database if it doesn't exist
	// for more information see QuoteT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS quotes (
		uidQuote serial PRIMARY KEY,
		uidTeacher integer REFERENCES teachers (uidTeacher), 
		context varchar,
		text varchar,
		unixtime bigint,
		upvotes integer)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("At Setup: " + err.Error())
	}

	// Create unverifiedQuotes table in PostgreSQL database if it doesn't exist
	// for more information see UnverifiedQuoteT declaration
	_, err = postgresDatabase.Exec(
		`CREATE TABLE IF NOT EXISTS unverifiedQuotes (
		uidQuote serial PRIMARY KEY,
		teacher varchar, 
		context varchar,
		text varchar,
		unixtime bigint,
		ipHash bigint)`)
	if err != nil {
		postgresDatabase.Close()
		return errors.New("At Setup: " + err.Error())
	}

	createLocalDatabase()

	return nil

}

// GetTeachers returns a slice containing all teachers
// The returned slice is not sorted
func GetTeachers() *[]TeacherT {
	localDatabase.mux.Lock()
	teacherSlice := localDatabase.teacherSlice
	localDatabase.mux.Unlock()
	return &teacherSlice
}

// GetQuotes returns a slice containing all quotes
// The weight variable will be zero
func GetQuotes() *[]QuoteT {
	localDatabase.mux.Lock()
	quoteSlice := localDatabase.quoteSlice
	localDatabase.mux.Unlock()
	return &quoteSlice
}

// Close database backend
func Close() {
	localDatabase.quoteSlice = nil
	localDatabase.teacherSlice = nil
	localDatabase.wordsMap = nil
	postgresDatabase.Close()
}

// GetQuotesByString returns a slice containing all quotes
// The weight variable will indicate how well the given text matches the corresponding quote
func GetQuotesByString() {

}

// StoreTeacher stores a new teacher
// If the uid is not zero, StoreTeacher will try to find the corresponding teacher and overwrite it
// If the uid is nil a new teacher will be created
func StoreTeacher(t TeacherT) {

}

// StoreQuote stores a new quote
// If the uid is not zero, StoreQuote will try to find the appropriate quote and overwrite it
// If the uid is zero a new quote will be created
func StoreQuote(q QuoteT) {

}

// DeleteQuote deletes the quote corresponding to the given uid from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote(uid int) {

}

/* -------------------------------------------------------------------------- */
/*                              PRIVATE FUNCTIONS                             */
/* -------------------------------------------------------------------------- */

func createLocalDatabase() error {
	var err error

	log.Print("Creating localDatabase from PostgreSQL database...")

	// initialize characterLookup table
	setupCharacterLookup()

	// initialize wordsMap of localDatabase
	localDatabase.wordsMap = make(map[string]wordsMapT)

	// Verify connection to PostgreSQL database
	err = postgresDatabase.Ping()
	if err != nil {
		return errors.New("At createLocalDatabase: " + err.Error())
	}

	/* --------------------------------- QUOTES --------------------------------- */

	// get all quotes from PostgreSQL database
	rows, err := postgresDatabase.Query(`SELECT 
		uidQuote,
		uidTeacher, 
		context,
		text,
		unixtime,
		upvotes FROM quotes`)

	if err != nil {
		return errors.New("At createLocalDatabase: " + err.Error())
	}
	// Iterrate over all quotes from PostgreSQL database
	for rows.Next() {
		// Get id and text of quote
		var q QuoteT
		rows.Scan(&q.uidQuote, &q.uidTeacher, &q.context, &q.text, &q.unixtime, &q.upvotes)
		// add to local database
		err = addQuoteToLocalDatabase(q)
		if err != nil {
			return errors.New("At createLocalDatabase: " + err.Error())
		}
	}

	rows.Close()

	/* -------------------------------- TEACHERS -------------------------------- */

	// get all teachers from PostgreSQL database
	rows, err = postgresDatabase.Query(`SELECT
		uidTeacher, 
		name, 
		title, 
		note FROM teachers`)

	if err != nil {
		return errors.New("At createLocalDatabase: " + err.Error())
	}
	// Iterate over all teachers from PostgreSQL database
	for rows.Next() {
		// Get id and text of quote
		var t TeacherT
		rows.Scan(&t.uidTeacher, &t.name, &t.title, &t.note)
		// add to local database
		addTeacherToLocalDatabase(t)
	}

	rows.Close()

	log.Print("Done")
	return nil
}

// Just adds quote to localDatabase (quoteSlice and wordsMap) without checking q.uidQuote
// using addQuoteToLocalDatabase without checking if q.uidQuote already exists may be fatal
func addQuoteToLocalDatabase(q QuoteT) error {

	localDatabase.quoteSlice = append(localDatabase.quoteSlice, q)
	enumid := len(localDatabase.quoteSlice) - 1

	if enumid < 0 {
		return errors.New("At addQuoteToLocalDatabase: Could not add quote to quoteSlice of localDatabase")
	}

	// Iterrate over all words of quote
	for word, count := range getWordsFromString(q.text) {
		wordsMapItem := localDatabase.wordsMap[word]
		wordsMapItem.totalOccurences += count

		wordsMapItem.occurenceSlice = append(wordsMapItem.occurenceSlice, occurenceSliceT{uint32(enumid), count})

		localDatabase.wordsMap[word] = wordsMapItem
	}

	return nil
}

func addTeacherToLocalDatabase(t TeacherT) {
	localDatabase.teacherSlice = append(localDatabase.teacherSlice, t)
}
