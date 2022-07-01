package domain

import (
	"fmt"
	"github.com/Rhymond/go-money"
)

type BankAccount struct {
	AggregateID string       `json:"aggregateID"`
	Email       string       `json:"email"`
	Address     string       `json:"address"`
	FirstName   string       `json:"firstName"`
	LastName    string       `json:"lastName"`
	Balance     *money.Money `json:"balance"`
	Status      string       `json:"status"`
}

func NewBankAccount(id string) *BankAccount {
	return &BankAccount{AggregateID: id, Balance: money.New(0, money.USD)}
}

func (b *BankAccount) DepositBalance(amount int64) error {
	result, err := b.Balance.Add(money.New(amount, money.USD))
	if err != nil {
		return err
	}
	b.Balance = result
	return nil
}

func (b *BankAccount) WithdrawBalance(amount int64) error {
	result, err := b.Balance.Subtract(money.New(amount, money.USD))
	if err != nil {
		return err
	}
	b.Balance = result
	return nil
}

func (b *BankAccount) String() string {
	return fmt.Sprintf("AggregateID: %s, Email: %s, Address: %s, FirstName: %s, LastName: %s, Status: %s, Balance: %s",
		b.AggregateID,
		b.Email,
		b.Address,
		b.FirstName,
		b.LastName,
		b.Status,
		b.Balance.Display(),
	)
}
