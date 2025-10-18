// Package grpc wires the domain services to a public gRPC endpoint.
// It defines the server, its dependencies, and bootstraps the listener
// with required interceptors.
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

type ctxKey string

const UserIDKey ctxKey = "userID"

// userSvc is the subset of user service methods required by the transport.
type userSvc interface {
	RefreshToken(ctx context.Context, refresh string) (*services.TokenPair, error)
	Register(ctx context.Context, username string, salt []byte, verifier []byte) (*models.User, error)
	GetSalt(ctx context.Context, username string) ([]byte, error)
	Login(ctx context.Context, username string, verifierCandidate []byte) (*services.TokenPair, error)
}

// entrySvc is the subset of entry service methods required by the transport.
type entrySvc interface {
	// Sync reconciles client changes and returns merged updates plus upload tasks.
	Sync(ctx context.Context, userID string, pendingEntries []*models.Entry, pendingFiles []*models.File,
		clientMaxVersion int64) (processed []*models.Entry, newEntries []*models.Entry, newFiles []*models.File, uploadTasks []*models.FileUploadTask, globalMaxVersion int64, err error)
	// MarkUploaded acknowledges completion of a client-side file upload.
	MarkUploaded(ctx context.Context, entryID string) error
	// GetPresignedGetURL returns a temporary URL to fetch an encrypted file.
	GetPresignedGetURL(ctx context.Context, entryID string) (string, error)
}

// GRPCServer hosts the GophKeeper gRPC API and delegates to domain services.
type GRPCServer struct {
	pb.UnimplementedGophKeeperServiceServer
	address   string
	users     userSvc
	entries   entrySvc
	logger    logging.Logger
	jwtSecret []byte
}

// NewgGRPCServer constructs a GRPCServer bound to the given address, logger,
// services, and JWT secret. The name is preserved for compatibility.
func NewgGRPCServer(a string, l logging.Logger, us *services.UserService, es *services.EntryService, secretKey string) (*GRPCServer, error) {
	return &GRPCServer{
		address:   a,
		logger:    l.With("module", "grpc_server"),
		users:     us,
		entries:   es,
		jwtSecret: []byte(secretKey),
	}, nil
}

// Run starts the gRPC server on the configured address and blocks until the
// context is canceled, at which point it performs a graceful stop.
func (s *GRPCServer) Run(ctx context.Context) error {
	// Announce address.
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// Create gRPC server with interceptors (auth, etc.).
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(s.accessTokenInterceptor))

	// Register service implementation.
	pb.RegisterGophKeeperServiceServer(srv, s)

	// Graceful shutdown on context cancellation.
	go func() {
		<-ctx.Done()
		s.logger.Info(ctx, "Stopping gPRC server...")
		srv.GracefulStop()
	}()

	s.logger.Info(ctx, "Starting gRPC server", "address", s.address)

	// Serve incoming connections.
	if err := srv.Serve(listen); err != nil {
		return err
	}
	return nil
}
