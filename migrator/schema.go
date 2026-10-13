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
	Indexes []*Index
}

func (t *Table) AddColumns(cols ...*Column) {
	if t.Columns == nil {
		t.Columns = []*Column{}
	}
	t.Columns = append(t.Columns, cols...)
}

func (t *Table) AddIndexes(idx ...*Index) {
	if t.Indexes == nil {
		t.Indexes = []*Index{}
	}
	t.Indexes = append(t.Indexes, idx...)
}

type Column struct {
	// Column name
	Name string
	// Parsed SQL data type
	Type DataType
}

type Index struct {
	// Column name
	Name string
	// Unique index
	Unique bool
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
		// 1. add the column
		var col Column
		switch extra.Type {
		case config.CustomStringFieldType:
			col = Column{Name: extra.Name, Type: VarCharDataType}
		default:
			return nil, fmt.Errorf("unknown data type: %s", extra.Name)
		}
		t.AddColumns(&col)

		// 2. add indexes
		if extra.Unique {
			extra.Index = true
		}
		if extra.Index {
			t.AddIndexes(&Index{Name: col.Name, Unique: extra.Unique})
		}
	}

	return &t, nil
}
