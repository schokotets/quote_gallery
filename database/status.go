package database

// InvalidUserIDError is used when the UserID is invalid
type InvalidUserIDError struct {
	Message string
}

func (err InvalidUserIDError) Error() string {
	return err.Message
}

// InvalidTeacherIDError is used when the TeacherID is invalid
type InvalidTeacherIDError struct {
	Message string
}

func (err InvalidTeacherIDError) Error() string {
	return err.Message
}

// InvalidQuoteIDError is used when the QuoteID is invalid
type InvalidQuoteIDError struct {
	Message string
}

func (err InvalidQuoteIDError) Error() string {
	return err.Message
}

// DBError is used when unspecific database operations fail / rows.Scan fails
type DBError struct {
	Message string
	InnerErr error
}

func (err DBError) Error() string {
	if err.InnerErr != nil {
		return err.Message + ": " + err.InnerErr.Error()
	}
	return err.Message
}
