package migrator

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ariga.io/atlas/sql/migrate"
	ps "ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var queryCol = `
SELECT EXISTS (
    SELECT 1 
    FROM information_schema.columns 
	WHERE table_name = 'users' AND column_name = $1
)
`

var queryColCount = `
SELECT count(*) 
FROM information_schema.columns 
WHERE table_schema = 'public' 
  AND table_name = 'users'
`

func initDB(ctx context.Context) (*postgres.PostgresContainer, *sql.DB, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, err
	}

	return pgContainer, db, nil
}

func runMigration(ctx context.Context, driver migrate.Driver, sc *schema.Schema) error {
	empty, err := driver.InspectSchema(ctx, "public", nil)
	if err != nil {
		return err
	}

	diff, err := driver.SchemaDiff(empty, sc)
	if err != nil {
		return err
	}

	return driver.ApplyChanges(ctx, diff)
}

func checkColumns(t *testing.T, db *sql.DB, cols ...string) {
	require.NotNil(t, cols)

	var exists bool

	for _, col := range cols {
		err := db.QueryRowContext(t.Context(), queryCol, col).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

func queryColumnCount(ctx context.Context, db *sql.DB) (int, error) {
	count := 0

	err := db.QueryRowContext(ctx, queryColCount).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func TestMigrate(t *testing.T) {
	// PREREQUISITES
	// 1. schema

	ctx := t.Context()

	// 2. db
	pg, db, err := initDB(ctx)
	require.NoError(t, err)
	defer pg.Terminate(ctx)

	// 3. atlas driver
	driver, err := ps.Open(db)
	require.NoError(t, err)

	// TESTING
	t.Run("Creating a new table from scratch", func(t *testing.T) {
		table := schema.NewTable("users").AddColumns(
			schema.NewColumn("id").SetType(&schema.UUIDType{T: "uuid"}),
			schema.NewColumn("created_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("updated_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("deleted_at").SetType(&schema.TimeType{T: "timestamptz"}),
		)
		sc := schema.New("public").AddTables(table)
		err = runMigration(ctx, driver, sc)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at"}
		checkColumns(t, db, cols...)

		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)
	})

	if t.Failed() {
		t.Skip("Failed basic functionality test, skipping")
	}

	// Storing how many columns in table
	var N int

	t.Run("Applying new columns", func(t *testing.T) {
		table := schema.NewTable("users").AddColumns(
			schema.NewColumn("id").SetType(&schema.UUIDType{T: "uuid"}),
			schema.NewColumn("created_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("updated_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("deleted_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("username").SetType(&schema.StringType{T: "character varying", Size: 255}),
			schema.NewColumn("email").SetType(&schema.StringType{T: "character varying", Size: 255}),
		)
		newSc := schema.New("public").AddTables(table)

		err = runMigration(ctx, driver, newSc)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
		checkColumns(t, db, cols...)
		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)

		N = n
	})

	if t.Failed() {
		t.Skip("Didn't apply new columns, skipping")
	}

	t.Run("Idempotency check", func(t *testing.T) {
		table := schema.NewTable("users").AddColumns(
			schema.NewColumn("id").SetType(&schema.UUIDType{T: "uuid"}),
			schema.NewColumn("created_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("updated_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("deleted_at").SetType(&schema.TimeType{T: "timestamptz"}),
			schema.NewColumn("username").SetType(&schema.StringType{T: "character varying", Size: 255}),
			schema.NewColumn("email").SetType(&schema.StringType{T: "character varying", Size: 255}),
		)
		newSc := schema.New("public").AddTables(table)

		err = runMigration(ctx, driver, newSc)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
		checkColumns(t, db, cols...)
		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)

		assert.Equal(t, N, n)
	})
}
