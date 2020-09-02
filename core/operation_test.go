package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddFieldOperationApply(t *testing.T) {
	s := GenesisSchema()
	op := AddFieldOperation{Type: IntegerType{}, Name: "age", Nullable: true}
	actualPtr, err := op.Apply(s)
	assert.NoError(t, err)
	actual := *actualPtr
	assert.Equal(t, Schema{Fields: []Field{{
		Type:     IntegerType{},
		Name:     "age",
		Nullable: true,
	}}}, actual)

	actualPtr, err = op.Apply(actual)
	assert.Error(t, err)
	assert.Equal(t, "Schema already has a field with name age", err.Error())
	assert.Nil(t, actualPtr)
}

func TestAddFieldOperationIsMigrationRequired(t *testing.T) {
	opNull := AddFieldOperation{
		Type:     IntegerType{},
		Name:     "age",
		Nullable: true,
	}
	assert.False(t, opNull.IsMigrationRequired())
	opReq := AddFieldOperation{
		Type:     IntegerType{},
		Name:     "age",
		Nullable: false,
	}
	assert.True(t, opReq.IsMigrationRequired())
}
