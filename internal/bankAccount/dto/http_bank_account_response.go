package dto

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
)

type HttpBankAccountResponse struct {
	AggregateID string         `json:"aggregateID" bson:"aggregateID,omitempty"`
	Email       string         `json:"email" bson:"email,omitempty"`
	Address     string         `json:"address" bson:"address,omitempty"`
	FirstName   string         `json:"firstName" bson:"firstName,omitempty"`
	LastName    string         `json:"lastName" bson:"lastName,omitempty"`
	Balance     domain.Balance `json:"balance" bson:"balance"`
	Status      string         `json:"status" bson:"status,omitempty"`
}
