package service

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/queries"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
)

type BankAccountService struct {
	Commands *commands.BankAccountCommands
	Queries  *queries.BankAccountQueries
}

func NewBankAccountService(log logger.Logger, es es.AggregateStore, mr domain.MongoRepository) *BankAccountService {

	bankAccountCommands := commands.NewBankAccountCommands(
		commands.NewChangeEmailCmdHandler(log, es),
		commands.NewDepositBalanceCmdHandler(log, es),
		commands.NewCreateBankAccountCmdHandler(log, es),
	)

	newBankAccountQueries := queries.NewBankAccountQueries(
		queries.NewGetBankAccountByIDQuery(log, es, mr),
	)

	return &BankAccountService{Commands: bankAccountCommands, Queries: newBankAccountQueries}
}
