package config

import "github.com/spf13/viper"

// Predefined field types
const (
	CustomStringFieldType string = "custom_string"
)

type Config struct {
	Db *Database `mapstructure:"db"`
}

type Database struct {
	User *User `mapstructure:"user"`
}

type User struct {
	CustomFields []CustomField `mapstructure:"custom_fields"`
}

type CustomField struct {
	Name string `mapstructure:"name"`
	Type string `mapstructure:"type"`
}

func LoadConfig(paths ...string) (*Config, error) {
	v := viper.New()

	setupDefaults(v)

	for _, path := range paths {
		v.AddConfigPath(path)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func setupDefaults(v *viper.Viper) {
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
}
