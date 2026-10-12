package migrator

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

func postgresMigrate(ctx context.Context, db *pgx.Conn, t *Table) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	// 1. create table if not exists
	if err := createTable(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// 2. in case the table already existed, make sure it has required columns
	if err := assertDefaultColumns(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// 3. add columns
	if err := addColumns(ctx, tx, t.Columns); err != nil {
		tx.Rollback(ctx)
		return err
	}

	// 4. add indexes

	// 5. add constraints

	// done
	return tx.Commit(ctx)
}

func createTable(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, _PGX_CREATE_USERS_TABLE_REQUEST)
	if err != nil {
		return err
	}
	return nil
}

func assertDefaultColumns(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, _PGX_SELECT_COLUMN_NAMES_REQUEST)
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

func findMissing(src, target []string) []string {
	var diff []string
	for _, t := range target {
		if !slices.Contains(src, t) {
			diff = append(diff, t)
		}
	}

	return diff
}

var colNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)

func addColumns(ctx context.Context, tx pgx.Tx, cols []*Column) error {
	b := strings.Builder{}
	b.WriteString("ALTER TABLE users ")

	var rows []string

	for _, col := range cols {
		if !colNameRegex.MatchString(col.Name) {
			return fmt.Errorf("invalid column name: %s", col.Name)
		}

		row := fmt.Sprintf("ADD COLUMN IF NOT EXISTS %s %s", col.Name, col.Type)
		rows = append(rows, row)
	}

	b.WriteString(strings.Join(rows, ",\n"))

	b.WriteRune(';')

	_, err := tx.Exec(ctx, b.String())
	if err != nil {
		return err
	}

	return nil
}
