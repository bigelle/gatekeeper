package migrator

import (
	"testing"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const queryTableIndexes = `
SELECT indexname
FROM pg_indexes
WHERE schemaname = 'public' 
AND tablename = 'users';
`

func Test_postgresMigrate(t *testing.T) {
	// PREPARE
	table := Table{
		Columns: []*Column{
			{Name: "email", Type: VarCharDataType},
			{Name: "username", Type: VarCharDataType},
		},
		Indexes: []*Index{
			{Name: "email", Unique: false},
			{Name: "username", Unique: true},
		},
	}
	ctx := t.Context()

	conn, err := pgx.Connect(ctx, connStr)
	require.NoError(t, err)

	m := newPostgresMigrator(ctx, conn)

	// RUN
	err = m.migrate(&table)
	require.NoError(t, err)

	//TEST for columns
	rows, err := conn.Query(ctx, _PGX_SELECT_COLUMN_NAMES_REQUEST)
	require.NoError(t, err)

	colsHave, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	colsExpect := []string{"id", "created_at", "updated_at", "deleted_at", "username", "email"}
	diff := findMissing(colsHave, colsExpect)
	assert.Empty(t, diff)

	// TEST for indexes
	rows, err = conn.Query(ctx, queryTableIndexes)
	require.NoError(t, err)

	indexHave, err := pgx.CollectRows(rows, pgx.RowTo[string])
	require.NoError(t, err)

	indexExpect := []string{"idx_users_email", "idx_users_username_unique"}
	diff = findMissing(indexHave, indexExpect)
	assert.Empty(t, diff)
}
