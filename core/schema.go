package schema

type Type interface {
	Name() string
}

type StringType struct{}

func (t StringType) Name() string {
	return "string"
}

type IntegerType struct{}

func (t IntegerType) Name() string {
	return "int"
}

type LongType struct{}

func (t LongType) Name() string {
	return "long"
}

type FloatType struct{}

func (t FloatType) Name() string {
	return "float"
}

type DoubleType struct{}

func (t DoubleType) Name() string {
	return "double"
}

type BooleanType struct{}

func (t BooleanType) Name() string {
	return "bool"
}

type Field struct {
	Type     Type
	Name     string
	Nullable bool
}

type Schema struct {
	Fields []Field
}
