package migrator

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var connStr string

func initPgContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", err
	}

	return pgContainer, connStr, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, c, err := initPgContainer(ctx)
	if err != nil {
		log.Fatalf("failed to start container: %s", err.Error())
	}

	connStr = c

	code := m.Run()

	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}
