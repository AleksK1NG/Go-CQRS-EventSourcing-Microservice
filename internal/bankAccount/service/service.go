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

func NewBankAccountService(
	log logger.Logger,
	aggregateStore es.AggregateStore,
	mongoRepository domain.MongoRepository,
	elasticSearchRepository domain.ElasticSearchRepository,
) *BankAccountService {

	bankAccountCommands := commands.NewBankAccountCommands(
		commands.NewChangeEmailCmdHandler(log, aggregateStore),
		commands.NewDepositBalanceCmdHandler(log, aggregateStore),
		commands.NewCreateBankAccountCmdHandler(log, aggregateStore),
		commands.NewWithdrawBalanceCommandHandler(log, aggregateStore),
	)

	newBankAccountQueries := queries.NewBankAccountQueries(
		queries.NewGetBankAccountByIDQuery(log, aggregateStore, mongoRepository),
		queries.NewSearchBankAccountsQuery(log, aggregateStore, elasticSearchRepository),
	)

	return &BankAccountService{Commands: bankAccountCommands, Queries: newBankAccountQueries}
}
