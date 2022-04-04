package metar

import (
	"errors"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrMETARIncomplete  = errors.New("the METAR is incomplete (end of string reached before time was found)")
	ErrInvalidMETARTime = errors.New("the METAR's issuing time is formatted incorrectly (expected 'ddddddZ')")
)

// METAR represents a stored METAR data point
type METAR struct {
	ID        uuid.UUID `json:"id"`
	StationID string    `json:"station_id"`
	IssuedAt  int64     `json:"issued_at"`
	Raw       string    `json:"raw"`
}

// OfString tries to decode a raw METAR string into a METAR object.
// This method is no replacement to a fully-featured METAR decoder & validator as it only reads and validates the METAR
// until the timestamp was decoded successfully.
func OfString(raw string) (*METAR, error) {
	raw = strings.TrimSpace(raw)

	// The report type (METAR or SPECI) is not included in the METAR object as no currently known data source provides them.
	// We still trim them as they are technically correct in a raw METAR and thus MAY be provided by a feeder.
	if strings.HasPrefix(raw, "METAR") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "METAR"))
	} else if strings.HasPrefix(raw, "SPECI") {
		raw = strings.TrimSpace(strings.TrimPrefix(raw, "SPECI"))
	}

	// To prevent data inconsistency, we will not include the report type either way
	initial := raw

	// The next 4 characters represent the station's ICAO code
	if len(raw) < 4 {
		return nil, ErrMETARIncomplete
	}
	stationID := raw[:4]
	raw = strings.TrimSpace(raw[4:])

	// The next 7 characters have to consist of 6 digits + a literal 'Z' as they represent the time the METAR was issued
	if len(raw) < 7 {
		return nil, ErrMETARIncomplete
	}
	timeSection := raw[:7]
	if timeSection[6] != 'Z' {
		return nil, ErrInvalidMETARTime
	}
	for i := 0; i < 6; i++ {
		if !unicode.IsDigit(rune(timeSection[i])) {
			return nil, ErrInvalidMETARTime
		}
	}

	// Now we need to split the digits into 3 parts of 2 digits each (dddddd -> dd dd dd).
	// The first pair represents the day of the current month, the second one the hour and the third one the minutes
	// the METAR was issued.
	//
	// Example: '161350' -> 16th day of the current month at 13:50 Zulu (UTC)
	day, _ := strconv.Atoi(timeSection[:2])
	hour, _ := strconv.Atoi(timeSection[2:4])
	minutes, _ := strconv.Atoi(timeSection[4:6])
	now := time.Now()
	issuedAt := time.Date(now.Year(), now.Month(), day, hour, minutes, 0, 0, time.UTC)

	return &METAR{
		ID:        uuid.New(),
		StationID: stationID,
		IssuedAt:  issuedAt.Unix(),
		Raw:       initial,
	}, nil
}
