package migrator

import (
	"testing"

	"github.com/bigelle/gatekeeper/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakingSchemaFromConfig(t *testing.T) {
	extras := []*config.CustomField{
		{
			Name: "email",
			Type: config.CustomStringFieldType,
		},
	}

	table, err := NewUsersTable(extras...)
	require.NoError(t, err)

	// skipping first 4 columns, checking 5th
	assert.Equal(t, table.Columns[4].Name, "email")
	assert.Equal(t, table.Columns[4].Type, VarCharDataType)
}
