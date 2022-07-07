package domain

import (
	"context"
	bankAccountErrors "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/Rhymond/go-money"
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
		a.BankAccount.Email = evt.Email
		a.BankAccount.Address = evt.Address
		a.BankAccount.Balance = evt.Balance
		a.BankAccount.FirstName = evt.FirstName
		a.BankAccount.LastName = evt.LastName
		a.BankAccount.Status = evt.Status
		return nil

	case *events.BalanceDepositedEventV1:
		return a.BankAccount.DepositBalance(evt.Amount)

	case *events.BalanceWithdrawnEventV1:
		return a.BankAccount.WithdrawBalance(evt.Amount)

	case *events.EmailChangedEventV1:
		a.BankAccount.Email = evt.Email
		return nil

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "event: %#v", event)
	}
}

func (a *BankAccountAggregate) CreateNewBankAccount(ctx context.Context, email, address, firstName, lastName, status string, amount int64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.CreateNewBankAccount")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if amount < 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}

	event := &events.BankAccountCreatedEventV1{
		Email:     email,
		Address:   address,
		FirstName: firstName,
		LastName:  lastName,
		Balance:   money.New(amount, money.USD),
		Status:    status,
		Metadata:  metaDataBytes,
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) DepositBalance(ctx context.Context, amount int64, paymentID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.DepositBalance")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if amount <= 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}

	event := &events.BalanceDepositedEventV1{
		Amount:    amount,
		PaymentID: paymentID,
		Metadata:  metaDataBytes,
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) WithdrawBalance(ctx context.Context, amount int64, paymentID string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.WithdrawBalance")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	if amount <= 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}

	balance, err := money.New(a.BankAccount.Balance.Amount(), money.USD).Subtract(money.New(amount, money.USD))
	if err != nil {
		return errors.Wrapf(err, "Balance.Subtract amount: %d", amount)
	}

	if balance.IsNegative() {
		return errors.Wrapf(bankAccountErrors.ErrNotEnoughBalance, "amount: %d", amount)
	}

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}

	event := &events.BalanceWithdrawnEventV1{
		Amount:    amount,
		PaymentID: paymentID,
		Metadata:  metaDataBytes,
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) ChangeEmail(ctx context.Context, email string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "BankAccountAggregate.ChangeEmail")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", a.GetID()))

	metaDataBytes, err := serializer.Marshal(tracing.ExtractTextMapCarrier(span.Context()))
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}

	event := &events.EmailChangedEventV1{Email: email, Metadata: metaDataBytes}

	return a.Apply(event)
}
