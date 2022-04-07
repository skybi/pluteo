package metar

import "errors"

// FormatError represents an error in a METAR format
type FormatError struct {
	Wrapping error
	Index    int
}

func (err *FormatError) Error() string {
	return err.Wrapping.Error()
}

var (
	ErrNoASCIIString    = &FormatError{Wrapping: errors.New("the given string has non-ASCII characters")}
	ErrMETARIncomplete  = &FormatError{Wrapping: errors.New("the METAR is incomplete (end of string reached before time was found)")}
	ErrInvalidMETARTime = &FormatError{Wrapping: errors.New("the METAR's issuing time is formatted incorrectly (expected 'ddddddZ')")}
)
