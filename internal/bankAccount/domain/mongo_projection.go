package domain

import (
	"fmt"
	"time"
)

type Balance struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

func (b *Balance) String() string {
	return fmt.Sprintf("Amount: %f, Currency: %s", b.Amount, b.Currency)
}

type BankAccountMongoProjection struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	AggregateID string    `json:"aggregateID" bson:"aggregateID,omitempty"`
	Email       string    `json:"email" bson:"email,omitempty"`
	Address     string    `json:"address" bson:"address,omitempty"`
	FirstName   string    `json:"firstName" bson:"firstName,omitempty"`
	LastName    string    `json:"lastName" bson:"lastName,omitempty"`
	Balance     Balance   `json:"balance" bson:"balance"`
	Status      string    `json:"status" bson:"status,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt" bson:"updatedAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt" bson:"createdAt,omitempty"`
}

func (b *BankAccountMongoProjection) String() string {
	return fmt.Sprintf("ID: %s, AggregateID: %s, Email: %s, Address: %s, FirstName: %s, LastName: %s, Status: %s, Balance: %s",
		b.ID,
		b.AggregateID,
		b.Email,
		b.Address,
		b.FirstName,
		b.LastName,
		b.Status,
		b.Balance.String(),
	)
}
