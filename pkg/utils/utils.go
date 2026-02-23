package utils

import (
	"errors"
	"time"
)

// ParseDate takes startDate and endDate as strings in the format "MM-YYYY" and parses them into time.Time objects.
// It returns the parsed start time, a pointer to the parsed end time (which can be nil if endDate is not provided),
//
//	and an error if any parsing fails or if the end date is before the start date.
func ParseDate(startDate, endDate string) (time.Time, *time.Time, error) {
	startTime, err := time.Parse("01-2006", startDate)
	if err != nil {
		return time.Time{}, nil, err
	}
	//

	// end date can be null, because the subscription can be active, so we will use pointer to time.Time
	var endTimePtr *time.Time = nil //

	if endDate != "" { // if end date is provided
		parsedEndTime, err := time.Parse("01-2006", endDate)
		if err != nil {
			return time.Time{}, nil, err
		}

		endTimePtr = &parsedEndTime // keep the pointer to end time if it is provided and valid
	}
	if endTimePtr != nil && endTimePtr.Before(startTime) {
		return time.Time{}, nil, errors.New("end date is before start date")
	}
	return startTime, endTimePtr, nil
}
