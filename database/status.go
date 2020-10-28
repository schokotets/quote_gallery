package database

// Status codes
const (
	StatusOK               int32 = 0
	StatusError            int32 = 1
	StatusInvalidTeacherID int32 = 2
	StatusInvalidQuoteID   int32 = 3
)

// Status holds the status code and the status message of any database function
type Status struct {
	Code    int32
	Message string
}
