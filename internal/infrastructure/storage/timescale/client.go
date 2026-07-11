package timescale

import (
	"context"
	"fmt"
	"os"
	"time"

	"coin-radar/internal/config"
	"coin-radar/internal/infrastructure/storage/prisma"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPrismaClient initializes Prisma client connection for SaaS/Relational database operations.
func NewPrismaClient(cfg *config.Config) (*prisma.PrismaClient, error) {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Prisma client Go reads the database URL from the DATABASE_URL environment variable.
	if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
		return nil, fmt.Errorf("failed to set DATABASE_URL env: %w", err)
	}

	client := prisma.NewClient()
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect prisma client: %w", err)
	}

	return client, nil
}

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
