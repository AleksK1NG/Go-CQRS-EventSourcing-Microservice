package grpc

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/grpc_client"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/interceptors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func SetupTestBankAccountGrpcClient() (bankAccountService.BankAccountServiceClient, func() error, error) {
	appLogger := logger.NewAppLogger(logger.LogConfig{
		LogLevel: "debug",
		DevMode:  false,
		Encoder:  "console",
	})
	interceptorManager := interceptors.NewInterceptorManager(appLogger, nil)
	return grpc_client.NewBankAccountGrpcClient(context.Background(), ":5001", interceptorManager)
}

func TestGrpcService_CreateBankAccount(t *testing.T) {
	client, closeFunc, err := SetupTestBankAccountGrpcClient()
	require.NoError(t, err)
	defer closeFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	response, err := client.CreateBankAccount(ctx, &bankAccountService.CreateBankAccountRequest{
		Email:     "alexander.bryksin@yandex.ru",
		FirstName: "Alexander",
		LastName:  "Bryksin",
		Address:   "Moscow",
		Status:    "PREMIUM",
		Balance:   5555555555,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEqual(t, response.Id, "")

	t.Logf("response: %s", response.String())
}
