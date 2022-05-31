package domain

import (
	"context"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
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

	bankAccountAggregate := &BankAccountAggregate{BankAccount: NewBankAccount(id)}
	aggregateBase := es.NewAggregateBase(bankAccountAggregate.When)
	aggregateBase.SetType(BankAccountAggregateType)
	aggregateBase.SetID(id)
	bankAccountAggregate.AggregateBase = aggregateBase
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
		return bankAccountErrors.ErrUnknownEventType
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
		a.BankAccount.DepositBalance(bankAccountCreatedEventV1.Balance)
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

func (a *BankAccountAggregate) CreateNewBankAccount(ctx context.Context, email, address, firstName, lastName, status string, balance float64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.CreateNewBankAccount")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	// TODO: check email availability

	event, err := events.NewBankAccountCreatedEventV1(a, email, address, firstName, lastName, status, balance)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		return tracing.TraceWithErr(span, errors.Wrap(err, "SetMetadata"))
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) DepositBalance(ctx context.Context, amount float64, paymentID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.DepositBalance")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	event, err := events.NewBalanceDepositedEventV1(a, amount, paymentID)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		return tracing.TraceWithErr(span, errors.Wrap(err, "SetMetadata"))
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) ChangeEmail(ctx context.Context, email string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.ChangeEmail")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	// TODO: check email availability

	event, err := events.NewEmailChangedEventV1(a, email)
	if err != nil {
		return tracing.TraceWithErr(span, err)
	}

	if err := event.SetMetadata(tracing.ExtractTextMapCarrier(span.Context())); err != nil {
		return tracing.TraceWithErr(span, errors.Wrap(err, "SetMetadata"))
	}

	return a.Apply(event)
}
