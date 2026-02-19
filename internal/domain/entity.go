package domain

import "github.com/google/uuid"

type Subscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"service_name"`
	Price     int       `json:"price"`
	StartDate string    `json:"start_date"`         
	EndDate   *string   `json:"end_date,omitempty"`
}
