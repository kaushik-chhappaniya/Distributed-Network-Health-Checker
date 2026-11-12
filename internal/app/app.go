package app

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"kaushik-chhappaniya/distributed-network-health-checker/internal/db"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/discovery"
	"kaushik-chhappaniya/distributed-network-health-checker/internal/ping"
	logs "kaushik-chhappaniya/distributed-network-health-checker/utils/logger"
)

type App struct {
	Logger           *log.Logger
	DiscoveryService *discovery.Service
	DBService        *db.Service
	DBPool           *pgxpool.Pool
	PingService      *ping.Service
	cleanupFuncs     []func()
}

func New() (*App, error) {
	app := &App{}

	// Helper to run cleanups if initialization fails midway
	cleanupOnError := func() {
		for i := len(app.cleanupFuncs) - 1; i >= 0; i-- {
			app.cleanupFuncs[i]()
		}
	}

	// ------------------- 1. Create All Dependency Injection Services -------------------
	// ---------------- App Logger ---------------------
	appLog, logComplete, err := logs.New("App.log", true)
	if err != nil {
		return nil, fmt.Errorf("failed to init app logger: %w", err)
	}
	app.cleanupFuncs = append(app.cleanupFuncs, logComplete)
	app.Logger = appLog

	// ---------------- Discovery Logger ---------------------
	discoveryLog, discLogComplete, discoveryErr := logs.New("Discovery.log", true)
	if discoveryErr != nil {
		appLog.Printf("<ERROR> Failed to start discovery log instance: %v", discoveryErr)
		cleanupOnError()
		return nil, fmt.Errorf("failed to init discovery logger: %w", discoveryErr)
	}
	app.cleanupFuncs = append(app.cleanupFuncs, discLogComplete)
	app.DiscoveryService = discovery.New(discoveryLog)

	// ---------------- Database Logger ---------------------
	dbLog, dbLogComplete, dbErr := logs.New("Db.log", true)
	if dbErr != nil {
		appLog.Printf("<ERROR> Database logger instance starting failed: %v", dbErr)
		cleanupOnError()
		return nil, fmt.Errorf("failed to init db logger: %w", dbErr)
	}
	app.cleanupFuncs = append(app.cleanupFuncs, dbLogComplete)

	// ---------------- Database Connection ---------------------
	pool, err := db.Connect(appLog)
	if err != nil {
		appLog.Printf("<ERROR> Error while connecting to database: %v", err)
		cleanupOnError()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	app.DBPool = pool
	app.cleanupFuncs = append(app.cleanupFuncs, pool.Close)
	app.DBService = db.New(pool, dbLog)

	// ---------------- Database Existence ---------------------
	// Ensure DB schema exists
	if _, err := app.DBService.CreateModels(); err != nil {
		appLog.Printf("<WARN> Failed to verify schema: %v", err)
	}

	// ---------------- Ping Logger ---------------------
	pingLog, pingLogComplete, pingErr := logs.New("Ping.log", true)
	if pingErr != nil {
		appLog.Printf("<ERROR> Ping logger instance starting failed: %v", pingErr)
		cleanupOnError()
		return nil, fmt.Errorf("failed to init ping logger: %w", pingErr)
	}
	app.cleanupFuncs = append(app.cleanupFuncs, pingLogComplete)
	app.PingService = ping.New(app.DBService, 5, pingLog)

	appLog.Println("<INFO> Application dependencies initialized successfully.")
	return app, nil
}

// Close executes all registered cleanup functions in LIFO (reverse) order.
func (a *App) Close() {
	// Close all resources in reverse order of initialization
	for i := len(a.cleanupFuncs) - 1; i >= 0; i-- {
		a.cleanupFuncs[i]()
	}
}
