package metar

import "errors"

// FormatError represents an error in a METAR format
type FormatError struct {
	Wrapping error
}

func (err *FormatError) Error() string {
	return err.Wrapping.Error()
}

var (
	ErrMETARIncomplete  = &FormatError{Wrapping: errors.New("the METAR is incomplete (end of string reached before time was found)")}
	ErrInvalidMETARTime = &FormatError{Wrapping: errors.New("the METAR's issuing time is formatted incorrectly (expected 'ddddddZ')")}
)
