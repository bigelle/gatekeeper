package migrator

import (
	"context"
	"database/sql"
	"testing"

	ps "ariga.io/atlas/sql/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func checkColumns(t *testing.T, db *sql.DB, cols ...string) {
	require.NotNil(t, cols)

	var (
		exists bool
		err    error
	)

	for _, col := range cols {
		err = db.QueryRowContext(t.Context(), queryCol, col).Scan(&exists)
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
	// PREPARE
	ctx := t.Context()

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	driver, err := ps.Open(db)
	require.NoError(t, err)

	// TEST
	t.Run("Creating a new table from scratch", func(t *testing.T) {
		table := &Table{
			Columns: []*Column{
				{Name: "id", Type: UuidDataType},
				{Name: "created_at", Type: TimestampTimeZoneDataType},
				{Name: "updated_at", Type: TimestampTimeZoneDataType},
				{Name: "deleted_at", Type: TimestampTimeZoneDataType},
			},
		}
		err = Migrate(ctx, driver, table)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at"}
		checkColumns(t, db, cols...)

		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)
	})

	// SKIP ON FAIL
	if t.Failed() {
		t.Skip("Failed basic functionality test, skipping")
	}

	// Storing how many columns in table
	var N int

	// TEST
	t.Run("Applying new columns", func(t *testing.T) {
		table := &Table{
			Columns: []*Column{
				{Name: "id", Type: UuidDataType},
				{Name: "created_at", Type: TimestampTimeZoneDataType},
				{Name: "updated_at", Type: TimestampTimeZoneDataType},
				{Name: "deleted_at", Type: TimestampTimeZoneDataType},
				{Name: "username", Type: VarCharDataType},
				{Name: "email", Type: VarCharDataType},
			},
		}

		err = Migrate(ctx, driver, table)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
		checkColumns(t, db, cols...)

		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)

		N = n
	})

	// SKIP ON FAIL
	if t.Failed() {
		t.Skip("Didn't apply new columns, skipping")
	}

	// TEST
	t.Run("Idempotency check", func(t *testing.T) {
		table := &Table{
			Columns: []*Column{
				{Name: "id", Type: UuidDataType},
				{Name: "created_at", Type: TimestampTimeZoneDataType},
				{Name: "updated_at", Type: TimestampTimeZoneDataType},
				{Name: "deleted_at", Type: TimestampTimeZoneDataType},
				{Name: "username", Type: VarCharDataType},
				{Name: "email", Type: VarCharDataType},
			},
		}

		err = Migrate(ctx, driver, table)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
		checkColumns(t, db, cols...)

		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)

		assert.Equal(t, N, n)
	})

	// TEST
	t.Run("Remove columns from the schema", func(t *testing.T) {
		table := &Table{
			Columns: []*Column{
				{Name: "id", Type: UuidDataType},
				{Name: "created_at", Type: TimestampTimeZoneDataType},
				{Name: "updated_at", Type: TimestampTimeZoneDataType},
				{Name: "deleted_at", Type: TimestampTimeZoneDataType},
				{Name: "username", Type: VarCharDataType},
			},
		}

		err = Migrate(ctx, driver, table)
		require.NoError(t, err)

		cols := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
		checkColumns(t, db, cols...)

		n, err := queryColumnCount(ctx, db)
		require.NoError(t, err)
		assert.Equal(t, len(cols), n)

		assert.Equal(t, N, n)
	})
}
