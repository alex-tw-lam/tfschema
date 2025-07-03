package types

import (
	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// TypeConverter defines the interface for converting a Terraform type expression to a JSON schema.
type TypeConverter interface {
	Convert(expr hcl.Expression) (*jsonschema.Schema, error)
}

// TypeConverterWithIsOptional is an interface that the main converter will implement
// to allow recursive type conversions without causing a circular dependency.
type TypeConverterWithIsOptional interface {
	ConvertType(expr hcl.Expression) (*jsonschema.Schema, error)
	IsOptionalType(expr hcl.Expression) bool
}

// TypeConverterRegistry holds a map of type converters.
type TypeConverterRegistry struct {
	converters map[string]TypeConverter
}

// NewTypeConverterRegistry creates a new registry.
func NewTypeConverterRegistry() *TypeConverterRegistry {
	return &TypeConverterRegistry{
		converters: make(map[string]TypeConverter),
	}
}

// Register adds a type converter to the registry.
func (r *TypeConverterRegistry) Register(typeName string, converter TypeConverter) {
	r.converters[typeName] = converter
}

// Get retrieves a type converter from the registry.
func (r *TypeConverterRegistry) Get(typeName string) (TypeConverter, error) {
	converter, found := r.converters[typeName]
	if !found {
		// Fallback for primitive types
		if isPrimitive(typeName) {
			return r.converters["primitive"], nil
		}
		return nil, &ErrUnknownType{TypeName: typeName}
	}
	return converter, nil
}

// isPrimitive checks if a type name is a primitive type.
func isPrimitive(typeName string) bool {
	switch typeName {
	case "string", "number", "bool":
		return true
	default:
		return false
	}
}

// ErrUnknownType is returned when a type converter for a given type name is not found.
type ErrUnknownType struct {
	TypeName string
}

func (e *ErrUnknownType) Error() string {
	return "unknown type: " + e.TypeName
}
