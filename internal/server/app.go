// Package server wires and runs the GophKeeper backend service.
//
// Responsibilities:
//   - Open and ping the database (via DSN) and run schema migrations.
//   - Construct repository manager and domain services.
//   - Start the public gRPC server and handle graceful shutdown on OS signals.
package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrijs2005/gophkeeper/internal/logging"
	"github.com/dmitrijs2005/gophkeeper/internal/server/config"
	"github.com/dmitrijs2005/gophkeeper/internal/server/repositories/repomanager"
	"github.com/dmitrijs2005/gophkeeper/internal/server/services"

	gs "github.com/dmitrijs2005/gophkeeper/internal/server/grpc"
)

// App is the composition root of the server process.
// It owns configuration, logging, and the domain services.
type App struct {
	config       *config.Config
	logger       logging.Logger
	userService  *services.UserService
	entryService *services.EntryService
}

// NewAppFromDSN opens a DB connection using cfg.DatabaseDSN, verifies it with a
// short ping, and then delegates to NewApp for full initialization.
func NewAppFromDSN(cfg *config.Config, logger logging.Logger) (*App, error) {
	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return NewApp(db, cfg, logger)
}

// NewApp constructs repositories, runs migrations, and builds domain services.
func NewApp(db *sql.DB, c *config.Config, l logging.Logger) (*App, error) {
	m, err := repomanager.NewPostgresRepositoryManager(db)
	if err != nil {
		return nil, fmt.Errorf("db init error: %w", err)
	}
	if err := m.RunMigrations(context.Background(), db); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}
	us := services.NewUserService(db, m, c)
	es := services.NewEntryService(db, m, c)
	return &App{config: c, logger: l, userService: us, entryService: es}, nil
}

// initSignalHandler installs SIGINT/SIGTERM/SIGQUIT handlers that cancel ctx.
func (app *App) initSignalHandler(cancelFunc context.CancelFunc) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigs
		cancelFunc()
	}()
}

// startGRPCServer builds and runs the gRPC server; on error it logs and cancels ctx.
func (app *App) startGRPCServer(ctx context.Context, cancelFunc context.CancelFunc) {
	s, err := gs.NewgGRPCServer(app.config.EndpointAddrGRPC, app.logger, app.userService, app.entryService, app.config.SecretKey)
	if err != nil {
		app.logger.Error(ctx, err.Error())
		cancelFunc()
		return
	}
	if err := s.Run(ctx); err != nil {
		app.logger.Error(ctx, err.Error())
		cancelFunc()
	}
}

// Run initializes context/cancellation, installs signal handling, and starts
// the gRPC server. The call blocks until the server goroutine returns.
func (app *App) Run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	app.logger.Info(ctx, "Starting app...")
	app.initSignalHandler(cancelFunc)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		app.startGRPCServer(ctx, cancelFunc)
	}()
	wg.Wait()
}
