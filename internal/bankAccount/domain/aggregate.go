package domain

import (
	"context"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
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

func (a *BankAccountAggregate) When(event any) error {
	switch evt := event.(type) {
	case *events.BankAccountCreatedEventV1:
		a.onBankAccountCreated(evt)
		return nil
	case *events.BalanceDepositedEventV1:
		a.onBalanceDeposited(evt)
		return nil
	case *events.EmailChangedEventV1:
		a.onEmailChanged(evt)
		return nil
	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "event: %+v", event)
	}
}

func (a *BankAccountAggregate) onBankAccountCreated(event *events.BankAccountCreatedEventV1) {
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

func (a *BankAccountAggregate) onBalanceDeposited(event *events.BalanceDepositedEventV1) {
	a.BankAccount.Balance += event.Amount
}

func (a *BankAccountAggregate) onEmailChanged(event *events.EmailChangedEventV1) {
	a.BankAccount.Email = event.Email
}

func (a *BankAccountAggregate) CreateNewBankAccount(ctx context.Context, email, address, firstName, lastName, status string, balance float64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.CreateNewBankAccount")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	event := &events.BankAccountCreatedEventV1{
		Email:     email,
		Address:   address,
		FirstName: firstName,
		LastName:  lastName,
		Balance:   balance,
		Status:    status,
	}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}
	event.Metadata = metaDataBytes

	return a.Apply(event)
}

func (a *BankAccountAggregate) DepositBalance(ctx context.Context, amount float64, paymentID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.DepositBalance")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	event := &events.BalanceDepositedEventV1{
		Amount:    amount,
		PaymentID: paymentID,
	}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}
	event.Metadata = metaDataBytes

	return a.Apply(event)
}

func (a *BankAccountAggregate) ChangeEmail(ctx context.Context, email string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.ChangeEmail")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	event := &events.EmailChangedEventV1{Email: email}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}
	event.Metadata = metaDataBytes

	return a.Apply(event)
}
