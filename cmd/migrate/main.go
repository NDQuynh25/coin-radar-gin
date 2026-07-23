// Command migrate applies the versioned SQL files in migrations/.
// Run from the repository root: go run ./cmd/migrate up
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"coin-radar-gin/config"

	"github.com/jackc/pgx/v5"
)

type migration struct {
	version int
	path    string
}

func main() {
	if len(os.Args) != 2 || os.Args[1] != "up" {
		fatal("usage: go run ./cmd/migrate up")
	}

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		fatal("load configuration: %v", err)
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.Name, cfg.Database.SSLMode)
	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		fatal("parse database URL: %v", err)
	}
	// Migrations contain multiple SQL statements per file.
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	ctx := context.Background()
	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		fatal("connect database: %v", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version BIGINT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		fatal("create migration tracking table: %v", err)
	}
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(4815162342)"); err != nil {
		fatal("lock migrations: %v", err)
	}
	defer conn.Exec(ctx, "SELECT pg_advisory_unlock(4815162342)")

	migrations, err := discover("migrations")
	if err != nil {
		fatal("discover migrations: %v", err)
	}
	for _, migration := range migrations {
		var applied bool
		if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", migration.version).Scan(&applied); err != nil {
			fatal("check migration %d: %v", migration.version, err)
		}
		if applied {
			continue
		}
		contents, err := os.ReadFile(migration.path)
		if err != nil {
			fatal("read %s: %v", migration.path, err)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			fatal("begin migration %d: %v", migration.version, err)
		}
		if _, err = tx.Exec(ctx, string(contents)); err == nil {
			_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.version)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			fatal("apply %s: %v", filepath.Base(migration.path), err)
		}
		if err := tx.Commit(ctx); err != nil {
			fatal("commit migration %d: %v", migration.version, err)
		}
		fmt.Printf("applied %s at %s\n", filepath.Base(migration.path), time.Now().Format(time.RFC3339))
	}
	fmt.Println("database is up to date")
}

func discover(dir string) ([]migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("invalid migration filename %q: %w", entry.Name(), err)
		}
		migrations = append(migrations, migration{version: version, path: filepath.Join(dir, entry.Name())})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].version == migrations[i].version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[i].version)
		}
	}
	return migrations, nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
