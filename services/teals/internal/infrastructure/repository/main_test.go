package repository_test

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed sql/migrations/*.up.sql
var migrationsFS embed.FS

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Deferral and os.Exit don't mix.
	// Wrap in run() so defers execute before os.Exit.
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	// 1. Start Container
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("teals_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Printf("start postgres container: %v\n", err)
		return 1
	}

	// GUARANTEE CONTAINER CLEANUP
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("terminate container: %v\n", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("get connection string: %v\n", err)
		return 1
	}

	// 2. Initialize Pool
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Printf("create pool: %v\n", err)
		return 1
	}

	// GUARANTEE POOL CLEANUP
	defer testPool.Close()

	// 3. Run Migrations
	if err := runMigrations(ctx, testPool); err != nil {
		log.Printf("run migrations: %v\n", err)
		return 1
	}

	// 4. Run Tests
	return m.Run()
}

// runMigrations executes all *.up.sql files in chronological filename order.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.Glob(migrationsFS, "sql/migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(entries)

	for _, entry := range entries {
		sql, err := migrationsFS.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("read %s: %w", entry, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", entry, err)
		}
	}
	return nil
}

// truncateTables clears all application tables and resets sequences.
func truncateTables(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := testPool.Exec(ctx, `
		TRUNCATE TABLE
			teals.log_entry,
			teals.mmr_node,
			teals.producer_key,
			teals.producer,
			teals.checkpoint,
			teals.subject_secret
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
