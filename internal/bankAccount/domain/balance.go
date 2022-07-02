package domain

import "fmt"

type Balance struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

func (b *Balance) String() string {
	return fmt.Sprintf("Amount: %f, Currency: %s", b.Amount, b.Currency)
}
