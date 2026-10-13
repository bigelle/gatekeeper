package migrator

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

const _PGX_CREATE_USERS_TABLE_REQUEST string = `
CREATE TABLE IF NOT EXISTS users (
	id uuid PRIMARY KEY DEFAULT uuidv7(),
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz,
	deleted_at timestamptz
);
`

const _PGX_SELECT_COLUMN_NAMES_REQUEST string = `
SELECT column_name
FROM information_schema.columns
WHERE table_schema = 'public'
AND table_name = 'users'
;
`

type postgresMigrator struct {
	// migration is a whole one procedure
	// so the single context is used for the whole proccess
	ctx  context.Context
	conn *pgx.Conn
}

func newPostgresMigrator(ctx context.Context, conn *pgx.Conn) *postgresMigrator {
	return &postgresMigrator{
		ctx:  ctx,
		conn: conn,
	}
}

func (m *postgresMigrator) migrate(t *Table) error {
	tx, err := m.conn.Begin(m.ctx)
	if err != nil {
		return err
	}

	// 1. create table if not exists
	if err := m.createTable(tx); err != nil {
		tx.Rollback(m.ctx)
		return err
	}

	// 2. in case the table already existed, make sure it has required columns
	if err := m.assertDefaultColumns(tx); err != nil {
		tx.Rollback(m.ctx)
		return err
	}

	// 3. add extra columns
	if err := m.addColumns(tx, t.Columns); err != nil {
		tx.Rollback(m.ctx)
		return err
	}

	// 4. add indexes
	if err := m.addIndexes(tx, t.Indexes); err != nil {
		tx.Rollback(m.ctx)
		return err
	}

	// 5. add constraints

	// done
	return tx.Commit(m.ctx)
}

func (m *postgresMigrator) createTable(tx pgx.Tx) error {
	_, err := tx.Exec(m.ctx, _PGX_CREATE_USERS_TABLE_REQUEST)
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresMigrator) assertDefaultColumns(tx pgx.Tx) error {
	rows, err := tx.Query(m.ctx, _PGX_SELECT_COLUMN_NAMES_REQUEST)
	if err != nil {
		return err
	}

	cols, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}

	diff := findMissing(cols, []string{"created_at", "updated_at", "deleted_at", "id"})
	if len(diff) != 0 {
		return fmt.Errorf("table 'users' doesn't have some of the required columns. diff: %s", strings.Join(diff, ", "))
	}

	return nil
}

// findMissing returns what strings are missing in "in" from "from"
func findMissing(in, from []string) []string {
	var diff []string
	for _, t := range from {
		if !slices.Contains(in, t) {
			diff = append(diff, t)
		}
	}

	return diff
}

func (m *postgresMigrator) addColumns(tx pgx.Tx, cols []*Column) error {
	b := strings.Builder{}
	b.WriteString("ALTER TABLE users ")

	var rows []string

	for _, col := range cols {
		colname := pgx.Identifier{col.Name}.Sanitize()
		coltype := pgx.Identifier{string(col.Type)}.Sanitize()
		row := fmt.Sprintf("ADD COLUMN IF NOT EXISTS %s %s", colname, coltype)
		rows = append(rows, row)
	}

	b.WriteString(strings.Join(rows, ",\n"))

	b.WriteRune(';')

	_, err := tx.Exec(m.ctx, b.String())
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresMigrator) addIndexes(tx pgx.Tx, idx []*Index) error {
	for _, i := range idx {
		name := "idx_users_" + i.Name
		if i.Unique {
			name += "_unique"
		}

		name = pgx.Identifier{name}.Sanitize()
		colname := pgx.Identifier{i.Name}.Sanitize()

		q := fmt.Sprintf("CREATE INDEX %s ON users (%s);", name, colname)

		_, err := tx.Exec(m.ctx, q)
		if err != nil {
			return err
		}
	}

	return nil
}
