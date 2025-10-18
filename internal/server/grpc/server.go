package grpc

import (
	"context"
	"net"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	pb "github.com/dmitrijs2005/gophkeeper/internal/proto"
	"github.com/dmitrijs2005/gophkeeper/internal/server/models"
	"github.com/dmitrijs2005/gophkeeper/internal/server/services"
	"google.golang.org/grpc"
)

type userSvc interface {
	RefreshToken(ctx context.Context, refresh string) (*services.TokenPair, error)
	Register(ctx context.Context, username string, salt []byte, verifier []byte) (*models.User, error)
	GetSalt(ctx context.Context, username string) ([]byte, error)
	Login(ctx context.Context, username string, verifierCandidate []byte) (*services.TokenPair, error)
}

type entrySvc interface {
	Sync(ctx context.Context, userID string, pendingEntries []*models.Entry, pendingFiles []*models.File,
		clientMaxVersion int64) (processed []*models.Entry, newEntries []*models.Entry, newFiles []*models.File, uploadTasks []*models.FileUploadTask, globalMaxVersion int64, err error)
	MarkUploaded(ctx context.Context, entryID string) error
	GetPresignedGetURL(ctx context.Context, entryID string) (string, error)
}

type GRPCServer struct {
	pb.UnimplementedGophKeeperServiceServer
	address   string
	users     userSvc
	entries   entrySvc
	logger    logging.Logger
	jwtSecret []byte
}

func NewgGRPCServer(a string, l logging.Logger, us *services.UserService, es *services.EntryService, secretKey string) (*GRPCServer, error) {
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
