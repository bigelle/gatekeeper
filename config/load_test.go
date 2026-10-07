package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	t.Log("trying tempdir:", dir)

	path := filepath.Join(dir, "config.yaml")
	file, err := os.Create(path)
	require.NoError(t, err)

	yaml := `
db:
  user:
    custom_fields:
      - name: username
        type: custom_string
      - name: email
        type: custom_string
`

	n, err := file.WriteString(yaml)
	require.NoError(t, err)
	require.Equal(t, len(yaml), n)

	require.NoError(t, file.Close())

	_, err = os.Stat(path)
	require.NoError(t, err)

	cfg, err := LoadConfig(dir)
	require.NoError(t, err)

	assert.Equal(t, &Config{
		Db: &Database{
			User: &User{
				CustomFields: []CustomField{
					{Name: "username", Type: "custom_string"},
					{Name: "email", Type: "custom_string"},
				},
			},
		},
	}, cfg)
}
