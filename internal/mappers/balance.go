package mappers

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/Rhymond/go-money"
)

func BalanceFromMoney(money *money.Money) domain.Balance {
	return domain.Balance{
		Amount:   money.AsMajorUnits(),
		Currency: money.Currency().Code,
	}
}

func BalanceToGrpc(balance domain.Balance) *bankAccountService.Balance {
	return &bankAccountService.Balance{
		Amount:   balance.Amount,
		Currency: balance.Currency,
	}
}

func BalanceMoneyToGrpc(money *money.Money) *bankAccountService.Balance {
	return &bankAccountService.Balance{
		Amount:   money.AsMajorUnits(),
		Currency: money.Currency().Code,
	}
}
