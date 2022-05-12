package domain

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/pkg/errors"
)

const (
	BankAccountAggregateType es.AggregateType = "BankAccount"
)

type BankAccountAggregate struct {
	*es.AggregateBase
	BankAccount *BankAccount
}

func NewBankAccountAggregate(id string) *BankAccountAggregate {
	if id == "" {
		return nil
	}

	bankAccountAggregate := &BankAccountAggregate{BankAccount: NewBankAccount()}
	aggregateBase := es.NewAggregateBase(bankAccountAggregate.When)
	aggregateBase.SetType(BankAccountAggregateType)
	bankAccountAggregate.AggregateBase = aggregateBase
	bankAccountAggregate.SetID(id)
	return bankAccountAggregate
}

func (a *BankAccountAggregate) When(event es.Event) error {
	switch event.GetEventType() {
	case events.BankAccountCreatedEventType:
		return a.onBankAccountCreated(event)
	case events.BalanceDepositedEventType:
		return a.onBalanceDeposited(event)
	case events.EmailChangedEventType:
		return a.onEmailChanged(event)
	default:
		return errors.New("unknown event type")
	}
}

func (a *BankAccountAggregate) onBankAccountCreated(event es.Event) error {
	bankAccountCreatedEventV1 := events.BankAccountCreatedEventV1{}
	if err := event.GetJsonData(&bankAccountCreatedEventV1); err != nil {
		return err
	}
	a.BankAccount.Email = bankAccountCreatedEventV1.Email
	a.BankAccount.Address = bankAccountCreatedEventV1.Address
	if bankAccountCreatedEventV1.Balance > 0 {
		a.BankAccount.Balance = bankAccountCreatedEventV1.Balance
	} else {
		a.BankAccount.Balance = 0
	}
	a.BankAccount.Balance = bankAccountCreatedEventV1.Balance
	a.BankAccount.FirstName = bankAccountCreatedEventV1.FirstName
	a.BankAccount.LastName = bankAccountCreatedEventV1.LastName
	a.BankAccount.Status = bankAccountCreatedEventV1.Address
	return nil
}

func (a *BankAccountAggregate) onBalanceDeposited(event es.Event) error {
	balanceDepositedEvent := events.BalanceDepositedEventV1{}
	if err := event.GetJsonData(&balanceDepositedEvent); err != nil {
		return err
	}
	a.BankAccount.Balance += balanceDepositedEvent.Amount
	return nil
}

func (a *BankAccountAggregate) onEmailChanged(event es.Event) error {
	emailChangedEvent := events.EmailChangedEventV1{}
	if err := event.GetJsonData(&emailChangedEvent); err != nil {
		return err
	}
	a.BankAccount.Email = emailChangedEvent.Email
	return nil
}

func (a *BankAccountAggregate) CreateNewBankAccount(ctx context.Context, email, address, firstName, lastName string, balance float64, status string) error {
	// TODO: check email availability

	event, err := events.NewBankAccountCreatedEventV1(a, email, address, firstName, lastName, balance, status)
	if err != nil {
		return err
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) DepositBalance(ctx context.Context, amount float64, paymentID string) error {
	event, err := events.NewBalanceDepositedEventV1(a, amount, paymentID)
	if err != nil {
		return err
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) ChangeEmail(ctx context.Context, email string) error {
	// TODO: check email availability

	event, err := events.NewEmailChangedEventV1(a, email)
	if err != nil {
		return err
	}

	return a.Apply(event)
}
