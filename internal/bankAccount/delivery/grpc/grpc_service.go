package grpc

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/queries"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/mappers"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/grpc_errors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/opentracing/opentracing-go/log"
	uuid "github.com/satori/go.uuid"
)

type grpcService struct {
	log logger.Logger
	cfg *config.Config
	bs  *service.BankAccountService
}

func NewGrpcService(log logger.Logger, cfg *config.Config, bs *service.BankAccountService) *grpcService {
	return &grpcService{log: log, cfg: cfg, bs: bs}
}

func (g *grpcService) CreateBankAccount(ctx context.Context, request *bankAccountService.CreateBankAccountRequest) (*bankAccountService.CreateBankAccountResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.CreateBankAccount")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))

	aggregateID := uuid.NewV4().String()
	command := commands.CreateBankAccountCommand{
		AggregateID: aggregateID,
		Email:       request.GetEmail(),
		Address:     request.GetAddress(),
		FirstName:   request.GetFirstName(),
		LastName:    request.GetLastName(),
		Balance:     0,
		Status:      request.GetStatus(),
	}

	err := g.bs.Commands.CreateBankAccount.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(CreateBankAccount.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	return &bankAccountService.CreateBankAccountResponse{Id: aggregateID}, nil
}

func (g *grpcService) DepositBalance(ctx context.Context, request *bankAccountService.DepositBalanceRequest) (*bankAccountService.DepositBalanceResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.DepositBalance")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))

	command := commands.DepositBalanceCommand{
		AggregateID: request.GetId(),
		Amount:      request.GetAmount(),
		PaymentID:   request.GetPaymentId(),
	}

	err := g.bs.Commands.DepositBalance.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(DepositBalance.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	return new(bankAccountService.DepositBalanceResponse), nil
}

func (g *grpcService) ChangeEmail(ctx context.Context, request *bankAccountService.ChangeEmailRequest) (*bankAccountService.ChangeEmailResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.ChangeEmail")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))

	command := commands.ChangeEmailCommand{AggregateID: request.GetId(), NewEmail: request.GetEmail()}

	err := g.bs.Commands.ChangeEmail.Handle(ctx, command)
	if err != nil {
		g.log.Errorf("(ChangeEmail.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	return new(bankAccountService.ChangeEmailResponse), nil
}

func (g *grpcService) GetById(ctx context.Context, request *bankAccountService.GetByIdRequest) (*bankAccountService.GetByIdResponse, error) {
	ctx, span := tracing.StartGrpcServerTracerSpan(ctx, "grpcService.GetById")
	defer span.Finish()
	span.LogFields(log.String("req", request.String()))

	query := queries.GetBankAccountByIDQuery{AggregateID: request.GetId()}
	bankAccount, err := g.bs.Queries.GetBankAccountByID.Handle(ctx, query)
	if err != nil {
		g.log.Errorf("(GetBankAccountByID.Handle) err: %v", err)
		return nil, grpc_errors.ErrResponse(tracing.TraceWithErr(span, err))
	}

	g.log.Infof("GetById bankAccount: %+v", bankAccount)

	return &bankAccountService.GetByIdResponse{BankAccount: mappers.BankAccountToProto(bankAccount)}, nil
}

func (g *grpcService) GetBankAccountByStatus(ctx context.Context, request *bankAccountService.GetBankAccountByStatusRequest) (*bankAccountService.GetBankAccountByStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}
