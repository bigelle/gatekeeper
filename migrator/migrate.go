package migrator

import (
	"context"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
)

func Migrate(ctx context.Context, driver migrate.Driver, t *schema.Table) error {
	current, err := driver.InspectSchema(ctx, "public", nil)
	if err != nil {
		return err
	}

	diff, err := driver.SchemaDiff(current, schema.New("public").AddTables(t), nil)
	if err != nil {
		return err
	}

	return driver.ApplyChanges(ctx, diff)
}
