package core

import "fmt"

type SchemaOperation interface {
	Apply(Schema) (*Schema, error)
	IsMigrationRequired() bool
}

type AddFieldOperation struct {
	Type     Type
	Name     string
	Nullable bool
}

func (op AddFieldOperation) Apply(s Schema) (*Schema, error) {
	for _, field := range s.Fields {
		if field.Name == op.Name {
			return nil, fmt.Errorf("Schema already has a field with name %s", op.Name)
		}
	}
	newSchema := s.WithAddedField(Field{op.Type, op.Name, op.Nullable})
	return &newSchema, nil
}

func (op AddFieldOperation) IsMigrationRequired() bool {
	return !op.Nullable
}
