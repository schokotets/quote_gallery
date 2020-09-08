package database

/* -------------------------------------------------------------------------- */
/*                              GLOBAL FUNCTIONS                              */
/* -------------------------------------------------------------------------- */

// Setup initializes the database backend
// *initialize postgres database
// *create quotes and teachers slices from database
// *create words map from database
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

// StoreTeacher stores a new quote in the database and the teachers slice
// If the uuid is not zero, StoreTeacher will try to find the corresponding teacher and overwrite it
// If the uuid is nil a new teacher will be created
func StoreTeacher() {

}

// StoreQuote stores a new quote in the database and the quotes slice
// It will also modifiy the words map
// If the uuid is not zero, StoreQuote will try to find the appropriate quote and overwrite it
// If the uuid is nil a new quote will be created
func StoreQuote() {

}

// DeleteQuote deletes the quote corresponding to the given uuid from the database and the quotes slice
// It will also modifiy the words map
func DeleteQuote() {

}

/* -------------------------------------------------------------------------- */
/*                              PRIVATE FUNCTIONS                             */
/* -------------------------------------------------------------------------- */
