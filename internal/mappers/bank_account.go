package mappers

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BankAccountToMongoProjection(bankAccount *domain.BankAccountAggregate) *domain.BankAccountMongoProjection {
	return &domain.BankAccountMongoProjection{
		AggregateID: bankAccount.BankAccount.AggregateID,
		Version:     bankAccount.Version,
		Email:       bankAccount.BankAccount.Email,
		Address:     bankAccount.BankAccount.Address,
		FirstName:   bankAccount.BankAccount.FirstName,
		LastName:    bankAccount.BankAccount.LastName,
		Balance: domain.Balance{
			Amount:   bankAccount.BankAccount.Balance.AsMajorUnits(),
			Currency: bankAccount.BankAccount.Balance.Currency().Code,
		},
		Status: bankAccount.BankAccount.Status,
	}
}

func BankAccountToElasticProjection(bankAccount *domain.BankAccountAggregate) *domain.ElasticSearchProjection {
	return &domain.ElasticSearchProjection{
		ID:          bankAccount.BankAccount.AggregateID,
		AggregateID: bankAccount.BankAccount.AggregateID,
		Version:     bankAccount.Version,
		Email:       bankAccount.BankAccount.Email,
		Address:     bankAccount.BankAccount.Address,
		FirstName:   bankAccount.BankAccount.FirstName,
		LastName:    bankAccount.BankAccount.LastName,
		Balance: domain.Balance{
			Amount:   bankAccount.BankAccount.Balance.AsMajorUnits(),
			Currency: bankAccount.BankAccount.Balance.Currency().Code,
		},
		Status: bankAccount.BankAccount.Status,
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
		UpdatedAt: timestamppb.New(bankAccount.UpdatedAt.UTC()),
		CreatedAt: timestamppb.New(bankAccount.CreatedAt.UTC()),
	}
}

func BankAccountElasticProjectionToProto(bankAccount *domain.ElasticSearchProjection) *bankAccountService.BankAccount {
	return &bankAccountService.BankAccount{
		Id:        bankAccount.AggregateID,
		Email:     bankAccount.Email,
		Address:   bankAccount.Address,
		FirstName: bankAccount.FirstName,
		LastName:  bankAccount.LastName,
		Balance:   BalanceToGrpc(bankAccount.Balance),
		Status:    bankAccount.Status,
		UpdatedAt: timestamppb.New(bankAccount.UpdatedAt.UTC()),
		CreatedAt: timestamppb.New(bankAccount.CreatedAt.UTC()),
	}
}

func SearchBankAccountsListToProto(list []*domain.ElasticSearchProjection) []*bankAccountService.BankAccount {
	result := make([]*bankAccountService.BankAccount, 0, len(list))
	for _, projection := range list {
		result = append(result, BankAccountElasticProjectionToProto(projection))
	}
	return result
}
