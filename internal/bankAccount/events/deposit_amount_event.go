package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	BalanceDepositedEventType es.EventType = "DEPOSIT_BALANCE_EVENT_V1"
)

type BalanceDepositedEventV1 struct {
	Amount    float64 `json:"amount"`
	PaymentID string  `json:"paymentID"`
	Metadata  []byte  `json:"-"`
}

func NewBalanceDepositedEventV1(amount float64, paymentID string) *BalanceDepositedEventV1 {
	return &BalanceDepositedEventV1{Amount: amount, PaymentID: paymentID}
}
