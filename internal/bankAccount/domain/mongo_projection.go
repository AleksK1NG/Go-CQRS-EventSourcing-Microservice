package domain

import "time"

type BankAccountMongoProjection struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	AggregateID string    `json:"aggregateID" bson:"aggregateID,omitempty"`
	Email       string    `json:"email" bson:"email,omitempty"`
	Address     string    `json:"address" bson:"address,omitempty"`
	FirstName   string    `json:"firstName" bson:"firstName,omitempty"`
	LastName    string    `json:"lastName" bson:"lastName,omitempty"`
	Balance     float64   `json:"balance" bson:"balance,omitempty"`
	Status      string    `json:"status" bson:"status,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt,omitempty"`
}
