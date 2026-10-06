package migrator

import (
	"testing"

	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"github.com/bigelle/gatekeeper/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakingSchemaFromConfig(t *testing.T) {
	var (
		cols []*schema.Column
		err  error
	)

	t.Run("Step 1: map from config to schema.Column", func(t *testing.T) {
		cols, err = ConfigToColumn([]config.CustomField{
			{
				Name: "username",
				Type: "custom_string",
			},
			{
				Name: "email",
				Type: "custom_string",
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, cols)
	})

	if t.Failed() {
		t.Skip("Skipping because Step 1 failed")
	}

	t.Run("Step 2: create and verify table", func(t *testing.T) {
		table := NewUsersTable(cols...)

		require.NotNil(t, table)

		assert.Equal(t, "users", table.Name)

		ps := postgres.Driver{}

		typ, err := ps.FormatType(table.Columns[0].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "uuid", typ)

		typ, err = ps.FormatType(table.Columns[1].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "timestamptz", typ)

		typ, err = ps.FormatType(table.Columns[2].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "timestamptz", typ)

		typ, err = ps.FormatType(table.Columns[3].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "timestamptz", typ)

		typ, err = ps.FormatType(table.Columns[4].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "character varying(255)", typ)

		typ, err = ps.FormatType(table.Columns[5].Type.Type)
		require.NoError(t, err)
		assert.Equal(t, "character varying(255)", typ)

	})
}
