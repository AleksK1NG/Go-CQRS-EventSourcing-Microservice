package grpc

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/grpc_client"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/interceptors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/brianvoe/gofakeit/v6"
	uuid "github.com/satori/go.uuid"
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
	t.Parallel()

	client, closeFunc, err := SetupTestBankAccountGrpcClient()
	require.NoError(t, err)
	defer closeFunc() // nolint: errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request := &bankAccountService.CreateBankAccountRequest{
		Email:     gofakeit.Email(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Address:   gofakeit.Address().Address,
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 555555555)),
	}
	t.Logf("request: %s", request.String())

	response, err := client.CreateBankAccount(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEqual(t, response.Id, "")

	t.Logf("response: %s", response.String())

	getByIdResponse, err := client.GetById(ctx, &bankAccountService.GetByIdRequest{
		Id:      response.GetId(),
		IsOwner: true,
	})

	require.NoError(t, err)
	require.NotNil(t, getByIdResponse)
	require.NotNil(t, getByIdResponse.GetBankAccount())
	require.Equal(t, request.Email, getByIdResponse.GetBankAccount().GetEmail())
	require.Equal(t, request.FirstName, getByIdResponse.GetBankAccount().GetFirstName())
	require.Equal(t, request.LastName, getByIdResponse.GetBankAccount().GetLastName())
	require.Equal(t, request.Address, getByIdResponse.GetBankAccount().GetAddress())

	t.Logf("getByIdResponse: %s", getByIdResponse.String())
}

func TestGrpcService_DepositBalance(t *testing.T) {
	t.Parallel()

	client, closeFunc, err := SetupTestBankAccountGrpcClient()
	require.NoError(t, err)
	defer closeFunc() // nolint: errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request := &bankAccountService.CreateBankAccountRequest{
		Email:     gofakeit.Email(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Address:   gofakeit.Address().Address,
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 555555555)),
	}
	t.Logf("request: %s", request.String())

	response, err := client.CreateBankAccount(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEqual(t, response.Id, "")

	t.Logf("response: %s", response.String())

	_, err = client.DepositBalance(ctx, &bankAccountService.DepositBalanceRequest{
		Id:        response.GetId(),
		PaymentId: uuid.NewV4().String(),
		Amount:    1000000,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestGrpcService_WithdrawBalance(t *testing.T) {
	t.Parallel()

	client, closeFunc, err := SetupTestBankAccountGrpcClient()
	require.NoError(t, err)
	defer closeFunc() // nolint: errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request := &bankAccountService.CreateBankAccountRequest{
		Email:     gofakeit.Email(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Address:   gofakeit.Address().Address,
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 5000000)),
	}
	t.Logf("request: %s", request.String())

	response, err := client.CreateBankAccount(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEqual(t, response.Id, "")

	t.Logf("response: %s", response.String())

	_, err = client.WithdrawBalance(ctx, &bankAccountService.WithdrawBalanceRequest{
		Id:        response.GetId(),
		PaymentId: uuid.NewV4().String(),
		Amount:    1000000,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestGrpcService_ChangeEmail(t *testing.T) {
	t.Parallel()

	client, closeFunc, err := SetupTestBankAccountGrpcClient()
	require.NoError(t, err)
	defer closeFunc() // nolint: errcheck

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	request := &bankAccountService.CreateBankAccountRequest{
		Email:     gofakeit.Email(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Address:   gofakeit.Address().Address,
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 5000000)),
	}
	t.Logf("request: %s", request.String())

	response, err := client.CreateBankAccount(ctx, request)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEqual(t, response.Id, "")

	changeEmailRequest := &bankAccountService.ChangeEmailRequest{
		Id:    response.GetId(),
		Email: gofakeit.Email(),
	}

	t.Logf("response: %s", response.String())

	_, err = client.ChangeEmail(ctx, changeEmailRequest)
	require.NoError(t, err)
	require.NotNil(t, response)

	getByIdResponse, err := client.GetById(ctx, &bankAccountService.GetByIdRequest{
		Id:      response.GetId(),
		IsOwner: true,
	})

	require.NoError(t, err)
	require.NotNil(t, getByIdResponse)
	require.NotNil(t, getByIdResponse.GetBankAccount())
	require.Equal(t, changeEmailRequest.GetEmail(), getByIdResponse.GetBankAccount().GetEmail())
}
