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
		bankAccountCreatedEventV1 := events.BankAccountCreatedEventV1{}
		if err := event.GetJsonData(&bankAccountCreatedEventV1); err != nil {
			return err
		}
		a.onBankAccountCreated(bankAccountCreatedEventV1)
		return nil
	case events.BalanceDepositedEventType:
		balanceDepositedEvent := events.BalanceDepositedEventV1{}
		if err := event.GetJsonData(&balanceDepositedEvent); err != nil {
			return err
		}
		a.onBalanceDeposited(balanceDepositedEvent)
		return nil
	case events.EmailChangedEventType:
		emailChangedEvent := events.EmailChangedEventV1{}
		if err := event.GetJsonData(&emailChangedEvent); err != nil {
			return err
		}
		a.onEmailChanged(emailChangedEvent)
		return nil
	default:
		return bankAccountErrors.ErrUnknownEventType
	}
}

func (a *BankAccountAggregate) onBankAccountCreated(event events.BankAccountCreatedEventV1) {
	a.BankAccount.Email = event.Email
	a.BankAccount.Address = event.Address
	if event.Balance > 0 {
		a.BankAccount.DepositBalance(event.Balance)
	} else {
		a.BankAccount.Balance = 0
	}
	a.BankAccount.Balance = event.Balance
	a.BankAccount.FirstName = event.FirstName
	a.BankAccount.LastName = event.LastName
	a.BankAccount.Status = event.Address
}

func (a *BankAccountAggregate) onBalanceDeposited(event events.BalanceDepositedEventV1) {
	a.BankAccount.Balance += event.Amount
}

func (a *BankAccountAggregate) onEmailChanged(event events.EmailChangedEventV1) {
	a.BankAccount.Email = event.Email
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
