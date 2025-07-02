package converter

// Applies each parsed validation rule to the matching spot in the schema.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/alex-tw-lam/tfschema/internal/validation"
	"github.com/hashicorp/hcl/v2"
)

// processValidations extracts the validation rules from a variable's blocks
// and applies them to the matching spots in its schema.
func processValidations(schema *jsonschema.Schema, blocks hcl.Blocks, varName string) error {
	rules, err := validation.ExtractValidationRules(blocks, varName)
	if err != nil {
		return fmt.Errorf("failed to extract validation rules: %w", err)
	}
	for _, scopedRule := range rules {
		target, err := findTargetSchema(varName, schema, scopedRule.Path)
		if err != nil {
			return fmt.Errorf("failed to find target schema for validation: %w", err)
		}
		scopedRule.Rule.Apply(target)
	}
	return nil
}

// findTargetSchema navigates the schema along a path of property names,
// "[N]" array indices, or "*" wildcards.
func findTargetSchema(varName string, schema *jsonschema.Schema, path []string) (*jsonschema.Schema, error) {
	currentSchema := schema
	if len(path) > 0 && path[0] == varName {
		path = path[1:] // skip the root variable name
	}

	for _, segment := range path {
		if currentSchema == nil {
			return nil, fmt.Errorf("cannot apply validation to a nil schema")
		}

		var next *jsonschema.Schema
		var err error
		switch {
		case segment == "*":
			next, err = descendWildcard(currentSchema, varName)
		case strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]"):
			next, err = descendIndex(currentSchema, segment)
		default:
			next, err = descendProperty(currentSchema, segment, varName)
		}
		if err != nil {
			return nil, err
		}
		currentSchema = next
	}
	return currentSchema, nil
}

// descendWildcard returns the schema that every element or value must match:
// the item schema of an array, or the value schema of a map.
func descendWildcard(s *jsonschema.Schema, varName string) (*jsonschema.Schema, error) {
	switch s.Type {
	case "object": // a map: all values share one schema
		if s.AdditionalProperties != nil && s.AdditionalProperties.ValueSchema != nil {
			return s.AdditionalProperties.ValueSchema, nil
		}
		return nil, fmt.Errorf("cannot apply wildcard validation to object without schema for additional properties in '%s'", varName)
	case "array": // a list or set (sets arrive here as arrays too)
		if s.Items != nil && s.Items.Single != nil {
			return s.Items.Single, nil
		}
		return nil, fmt.Errorf("cannot apply wildcard validation to array with no item schema in '%s'", varName)
	}
	return nil, fmt.Errorf("wildcard validation can only be applied to object or array types, not '%s' in '%s'", s.Type, varName)
}

// descendIndex returns the schema at one array position, e.g. "[2]".
func descendIndex(s *jsonschema.Schema, segment string) (*jsonschema.Schema, error) {
	index, err := strconv.Atoi(strings.Trim(segment, "[]"))
	if err != nil {
		return nil, fmt.Errorf("invalid index in path: %s", segment)
	}
	if s.Items == nil {
		return nil, fmt.Errorf("cannot apply indexed validation to a schema without items")
	}
	if s.Items.Single != nil { // a list: every position shares one schema
		return s.Items.Single, nil
	}
	if index < len(s.Items.Tuple) { // a tuple: each position has its own schema
		return s.Items.Tuple[index], nil
	}
	return nil, fmt.Errorf("index %d out of range for tuple with %d positions", index, len(s.Items.Tuple))
}

// descendProperty returns the schema of one object property.
func descendProperty(s *jsonschema.Schema, name, varName string) (*jsonschema.Schema, error) {
	if s.Type != "object" {
		return nil, fmt.Errorf("cannot apply validation to path segment '%s' on non-object type in '%s'", name, varName)
	}
	prop, ok := s.Properties[name]
	if !ok {
		return nil, fmt.Errorf("property '%s' not found in schema for '%s'", name, varName)
	}
	return prop, nil
}
