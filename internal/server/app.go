// Package server initializes and runs the main application server.
// It configures storage backends, handles graceful shutdown, restores and saves metric dumps,
// and starts the HTTP server for metric collection.
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

type App struct {
	config       *config.Config
	logger       logging.Logger
	userService  *services.UserService
	entryService *services.EntryService
}

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

func NewApp(db *sql.DB, c *config.Config, l logging.Logger) (*App, error) {

	//config := config.LoadConfig()

	m, err := repomanager.NewPostgresRepositoryManager(db)
	if err != nil {
		return nil, fmt.Errorf("db init error: %w", err)
	}

	err = m.RunMigrations(context.Background(), db)
	if err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}

	us := services.NewUserService(db, m, c)
	es := services.NewEntryService(db, m, c)

	return &App{config: c, logger: l, userService: us, entryService: es}, nil
}

func (app *App) initSignalHandler(cancelFunc context.CancelFunc) {
	// Channel to catch OS signals.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sigs
		cancelFunc()
	}()
}

func (app *App) startGRPCServer(ctx context.Context, cancelFunc context.CancelFunc) {

	s, err := gs.NewgGRPCServer(app.config.EndpointAddrGRPC, app.logger, app.userService, app.entryService, app.config.SecretKey)

	if err != nil {
		app.logger.Error(ctx, err.Error())
		cancelFunc()
	} else {

		if err := s.Run(ctx); err != nil {
			app.logger.Error(ctx, err.Error())
			cancelFunc()
		}
	}
}

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

// func (app *App) closeDBIfNeeded(s storage.Storage) (bool, error) {

// 	db, ok := s.(storage.DBStorage)
// 	if ok {
// 		err := db.Close()
// 		return true, err
// 	}

// 	return false, nil

// }
