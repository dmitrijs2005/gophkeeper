package grpc

import (
	"context"
	"net"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/dmitrijs2005/gophkeeper/internal/server/entries"
	"github.com/dmitrijs2005/gophkeeper/internal/server/users"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	pb.UnimplementedGophKeeperServiceServer
	address   string
	users     *users.Service
	entries   *entries.Service
	logger    logging.Logger
	jwtSecret []byte
}

func NewgGRPCServer(a string, l logging.Logger, us *users.Service, es *entries.Service, secretKey string) (*GRPCServer, error) {
	return &GRPCServer{
		address:   a,
		logger:    l.With("module", "grpc_server"),
		users:     us,
		entries:   es,
		jwtSecret: []byte(secretKey),
	}, nil
}

func (s *GRPCServer) Run(ctx context.Context) error {

	// announces address
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// creates gRPC-server
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(s.accessTokenInterceptor))

	// registers service
	pb.RegisterGophKeeperServiceServer(srv, s)

	go func() {
		<-ctx.Done()
		s.logger.Info(ctx, "Stopping gPRC server...")
		srv.GracefulStop()
	}()

	s.logger.Info(ctx, "Starting gRPC server", "address", s.address)

	// starts accepting incoming connections
	if err := srv.Serve(listen); err != nil {
		return err
	}

	return nil
}
