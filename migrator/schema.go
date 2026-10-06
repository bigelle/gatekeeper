package migrator

import (
	"fmt"

	"ariga.io/atlas/sql/schema"
	"github.com/bigelle/gatekeeper/config"
)

func NewUsersTable(extra ...*schema.Column) *schema.Table {
	t := schema.NewTable("users")

	idCol := schema.NewColumn("id").
		SetType(&schema.UUIDType{T: "uuid"}). // NOTE: postgres exclusive
		SetDefault(&schema.RawExpr{X: "uuidv7()"})
	createdAtCol := schema.NewColumn("created_at").
		SetType(&schema.TimeType{T: "timestamp with time zone"}).
		SetDefault(&schema.RawExpr{X: "CURRENT_TIMESTAMP"})
	updatedAtCol := schema.NewColumn("updated_at").
		SetType(&schema.TimeType{T: "timestamp with time zone"})
	deletedAtCol := schema.NewColumn("deleted_at").
		SetType(&schema.TimeType{T: "timestamp with time zone"})

	t.AddColumns(
		idCol,
		createdAtCol,
		updatedAtCol,
		deletedAtCol,
	)

	t.Columns = append(t.Columns, extra...)

	return t
}

func ConfigToColumn(fields []config.CustomField) ([]*schema.Column, error) {
	cols := make([]*schema.Column, len(fields))

	for i, field := range fields {
		col := schema.NewColumn(field.Name)

		if err := assignColType(col, field.Type); err != nil {
			return nil, err
		}

		cols[i] = col
	}

	return cols, nil
}

func assignColType(col *schema.Column, t string) error {
	switch t {
	case config.CustomStringFieldType:
		col.SetType(&schema.StringType{T: "varchar", Size: 255})

	default:
		return fmt.Errorf("unsupported data type: %s", t)
	}

	return nil
}
