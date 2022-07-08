package v1

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/http_client"
	"github.com/brianvoe/gofakeit/v6"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestBankAccountHandlers_CreateBankAccount(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())
}

func TestBankAccountHandlers_DepositBalance(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(0, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())

	depositBalanceCommand := commands.DepositBalanceCommand{
		AggregateID: id,
		Amount:      10000000,
		PaymentID:   uuid.NewV4().String(),
	}

	response, err = client.R().
		SetContext(ctx).
		SetBody(depositBalanceCommand).
		SetPathParam("id", id).
		Put("http://localhost:5007/api/v1/accounts/deposit/{id}")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusOK)
}

func TestBankAccountHandlers_WithdrawBalance(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(30000000, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())

	withdrawBalanceCommand := commands.WithdrawBalanceCommand{
		AggregateID: id,
		Amount:      100000,
		PaymentID:   uuid.NewV4().String(),
	}

	response, err = client.R().
		SetContext(ctx).
		SetBody(withdrawBalanceCommand).
		SetPathParam("id", id).
		Put("http://localhost:5007/api/v1/accounts/withdraw/{id}")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusOK)
}

func TestBankAccountHandlers_ChangeEmail(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(30000000, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())

	changeEmailCommand := commands.ChangeEmailCommand{
		AggregateID: id,
		NewEmail:    gofakeit.Email(),
	}

	response, err = client.R().
		SetContext(ctx).
		SetBody(changeEmailCommand).
		SetPathParam("id", id).
		Put("http://localhost:5007/api/v1/accounts/email/{id}")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusOK)
}

func TestBankAccountHandlers_GetByID(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(30000000, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())

	response, err = client.R().
		SetContext(ctx).
		SetPathParam("id", id).
		Get("http://localhost:5007/api/v1/accounts/{id}")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusOK)
	t.Logf("response: %s", response.String())
}

func TestBankAccountHandlers_Search(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := commands.CreateBankAccountCommand{
		Email:     gofakeit.Email(),
		Address:   gofakeit.Address().Address,
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Status:    "PREMUIM",
		Balance:   int64(gofakeit.Number(30000000, 5000000)),
	}

	response, err := client.R().
		SetContext(ctx).
		SetBody(command).
		Post("http://localhost:5007/api/v1/accounts")
	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var id string
	err = json.Unmarshal(response.Body(), &id)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	t.Logf("response: %s", response.String())

	response, err = client.R().
		SetContext(ctx).
		SetQueryParam("search", "a").
		SetQueryParam("page", "1").
		SetQueryParam("size", "10").
		Get("http://localhost:5007/api/v1/accounts/search")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusOK)
	t.Logf("response: %s", response.String())
}
