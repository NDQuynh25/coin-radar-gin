package prisma

import (
	"fmt"
	"os"

	"coin-radar-gin/config"
)

// Open creates and connects the generated Prisma client.
func Open(cfg *config.Config) (*PrismaClient, error) {
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	// Prisma Client Go reads its connection string from DATABASE_URL.
	if err := os.Setenv("DATABASE_URL", dbURL); err != nil {
		return nil, fmt.Errorf("set DATABASE_URL: %w", err)
	}

	client := NewClient()
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("connect Prisma client: %w", err)
	}

	return client, nil
}
