package grpc

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/queries"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/mappers"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/metrics"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/grpc_errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/go-playground/validator"
	"github.com/opentracing/opentracing-go/log"
	uuid "github.com/satori/go.uuid"
)

type grpcService struct {
	log                logger.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	validate           *validator.Validate
	metrics            *metrics.ESMicroserviceMetrics
}

func NewGrpcService(
	log logger.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	validate *validator.Validate,
	metrics *metrics.ESMicroserviceMetrics,
) *grpcService {
	return &grpcService{log: log, cfg: cfg, bankAccountService: bankAccountService, validate: validate, metrics: metrics}
}

func (g *grpcService) CreateBankAccount(ctx context.Context, request *bankAccountService.CreateBankAccountRequest) (*bankAccountService.CreateBankAccountResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.CreateBankAccount")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))
	g.metrics.GrpcCreateBankAccountRequests.Inc()

	aggregateID := uuid.NewV4().String()
	command := commands.CreateBankAccountCommand{
		AggregateID: aggregateID,
		Email:       request.GetEmail(),
		Address:     request.GetAddress(),
		FirstName:   request.GetFirstName(),
		LastName:    request.GetLastName(),
		Balance:     request.GetBalance(),
		Status:      request.GetStatus(),
	}

	if err := g.validate.StructCtx(ctx, command); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	err := g.bankAccountService.Commands.CreateBankAccount.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(CreateBankAccount.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [created account] aggregateID: %s", aggregateID)
	return &bankAccountService.CreateBankAccountResponse{Id: aggregateID}, nil
}

func (g *grpcService) DepositBalance(ctx context.Context, request *bankAccountService.DepositBalanceRequest) (*bankAccountService.DepositBalanceResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.DepositBalance")
	defer span.Finish()
	span.LogFields(log.String("request", request.String()))
	g.metrics.GrpcDepositBalanceRequests.Inc()

	command := commands.DepositBalanceCommand{
		AggregateID: request.GetId(),
		Amount:      request.GetAmount(),
		PaymentID:   request.GetPaymentId(),
	}

	if err := g.validate.StructCtx(ctx, command); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	err := g.bankAccountService.Commands.DepositBalance.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(DepositBalance.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [deposited balance] aggregateID: %s, amount: %v", request.GetId(), request.GetAmount())
	return new(bankAccountService.DepositBalanceResponse), nil
}

func (g *grpcService) WithdrawBalance(ctx context.Context, request *bankAccountService.WithdrawBalanceRequest) (*bankAccountService.WithdrawBalanceResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.WithdrawBalance")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))
	g.metrics.GrpcWithdrawBalanceRequests.Inc()

	command := commands.WithdrawBalanceCommand{
		AggregateID: request.GetId(),
		Amount:      request.GetAmount(),
		PaymentID:   request.GetPaymentId(),
	}

	if err := g.validate.StructCtx(ctx, command); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	err := g.bankAccountService.Commands.WithdrawBalance.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(WithdrawBalance.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [withdraw balance] aggregateID: %s, amount: %v", request.GetId(), request.GetAmount())
	return new(bankAccountService.WithdrawBalanceResponse), nil
}

func (g *grpcService) ChangeEmail(ctx context.Context, request *bankAccountService.ChangeEmailRequest) (*bankAccountService.ChangeEmailResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.ChangeEmail")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))
	g.metrics.GrpcChangeEmailRequests.Inc()

	command := commands.ChangeEmailCommand{AggregateID: request.GetId(), NewEmail: request.GetEmail()}

	if err := g.validate.StructCtx(ctx, command); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	err := g.bankAccountService.Commands.ChangeEmail.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(ChangeEmail.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [changed email] aggregateID: %s, newEmail: %s", request.GetId(), request.GetEmail())
	return new(bankAccountService.ChangeEmailResponse), nil
}

func (g *grpcService) GetById(ctx context.Context, request *bankAccountService.GetByIdRequest) (*bankAccountService.GetByIdResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.GetById")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))
	g.metrics.GrpcGetBuIdRequests.Inc()

	query := queries.GetBankAccountByIDQuery{AggregateID: request.GetId(), FromEventStore: request.IsOwner}

	if err := g.validate.StructCtx(ctx, query); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	bankAccountProjection, err := g.bankAccountService.Queries.GetBankAccountByID.Handle(ctx, query)
	if err != nil {
		g.log.Errorf("(GetBankAccountByID.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [get account by id] projection: %+v", bankAccountProjection)
	return &bankAccountService.GetByIdResponse{BankAccount: mappers.BankAccountMongoProjectionToProto(bankAccountProjection)}, nil
}

func (g *grpcService) SearchBankAccounts(ctx context.Context, request *bankAccountService.SearchBankAccountsRequest) (*bankAccountService.SearchBankAccountsResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.SearchBankAccounts")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))
	g.metrics.GrpcSearchRequests.Inc()

	query := queries.SearchBankAccountsQuery{
		QueryTerm: request.GetSearchText(),
		Pagination: &utils.Pagination{
			Size: int(request.GetSize()),
			Page: int(request.GetPage()),
		},
	}

	if err := g.validate.StructCtx(ctx, query); err != nil {
		g.log.Errorf("validation err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	searchQueryResult, err := g.bankAccountService.Queries.SearchBankAccounts.Handle(ctx, query)
	if err != nil {
		g.log.Errorf("(SearchBankAccounts.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("(grpcService) [search] result: %+vv", searchQueryResult.PaginationResponse)
	return &bankAccountService.SearchBankAccountsResponse{
		BankAccounts: mappers.SearchBankAccountsListToProto(searchQueryResult.List),
		Pagination:   mappers.PaginationResponseToProto(searchQueryResult.PaginationResponse),
	}, nil
}
