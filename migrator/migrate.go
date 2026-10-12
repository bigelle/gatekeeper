package migrator

import (
	"context"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
)

func Migrate(ctx context.Context, driver migrate.Driver, t *Table) error {
	atl := mapToAtlasTable(t)

	current, err := driver.InspectSchema(ctx, "public", nil)
	if err != nil {
		return err
	}

	skip := []schema.Change{
		&schema.DropAttr{},
		&schema.DropCheck{},
		&schema.DropIndex{},
		&schema.DropTable{},
		&schema.DropColumn{},
		&schema.DropObject{},
		&schema.DropSchema{},
		&schema.DropForeignKey{},
		&schema.DropPrimaryKey{},
	}
	diff, err := driver.SchemaDiff(current, schema.New("public").AddTables(atl), schema.DiffSkipChanges(skip...))
	if err != nil {
		return err
	}

	return driver.ApplyChanges(ctx, diff)
}

func mapToAtlasTable(t *Table) *schema.Table {
	atl := schema.NewTable("users")

	// First the columns
	for _, c := range t.Columns {
		col := schema.NewColumn(c.Name)

		pgType, err := postgres.ParseType(string(c.Type))
		if err != nil {
			// it means i messed up in the config->schema mapper,
			// there's nothing you can do
			panic(err)
		}
		col.SetType(pgType)

		atl.AddColumns(col)
	}

	return atl
}
