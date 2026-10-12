package migrator

import (
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const queryTableCols = `
SELECT
FROM information_schema.columns 
WHERE table_schema = 'public' 
  AND table_name = 'users'
`

func Test_postgresMigrate(t *testing.T) {
	// PREPARE
	table := Table{
		Columns: []*Column{
			{Name: "email", Type: VarCharDataType},
			{Name: "username", Type: VarCharDataType},
		},
	}

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)

	conn, err := db.Conn(t.Context())
	require.NoError(t, err)

	// RUN
	err = conn.Raw(func(driverConn any) error {
		conn := driverConn.(*stdlib.Conn).Conn()

		return postgresMigrate(t.Context(), conn, &table)
	})
	require.NoError(t, err)

	//TEST
	rows, err := db.Query(queryTableCols)
	require.NoError(t, err)
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var col string
		rows.Scan(&col)
		cols = append(cols, col)
	}
	err = rows.Err()
	require.NoError(t, err)

	assert.Equal(t, 6, len(cols))
}
