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
}

func NewBankAccountCreatedEventV1(aggregate es.Aggregate, email, address, firstName, lastName string, balance float64, status string) (es.Event, error) {
	bankAccountCreatedEvent := BankAccountCreatedEventV1{Email: email, Address: address, FirstName: firstName, LastName: lastName, Balance: balance, Status: status}
	return es.NewEvent(aggregate, BalanceDepositedEventType, &bankAccountCreatedEvent)
}
