package mappers

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
)

func BankAccountToMongoProjection(bankAccount *domain.BankAccount) *domain.BankAccountMongoProjection {
	return &domain.BankAccountMongoProjection{
		AggregateID: bankAccount.AggregateID,
		Email:       bankAccount.Email,
		Address:     bankAccount.Address,
		FirstName:   bankAccount.FirstName,
		LastName:    bankAccount.LastName,
		Balance: domain.Balance{
			Amount:   bankAccount.Balance.AsMajorUnits(),
			Currency: bankAccount.Balance.Currency().Code,
		},
		Status: bankAccount.Status,
	}
}

func BankAccountToProto(bankAccount *domain.BankAccount) *bankAccountService.BankAccount {
	return &bankAccountService.BankAccount{
		Id:        bankAccount.AggregateID,
		Email:     bankAccount.Email,
		Address:   bankAccount.Address,
		FirstName: bankAccount.FirstName,
		LastName:  bankAccount.LastName,
		Balance:   BalanceMoneyToGrpc(bankAccount.Balance),
		Status:    bankAccount.Status,
	}
}

func BankAccountMongoProjectionToProto(bankAccount *domain.BankAccountMongoProjection) *bankAccountService.BankAccount {
	return &bankAccountService.BankAccount{
		Id:        bankAccount.AggregateID,
		Email:     bankAccount.Email,
		Address:   bankAccount.Address,
		FirstName: bankAccount.FirstName,
		LastName:  bankAccount.LastName,
		Balance:   BalanceToGrpc(bankAccount.Balance),
		Status:    bankAccount.Status,
	}
}
