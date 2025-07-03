// Package extensions provides backward compatibility bridges for the legacy registration systems
package extensions

import (
	"github.com/alex-tw-lam/tfschema/internal/converter/types"
	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/alex-tw-lam/tfschema/internal/validation"
	"github.com/hashicorp/hcl/v2"
)

// LegacyAttributeApplierBridge wraps old attribute appliers for the new system
type LegacyAttributeApplierBridge struct {
	name      string
	applyFunc func(schema *jsonschema.Schema, attribute *hcl.Attribute) error
}

func (b *LegacyAttributeApplierBridge) Name() string {
	return b.name
}

func (b *LegacyAttributeApplierBridge) Apply(schema interface{}, attribute interface{}) error {
	// Type assert the parameters to their legacy types
	jsonSchema, ok := schema.(*jsonschema.Schema)
	if !ok {
		return nil // Skip if not the expected type
	}

	hclAttr, ok := attribute.(*hcl.Attribute)
	if !ok {
		return nil // Skip if not the expected type
	}

	return b.applyFunc(jsonSchema, hclAttr)
}

// RegisterLegacyTypeConverter bridges old type converter registration to new system
func RegisterLegacyTypeConverter(name string, converter types.TypeConverter) {
	GetGlobalRegistry().RegisterTypeConverter(name, converter)
}

// RegisterLegacyValidationRule bridges old validation rule registration to new system
func RegisterLegacyValidationRule(parser validation.ParserFunc) {
	GetGlobalRegistry().RegisterValidationRule(parser)
}

// RegisterLegacyAttributeApplier bridges old attribute applier registration to new system
func RegisterLegacyAttributeApplier(name string, applyFunc func(schema *jsonschema.Schema, attribute *hcl.Attribute) error) {
	bridge := &LegacyAttributeApplierBridge{
		name:      name,
		applyFunc: applyFunc,
	}
	GetGlobalRegistry().RegisterAttributeApplier(bridge)
}

// GetLegacyTypeConverterRegistry returns the type converter registry for legacy code
func GetLegacyTypeConverterRegistry() *types.TypeConverterRegistry {
	return GetGlobalRegistry().GetTypeConverterRegistry()
}

// GetLegacyValidationRuleParsers returns validation rule parsers for legacy code
func GetLegacyValidationRuleParsers() []validation.ParserFunc {
	return GetGlobalRegistry().GetValidationRuleParsers()
}

// GetLegacyAttributeAppliers returns attribute appliers for legacy code
func GetLegacyAttributeAppliers() map[string]func(*jsonschema.Schema, *hcl.Attribute) error {
	appliers := GetGlobalRegistry().GetAttributeAppliers()
	legacy := make(map[string]func(*jsonschema.Schema, *hcl.Attribute) error)

	for name, applier := range appliers {
		if bridge, ok := applier.(*LegacyAttributeApplierBridge); ok {
			legacy[name] = bridge.applyFunc
		}
	}

	return legacy
}
