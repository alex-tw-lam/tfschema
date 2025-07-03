package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/alex-tw-lam/tfschema/internal/validation"
	"github.com/hashicorp/hcl/v2"
)

// ValidationProcessor handles the extraction and application of validation rules.
type ValidationProcessor struct{}

// NewValidationProcessor creates a new ValidationProcessor.
func NewValidationProcessor() *ValidationProcessor {
	return &ValidationProcessor{}
}

// Process extracts and applies validation rules from the variable's blocks to the schema.
func (p *ValidationProcessor) Process(schema *jsonschema.Schema, blocks hcl.Blocks, varName string) error {
	rules, err := validation.ExtractValidationRules(blocks, varName)
	if err != nil {
		return fmt.Errorf("failed to extract validation rules: %w", err)
	}

	for _, scopedRule := range rules {
		targetSchema, err := p.findTargetSchema(varName, schema, scopedRule.Path)
		if err != nil {
			return fmt.Errorf("failed to find target schema for validation: %w", err)
		}

		if err := scopedRule.Rule.Apply(targetSchema); err != nil {
			return fmt.Errorf("failed to apply validation rule: %w", err)
		}
	}
	return nil
}

// findTargetSchema navigates the schema to find the target for a validation rule.
func (p *ValidationProcessor) findTargetSchema(
	varName string,
	schema *jsonschema.Schema,
	path []string,
) (*jsonschema.Schema, error) {
	currentSchema := schema
	if len(path) > 0 && path[0] == varName {
		path = path[1:] // Skip the root variable name
	}

	for _, segment := range path {
		if currentSchema == nil {
			return nil, fmt.Errorf("cannot apply validation to a nil schema")
		}

		// The "*" segment is a wildcard.
		baseType := getBaseType(currentSchema.Type)
		if segment == "*" {
			switch baseType {
			case "object":
				if ap, ok := currentSchema.AdditionalProperties.(*jsonschema.Schema); ok {
					currentSchema = ap
					continue
				}
				return nil, fmt.Errorf("cannot apply wildcard validation to object without schema for additional properties in '%s'", varName)
			case "array", "set":
				if currentSchema.Items != nil {
					if itemsSchema, ok := currentSchema.Items.(*jsonschema.Schema); ok {
						currentSchema = itemsSchema
						continue
					}
					// For arrays of schemas (tuples), we can't apply wildcard validation
					return nil, fmt.Errorf("cannot apply wildcard validation to tuple types in '%s'", varName)
				}
				return nil, fmt.Errorf("cannot apply wildcard validation to array with no item schema in '%s'", varName)
			default:
				return nil, fmt.Errorf("wildcard validation can only be applied to object or array/set types, not '%s' in '%s'", baseType, varName)
			}
		}

		// Handle array/tuple indexing
		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			indexStr := strings.Trim(segment, "[]")
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, fmt.Errorf("invalid index in path: %s", segment)
			}

			if currentSchema.PrefixItems != nil && index < len(currentSchema.PrefixItems) {
				currentSchema = currentSchema.PrefixItems[index]
				continue
			}

			if currentSchema.Items != nil {
				if itemsSchema, ok := currentSchema.Items.(*jsonschema.Schema); ok {
					currentSchema = itemsSchema
					continue
				}
				// For tuples with array of schemas, use the indexed schema if available
				if itemsArray, ok := currentSchema.Items.([]*jsonschema.Schema); ok && index < len(itemsArray) {
					currentSchema = itemsArray[index]
					continue
				}
			}

			return nil, fmt.Errorf("cannot apply indexed validation to a schema without Items or PrefixItems")
		}

		// Handle object property access
		if baseType == "object" && currentSchema.Properties != nil {
			if propSchema, ok := currentSchema.Properties[segment]; ok {
				currentSchema = propSchema
			} else {
				return nil, fmt.Errorf("property '%s' not found in schema for '%s'", segment, varName)
			}
		} else if baseType != "object" {
			return nil, fmt.Errorf("cannot apply validation to path segment '%s' on non-object type in '%s'", segment, varName)
		}
	}
	return currentSchema, nil
}

// getBaseType extracts the primary non-null type from a schema's Type field.
func getBaseType(t interface{}) string {
	if typeStr, ok := t.(string); ok {
		return typeStr
	}
	if typeSlice, ok := t.([]interface{}); ok {
		for _, v := range typeSlice {
			if typeStr, ok := v.(string); ok && typeStr != "null" {
				return typeStr
			}
		}
	}
	return ""
}
