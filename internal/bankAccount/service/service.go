package service

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
)

type BankAccountService struct {
	Commands *commands.BankAccountCommands
}

func NewBankAccountService(log logger.Logger, es es.AggregateStore) *BankAccountService {

	bankAccountCommands := commands.NewBankAccountCommands(
		commands.NewChangeEmailCmdHandler(log, es),
		commands.NewDepositBalanceCmdHandler(log, es),
		commands.NewCreateBankAccountCmdHandler(log, es),
	)

	return &BankAccountService{Commands: bankAccountCommands}
}
