package database

import (
	"database/sql"
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
// text       the text of the quote itself
// match  	  used by GetQuotesFromString to quantify how well this quote fits the string
// unixtime   optional
type QuoteT struct {
	uidQuote   uint32
	uidTeacher uint32
	text       string
	match      float32
	unixtime   uint64
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

/* -------------------------------------------------------------------------- */
/*                          GLOBAL PACKAGE VARIABLES                          */
/* -------------------------------------------------------------------------- */

// Handle to the PostgreSQL database, used as long time storage
var postgresDatabase *sql.DB

// Created from PostgreSQL database at (re)start
// important: the index of a quote in quoteSlice is called its enumid
// which is used to quickly identify a quote with the wordsMap
var localDatabase struct {
	quoteSlice   []QuoteT
	teacherSlice []TeacherT
	wordsMap     map[string]struct {
		totalOccurences uint32
		occurenceSlice  []struct {
			enumid uint32
			count  uint32
		}
	}
	mux sync.Mutex
}

/* -------------------------------------------------------------------------- */
/*                              GLOBAL FUNCTIONS                              */
/* -------------------------------------------------------------------------- */

// Setup initializes the database backend
// Initialize postgres database
// Create localDatabase from postgresDatabase
func Setup() {

}

// GetTeachers returns a slice containing all teachers
// The returned slice is not sorted
func GetTeachers() {

}

// GetQuotes returns a slice containing all quotes
// The weight variable will be zero
func GetQuotes() {

}

// GetQuotesByString returns a slice containing all quotes
// The weight variable will indicate how well the given text matches the corresponding quote
func GetQuotesByString() {

}

// StoreTeacher stores a new teacher
// If the uid is not zero, StoreTeacher will try to find the corresponding teacher and overwrite it
// If the uid is nil a new teacher will be created
func StoreTeacher() {

}

// StoreQuote stores a new quote
// If the uid is not zero, StoreQuote will try to find the appropriate quote and overwrite it
// If the uid is nil a new quote will be created
func StoreQuote(q QuoteT) {

}

// DeleteQuote deletes the quote corresponding to the given uid from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote() {

}

/* -------------------------------------------------------------------------- */
/*                              PRIVATE FUNCTIONS                             */
/* -------------------------------------------------------------------------- */
