package grpc_client

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/interceptors"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
)

func NewBankAccountGrpcClient(ctx context.Context, port string, im interceptors.InterceptorManager) (bankAccountService.BankAccountServiceClient, func() error, error) {
	grpcServiceConn, err := NewGrpcServiceConn(ctx, port, im)
	if err != nil {
		return nil, nil, err
	}

	serviceClient := bankAccountService.NewBankAccountServiceClient(grpcServiceConn)

	return serviceClient, grpcServiceConn.Close, nil
}
