package core

import "fmt"

type FieldSpec struct {
	Type     string `mapstructure:"type"`
	Nullable bool   `mapstructure:"nullable"`
}

type Schema struct {
	Fields     map[string]FieldSpec `mapstructure:"fields"`
	PrimaryKey []string             `mapstructure:"primaryKey"`
}

func GenesisSchema() Schema {
	return Schema{Fields: map[string]FieldSpec{}}
}

func createError(fieldName string, expectedType string, v interface{}) error {
	return fmt.Errorf("Invalid type %T for field %s. Expected %s.", v, fieldName, expectedType)
}

func isValidType(fieldName string, typeStr string, v interface{}) error {
	switch typeStr {
	case "string":
		if _, ok := v.(string); !ok {
			return createError(fieldName, "string", v)
		}
	case "int":
		if _, ok := v.(int32); !ok {
			return createError(fieldName, "int32", v)
		}
	case "long":
		if _, ok := v.(int64); !ok {
			return createError(fieldName, "int64", v)
		}
	case "float":
		if _, ok := v.(float32); !ok {
			return createError(fieldName, "float32", v)
		}
	case "double":
		if _, ok := v.(float64); !ok {
			return createError(fieldName, "float64", v)
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return createError(fieldName, "bool", v)
		}
	default:
		return fmt.Errorf("Unsupported type %s", typeStr)
	}
	return nil
}

func (s Schema) ValidateRow(r Row) error {
	for fieldName, fieldSpec := range s.Fields {
		v, ok := r.Data[fieldName]
		if !ok {
			if !fieldSpec.Nullable {
				return fmt.Errorf("Required field missing %s", fieldName)
			}
		} else {
			if err := isValidType(fieldName, fieldSpec.Type, v); err != nil {
				return err
			}
		}
	}
	for fieldName := range r.Data {
		if _, ok := s.Fields[fieldName]; !ok {
			return fmt.Errorf("Extraneous field %s", fieldName)
		}
	}
	return nil
}
