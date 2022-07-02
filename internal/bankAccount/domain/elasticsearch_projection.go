package domain

import (
	"fmt"
	"time"
)

type ElasticSearchProjection struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	AggregateID string    `json:"aggregateID" bson:"aggregateID,omitempty"`
	Version     uint64    `json:"version" bson:"version"`
	Email       string    `json:"email" bson:"email,omitempty"`
	Address     string    `json:"address" bson:"address,omitempty"`
	FirstName   string    `json:"firstName" bson:"firstName,omitempty"`
	LastName    string    `json:"lastName" bson:"lastName,omitempty"`
	Balance     Balance   `json:"balance" bson:"balance"`
	Status      string    `json:"status" bson:"status,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt,omitempty"`
}

func (b *ElasticSearchProjection) String() string {
	return fmt.Sprintf("ID: %s, AggregateID: %s, Version: %d, Email: %s, Address: %s, FirstName: %s, LastName: %s, Status: %s, Balance: %s,  UpdatedAt: %s, CreatedAt: %s",
		b.ID,
		b.AggregateID,
		b.Version,
		b.Email,
		b.Address,
		b.FirstName,
		b.LastName,
		b.Status,
		b.Balance.String(),
		b.UpdatedAt,
		b.CreatedAt,
	)
}
