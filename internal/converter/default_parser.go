package converter

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// DefaultParser handles parsing of HCL default values
type DefaultParser struct{}

// NewDefaultParser creates a new DefaultParser
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// ParseDefaultValue parses an HCL expression into a native Go type
func (p *DefaultParser) ParseDefaultValue(expr hcl.Expression) (interface{}, error) {
	val, diags := expr.Value(nil)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to evaluate default value: %v", diags)
	}
	return p.convertCtyValue(val)
}

// convertCtyValue recursively converts a cty.Value to a native Go type.
func (p *DefaultParser) convertCtyValue(val cty.Value) (interface{}, error) {
	if val.IsNull() || !val.IsKnown() {
		return nil, nil
	}

	switch val.Type() {
	case cty.String:
		return val.AsString(), nil
	case cty.Number:
		f, _ := val.AsBigFloat().Float64()
		return f, nil
	case cty.Bool:
		return val.True(), nil
	}

	if val.Type().IsListType() || val.Type().IsSetType() || val.Type().IsTupleType() {
		var list []interface{}
		for it := val.ElementIterator(); it.Next(); {
			_, elem := it.Element()
			converted, err := p.convertCtyValue(elem)
			if err != nil {
				return nil, err
			}
			list = append(list, converted)
		}
		return list, nil
	}

	if val.Type().IsMapType() || val.Type().IsObjectType() {
		obj := make(map[string]interface{})
		for it := val.ElementIterator(); it.Next(); {
			key, elem := it.Element()
			converted, err := p.convertCtyValue(elem)
			if err != nil {
				return nil, err
			}
			obj[key.AsString()] = converted
		}
		return obj, nil
	}

	return nil, fmt.Errorf("unsupported cty value type for default: %s", val.Type().FriendlyName())
}
