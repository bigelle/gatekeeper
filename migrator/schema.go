package migrator

import (
	"fmt"

	"github.com/bigelle/gatekeeper/config"
)

type DataType string

const (
	UuidDataType              DataType = "uuid"
	TimestampTimeZoneDataType DataType = "timestamptz"
	VarCharDataType           DataType = "varchar"
)

type Table struct {
	Columns []*Column
}

func (t *Table) AddColumns(cols ...*Column) {
	t.Columns = append(t.Columns, cols...)
}

type Column struct {
	// Column name
	Name string
	// Parsed SQL data type
	Type DataType
}

func NewUsersTable(extras ...*config.CustomField) (*Table, error) {
	t := Table{}

	t.AddColumns(
		&Column{Name: "id", Type: "uuid"},
		&Column{Name: "created_at", Type: TimestampTimeZoneDataType},
		&Column{Name: "updated_at", Type: TimestampTimeZoneDataType},
		&Column{Name: "deleted_at", Type: TimestampTimeZoneDataType},
	)

	for _, extra := range extras {
		// First add the column
		var col Column
		switch extra.Type {
		case config.CustomStringFieldType:
			col = Column{Name: extra.Name, Type: VarCharDataType}
		default:
			return nil, fmt.Errorf("unknown data type: %s", extra.Name)
		}
		t.AddColumns(&col)

	}

	return &t, nil
}
