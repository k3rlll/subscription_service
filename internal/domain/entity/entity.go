package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	// ID is the unique identifier for the subscription, generated as a UUID.
	ID uuid.UUID
	// UserID is the unique identifier for the user associated with the subscription, also generated as a UUID.
	UserID      uuid.UUID
	ServiceName string
	//write price as int64 for the future scalability of the project
	Price int64
	// StartDate is the date when the subscription starts. MM-YYYY
	StartDate time.Time
	// EndDate is a pointer to time.Time to allow null values in the database,
	// indicating an active subscription without a defined end date.
	EndDate *time.Time
}
