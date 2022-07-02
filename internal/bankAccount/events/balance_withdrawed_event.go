package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	BalanceWithdrawnEventType es.EventType = "BALANCE_WITHDRAWN_V1"
)

type BalanceWithdrawnEventV1 struct {
	Amount    int64  `json:"amount"`
	PaymentID string `json:"paymentID"`
	Metadata  []byte `json:"-"`
}
