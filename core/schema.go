package core

type FieldSpec struct {
	Type     string `mapstructure:"type"`
	Nullable bool   `mapstructure:"nullable"`
}

type Schema struct {
	Fields map[string]FieldSpec `mapstructure:"fields"`
}

func GenesisSchema() Schema {
	return Schema{Fields: map[string]FieldSpec{}}
}
