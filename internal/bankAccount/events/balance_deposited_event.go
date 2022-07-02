package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	BalanceDepositedEventType es.EventType = "BALANCE_DEPOSITED_V1"
)

type BalanceDepositedEventV1 struct {
	Amount    int64  `json:"amount"`
	PaymentID string `json:"paymentID"`
	Metadata  []byte `json:"-"`
}

func NewBalanceDepositedEventV1(amount int64, paymentID string) *BalanceDepositedEventV1 {
	return &BalanceDepositedEventV1{Amount: amount, PaymentID: paymentID}
}
