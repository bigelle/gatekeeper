package migrator

import "ariga.io/atlas/sql/schema"

func NewUsersTable() *schema.Table {
	t := schema.NewTable("users")

	uuidCol := schema.NewColumn("id").
		SetType(&schema.UUIDType{T: "uuid"}). // NOTE: postgres exclusive
		SetDefault(&schema.RawExpr{X: "uuidv7()"})
	createdAtCol := schema.NewColumn("created_at").
		SetType(&schema.TimeType{T: "timestamptz"}).
		SetDefault(&schema.RawExpr{X: "CURRENT_TIMESTAMP"})
	updatedAtCol := schema.NewColumn("updated_at").
		SetType(&schema.TimeType{T: "timestamptz"})
	deletedAtCol := schema.NewColumn("deleted_at").
		SetType(&schema.TimeType{T: "timestamptz"})

	t.AddColumns(
		uuidCol,
		createdAtCol,
		updatedAtCol,
		deletedAtCol,
	)

	return t
}
