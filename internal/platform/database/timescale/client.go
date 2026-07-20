package timescale

import (
	"context"
	"fmt"
	"time"

	"coin-radar-gin/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool initializes pgx v5 connection pool for high-throughput TimescaleDB market data operations.
func NewPgxPool(cfg *config.Config) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Configure pool parameters
	poolConfig.MaxConns = 50
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
