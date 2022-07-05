package app

import (
	bankAccountGrpc "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/delivery/grpc"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	bankAccountService "github.com/AleksK1NG/go-cqrs-eventsourcing/proto/bank_account"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

func (a *app) newBankAccountGrpcServer() (func() error, *grpc.Server, error) {

	l, err := net.Listen(constants.Tcp, a.cfg.GRPC.Port)
	if err != nil {
		return nil, nil, errors.Wrap(err, "net.Listen")
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: maxConnectionIdle * time.Minute,
			Timeout:           gRPCTimeout * time.Second,
			MaxConnectionAge:  maxConnectionAge * time.Minute,
			Time:              gRPCTime * time.Minute,
		}),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			a.interceptorManager.Logger,
		),
		),
	)

	bankAccountGrpcService := bankAccountGrpc.NewGrpcService(a.log, &a.cfg, a.bankAccountService, a.validate, a.metrics)
	bankAccountService.RegisterBankAccountServiceServer(grpcServer, bankAccountGrpcService)
	grpc_prometheus.Register(grpcServer)

	if a.cfg.GRPC.Development {
		reflection.Register(grpcServer)
	}

	go func() {
		a.log.Infof("(%s gRPC app is listening) on port: %s, app info: %+v", GetMicroserviceName(a.cfg), a.cfg.GRPC.Port, grpcServer.GetServiceInfo())
		a.log.Errorf("(newAssignmentGrpcServer) err: %v", grpcServer.Serve(l))
	}()

	return l.Close, grpcServer, nil
}
