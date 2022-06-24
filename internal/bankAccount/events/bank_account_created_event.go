package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	BankAccountCreatedEventType es.EventType = "BANK_ACCOUNT_CREATED_EVENT_V1"
)

type BankAccountCreatedEventV1 struct {
	Email     string  `json:"email"`
	Address   string  `json:"address"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
	Metadata  []byte  `json:"-"`
}

func NewBankAccountCreatedEventV1(email, address, firstName, lastName, status string, balance float64) *BankAccountCreatedEventV1 {
	return &BankAccountCreatedEventV1{
		Email:     email,
		Address:   address,
		FirstName: firstName,
		LastName:  lastName,
		Balance:   balance,
		Status:    status,
	}
}
