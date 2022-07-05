package metrics

import (
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ESMicroserviceMetrics struct {
	SuccessGrpcRequests prometheus.Counter
	ErrorGrpcRequests   prometheus.Counter

	SuccessHttpRequests prometheus.Counter
	ErrorHttpRequests   prometheus.Counter

	SuccessKafkaMessages prometheus.Counter
	ErrorKafkaMessages   prometheus.Counter

	HttpCreateBankAccountRequests prometheus.Counter
	HttpDepositBalanceRequests    prometheus.Counter
	HttpWithdrawBalanceRequests   prometheus.Counter
	HttpChangeEmailRequests       prometheus.Counter
	HttpGetBuIdRequests           prometheus.Counter
	HttpSearchRequests            prometheus.Counter

	GrpcCreateBankAccountRequests prometheus.Counter
	GrpcDepositBalanceRequests    prometheus.Counter
	GrpcWithdrawBalanceRequests   prometheus.Counter
	GrpcChangeEmailRequests       prometheus.Counter
	GrpcGetBuIdRequests           prometheus.Counter
	GrpcSearchRequests            prometheus.Counter
}

func NewESMicroserviceMetrics(cfg config.Config) *ESMicroserviceMetrics {
	return &ESMicroserviceMetrics{
		SuccessGrpcRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_success_grpc_requests_total", cfg.ServiceName),
			Help: "The total number of success grpc requests",
		}),
		ErrorGrpcRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_error_grpc_requests_total", cfg.ServiceName),
			Help: "The total number of error grpc requests",
		}),

		SuccessHttpRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_success_http_requests_total", cfg.ServiceName),
			Help: "The total number of success http requests",
		}),
		ErrorHttpRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_error_http_requests_total", cfg.ServiceName),
			Help: "The total number of error http requests",
		}),

		HttpCreateBankAccountRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_create_bank_account_http_requests_total", cfg.ServiceName),
			Help: "The total number create_bank_account http requests",
		}),
		HttpDepositBalanceRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_deposit_balance_http_requests_total", cfg.ServiceName),
			Help: "The total number deposit_balance http requests",
		}),
		HttpWithdrawBalanceRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_withdraw_balance_http_requests_total", cfg.ServiceName),
			Help: "The total number withdraw_balance http requests",
		}),
		HttpChangeEmailRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_change_email_http_requests_total", cfg.ServiceName),
			Help: "The total number change_email http requests",
		}),
		HttpGetBuIdRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_get_by_id_http_requests_total", cfg.ServiceName),
			Help: "The total number get_by_id http requests",
		}),
		HttpSearchRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_search_http_requests_total", cfg.ServiceName),
			Help: "The total number search http requests",
		}),

		GrpcCreateBankAccountRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_create_bank_account_grpc_requests_total", cfg.ServiceName),
			Help: "The total number create_bank_account grpc requests",
		}),
		GrpcDepositBalanceRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_deposit_balance_grpc_requests_total", cfg.ServiceName),
			Help: "The total number deposit_balance grpc requests",
		}),
		GrpcWithdrawBalanceRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_withdraw_balance_grpc_requests_total", cfg.ServiceName),
			Help: "The total number withdraw_balance grpc requests",
		}),
		GrpcChangeEmailRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_change_email_grpc_requests_total", cfg.ServiceName),
			Help: "The total number change_email grpc requests",
		}),
		GrpcGetBuIdRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_get_by_id_grpc_requests_total", cfg.ServiceName),
			Help: "The total number get_by_id grpc requests",
		}),
		GrpcSearchRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("%s_search_grpc_requests_total", cfg.ServiceName),
			Help: "The total number search grpc requests",
		}),
	}
}
