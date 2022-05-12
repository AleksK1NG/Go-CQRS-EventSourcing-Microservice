package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	BalanceDepositedEventType es.EventType = "DEPOSIT_BALANCE_EVENT_V1"
)

type BalanceDepositedEventV1 struct {
	Amount    float64 `json:"amount"`
	PaymentID string  `json:"paymentID"`
}

func NewBalanceDepositedEventV1(aggregate es.Aggregate, amount float64, paymentID string) (es.Event, error) {
	balanceDepositedEvent := BalanceDepositedEventV1{Amount: amount, PaymentID: paymentID}
	return es.NewEvent(aggregate, BankAccountCreatedEventType, &balanceDepositedEvent)
}
