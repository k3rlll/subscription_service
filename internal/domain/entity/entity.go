package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ServiceName string
	//write price as int64 for the future scalability of the project
	Price     int64
	StartDate time.Time
	EndDate   *time.Time
}
