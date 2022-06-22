package server

import (
	grpc2 "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/delivery/grpc"
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

func (s *server) newBankAccountGrpcServer() (func() error, *grpc.Server, error) {

	l, err := net.Listen(constants.Tcp, s.cfg.GRPC.Port)
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
			s.im.Logger,
		),
		),
	)

	bankAccountGrpcService := grpc2.NewGrpcService(s.log, &s.cfg, s.bs)
	bankAccountService.RegisterBankAccountServiceServer(grpcServer, bankAccountGrpcService)
	grpc_prometheus.Register(grpcServer)

	if s.cfg.GRPC.Development {
		reflection.Register(grpcServer)
	}

	go func() {
		s.log.Infof("(%s gRPC server is listening) on port: %s, server info: %+v", GetMicroserviceName(s.cfg), s.cfg.GRPC.Port, grpcServer.GetServiceInfo())
		s.log.Errorf("(newAssignmentGrpcServer) err: %v", grpcServer.Serve(l))
	}()

	return l.Close, grpcServer, nil
}
