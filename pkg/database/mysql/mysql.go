package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/jackc/pgx/v5"

	"golang.org/x/exp/slog"

	"github.com/Masterminds/squirrel"
)

// Mysql - Structure for Mysql instance
type Mysql struct {
	ReadPool  *sql.DB
	WritePool *sql.DB

	Builder squirrel.StatementBuilderType
	// options
	maxDataPerWrite       int
	maxRetries            int
	watchBufferSize       int
	maxConnectionLifeTime time.Duration
	maxConnectionIdleTime time.Duration
	maxOpenConnections    int
	maxIdleConnections    int
}

type poolConfig struct {
	dsn                   string
	maxDataPerWrite       int
	maxRetries            int
	watchBufferSize       int
	maxConnectionLifeTime time.Duration
	maxConnectionIdleTime time.Duration
	maxOpenConnections    int
	maxIdleConnections    int
}

// New -
func New(uri string, opts ...Option) (*Mysql, error) {
	return newDB(uri, uri, opts...)
}

// NewWithSeparateURIs -
func NewWithSeparateURIs(writerUri, readerUri string, opts ...Option) (*Mysql, error) {
	return newDB(writerUri, readerUri, opts...)
}

// new - Creates new mysql db instance
func newDB(writerUri, readerUri string, opts ...Option) (*Mysql, error) {
	var err error

	mysql := &Mysql{
		maxOpenConnections: _defaultMaxOpenConnections,
		maxIdleConnections: _defaultMaxIdleConnections,
		maxDataPerWrite:    _defaultMaxDataPerWrite,
		maxRetries:         _defaultMaxRetries,
		watchBufferSize:    _defaultWatchBufferSize,
	}

	// Custom options
	for _, opt := range opts {
		opt(mysql)
	}

	mysql.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	var readConfig, writeConfig poolConfig

	// Set uri for both pools
	readConfig.dsn = readerUri
	writeConfig.dsn = writerUri

	// Set the default execution mode for queries using the write and read configurations.
	// setDefaultQueryExecMode(writeConfig.ConnConfig)
	// setDefaultQueryExecMode(readConfig.ConnConfig)

	// Set the plan cache mode for both write and read configurations to optimize query planning.
	// setPlanCacheMode(writeConfig.ConnConfig)
	// setPlanCacheMode(readConfig.ConnConfig)

	// Set the minimum number of idle connections in the pool for both write and read configurations.
	// writeConfig.MinConns = int32(mysql.maxIdleConnections)
	// readConfig.MinConns = int32(mysql.maxIdleConnections)

	// Set the maximum number of active connections in the pool for both write and read configurations.
	// writeConfig.MaxConns = int32(mysql.maxOpenConnections)
	// readConfig.MaxConns = int32(mysql.maxOpenConnections)
	//
	// // Set the maximum amount of time a connection may be idle before being closed for both configurations.
	// writeConfig.MaxConnIdleTime = mysql.maxConnectionIdleTime
	// readConfig.MaxConnIdleTime = mysql.maxConnectionIdleTime
	//
	// // Set the maximum lifetime of a connection in the pool for both configurations.
	// writeConfig.MaxConnLifetime = mysql.maxConnectionLifeTime
	// readConfig.MaxConnLifetime = mysql.maxConnectionLifeTime
	//
	// // Set a jitter to the maximum connection lifetime to prevent all connections from expiring at the same time.
	// writeConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(mysql.maxConnectionLifeTime))
	// readConfig.MaxConnLifetimeJitter = time.Duration(0.2 * float64(mysql.maxConnectionLifeTime))
	//
	// writeConfig.ConnConfig.Tracer = otelpgx.NewTracer()
	// readConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	// Create connection pools for both writing and reading operations using the configured settings.
	mysql.WritePool, mysql.ReadPool, err = createPools(
		context.Background(), // Context used to control the lifecycle of the pools.
		&writeConfig,         // Configuration settings for the write pool.
		&readConfig,          // Configuration settings for the read pool.
	)
	// Handle errors during the creation of the connection pools.
	if err != nil {
		return nil, err
	}

	return mysql, nil
}

func (m *Mysql) GetMaxDataPerWrite() int {
	return m.maxDataPerWrite
}

func (m *Mysql) GetMaxRetries() int {
	return m.maxRetries
}

func (m *Mysql) GetWatchBufferSize() int {
	return m.watchBufferSize
}

// GetEngineType - Get the engine type which is mysql in string
func (m *Mysql) GetEngineType() string {
	return "mysql"
}

// Close - Close mysql instance
func (m *Mysql) Close() error {
	m.ReadPool.Close()
	m.WritePool.Close()
	return nil
}

// IsReady - Check if database is ready
func (m *Mysql) IsReady(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := m.ReadPool.Ping(); err != nil {
		return false, err
	}
	return true, nil
}

var queryExecModes = map[string]pgx.QueryExecMode{
	"cache_statement": pgx.QueryExecModeCacheStatement,
	"cache_describe":  pgx.QueryExecModeCacheDescribe,
	"describe_exec":   pgx.QueryExecModeDescribeExec,
	"mode_exec":       pgx.QueryExecModeExec,
	"simple_protocol": pgx.QueryExecModeSimpleProtocol,
}

func setDefaultQueryExecMode(config *pgx.ConnConfig) {
	// Default mode if no specific mode is found in the connection string
	defaultMode := "cache_statement"

	// Iterate through the map keys to check if any are mentioned in the connection string
	for key := range queryExecModes {
		if strings.Contains(config.ConnString(), "default_query_exec_mode="+key) {
			config.DefaultQueryExecMode = queryExecModes[key]
			slog.Info("setDefaultQueryExecMode", slog.String("mode", key))
			return
		}
	}

	// Set to default mode if no matching mode is found
	config.DefaultQueryExecMode = queryExecModes[defaultMode]
	slog.Warn("setDefaultQueryExecMode", slog.String("mode", defaultMode))
}

var planCacheModes = map[string]string{
	"auto":              "auto",
	"force_custom_plan": "force_custom_plan",
	"disable":           "disable",
}

func setPlanCacheMode(config *pgx.ConnConfig) {
	// Default plan cache mode
	const defaultMode = "auto"

	// Extract connection string
	connStr := config.ConnString()
	planCacheMode := defaultMode

	// Check for specific plan cache modes in the connection string
	for key, value := range planCacheModes {
		if strings.Contains(connStr, "plan_cache_mode="+key) {
			if key == "disable" {
				delete(config.Config.RuntimeParams, "plan_cache_mode")
				slog.Info("setPlanCacheMode", slog.String("mode", "disabled"))
				return
			}
			planCacheMode = value
			slog.Info("setPlanCacheMode", slog.String("mode", key))
			break
		}
	}

	// Set the plan cache mode
	config.Config.RuntimeParams["plan_cache_mode"] = planCacheMode
	if planCacheMode == defaultMode {
		slog.Warn("setPlanCacheMode", slog.String("mode", defaultMode))
	}
}

// createPools initializes read and write connection pools with appropriate configurations and error handling.
func createPools(ctx context.Context, wConfig, rConfig *poolConfig) (*sql.DB, *sql.DB, error) {
	// Context with timeout for creating the pools
	_, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create write pool and configuaring it
	writePool, err := sql.Open("mysql", wConfig.dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create write pool: %w", err)
	}
	writePool.SetMaxOpenConns(wConfig.maxOpenConnections)
	writePool.SetMaxIdleConns(wConfig.maxIdleConnections)
	writePool.SetConnMaxIdleTime(wConfig.maxConnectionIdleTime)
	writePool.SetConnMaxLifetime(wConfig.maxConnectionLifeTime)

	// Create read pool and configuaring it
	readPool, err := sql.Open("mysql", rConfig.dsn)
	if err != nil {
		writePool.Close() // Ensure write pool is closed on failure
		return nil, nil, fmt.Errorf("failed to create read pool: %w", err)
	}
	readPool.SetMaxOpenConns(rConfig.maxOpenConnections)
	readPool.SetMaxIdleConns(rConfig.maxIdleConnections)
	readPool.SetConnMaxIdleTime(rConfig.maxConnectionIdleTime)
	readPool.SetConnMaxLifetime(rConfig.maxConnectionLifeTime)

	// Set up retry policy for pinging pools
	retryPolicy := backoff.NewExponentialBackOff()
	retryPolicy.MaxElapsedTime = 1 * time.Minute

	// Attempt to ping both pools to confirm connectivity
	err = backoff.Retry(func() error {
		if err := writePool.Ping(); err != nil {
			return fmt.Errorf("write pool ping failed: %w", err)
		}
		if err := readPool.Ping(); err != nil {
			return fmt.Errorf("read pool ping failed: %w", err)
		}
		return nil
	}, retryPolicy)

	// Handle errors from pinging
	if err != nil {
		writePool.Close()
		readPool.Close()
		return nil, nil, fmt.Errorf("pinging pools failed: %w", err)
	}

	return writePool, readPool, nil
}
