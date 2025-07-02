package converter

import (
	"fmt"

	"github.com/atwlam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// TypeInferenceHandler handles inference of specific object schemas from default values
type TypeInferenceHandler struct {
	defaultParser *DefaultParser
	converter     *Converter
}

// NewTypeInferenceHandler creates a new type inference handler
func NewTypeInferenceHandler(defaultParser *DefaultParser, converter *Converter) *TypeInferenceHandler {
	return &TypeInferenceHandler{
		defaultParser: defaultParser,
		converter:     converter,
	}
}

// InferSchemaFromDefault analyzes a default value and infers a more specific schema
// when the default value structure is more specific than the type definition
func (t *TypeInferenceHandler) InferSchemaFromDefault(typeSchema *jsonschema.Schema, defaultValue interface{}) *jsonschema.Schema {
	// Recursively apply type inference to the schema
	return t.inferSchemaRecursive(typeSchema, defaultValue)
}

// inferSchemaRecursive recursively applies type inference to a schema and its nested components
func (t *TypeInferenceHandler) inferSchemaRecursive(typeSchema *jsonschema.Schema, defaultValue interface{}) *jsonschema.Schema {
	fmt.Printf("DEBUG: inferSchemaRecursive called with typeSchema.Type=%s, hasProperties=%v, hasAdditionalProperties=%v\n",
		typeSchema.Type, typeSchema.Properties != nil, typeSchema.AdditionalProperties != nil)

	// If the type is already a specific object (not a map), process its properties
	if typeSchema.Type == "object" && typeSchema.Properties != nil {
		fmt.Printf("DEBUG: Processing specific object with properties\n")
		// Recursively process each property
		for propName, propSchema := range typeSchema.Properties {
			if defaultObj, ok := defaultValue.(map[string]interface{}); ok {
				if propDefault, exists := defaultObj[propName]; exists {
					fmt.Printf("DEBUG: Processing property '%s' with default value: %+v\n", propName, propDefault)
					inferredPropSchema := t.inferSchemaRecursive(propSchema, propDefault)
					if inferredPropSchema != propSchema {
						fmt.Printf("DEBUG: Property '%s' schema was inferred and changed\n", propName)
						typeSchema.Properties[propName] = inferredPropSchema
					}
				}
			}
		}
		return typeSchema
	}

	// If the default value is not an object, return the original schema
	if typeSchema.Type != "object" || (typeSchema.Properties == nil && typeSchema.AdditionalProperties == nil) {
		return typeSchema
	}

	// If the type is a map (has additionalProperties), try to infer a specific object
	fmt.Printf("DEBUG: Checking for map type with additionalProperties. typeSchema.Type=%s, hasAdditionalProperties=%v, defaultValue=%#v\n", typeSchema.Type, typeSchema.AdditionalProperties != nil, defaultValue)
	if typeSchema.Type == "object" && typeSchema.AdditionalProperties != nil {
		if defaultObj, ok := defaultValue.(map[string]interface{}); ok && t.hasConsistentStructure(defaultObj) {
			fmt.Printf("DEBUG: Inferring specific object schema for map type with default: %+v\n", defaultObj)
			return t.createInferredObjectSchema(defaultObj, typeSchema.AdditionalProperties)
		}
	}

	// Handle arrays by processing their items
	if typeSchema.Type == "array" && typeSchema.Items != nil {
		if defaultArray, ok := defaultValue.([]interface{}); ok && len(defaultArray) > 0 {
			itemDefault := defaultArray[0]
			// Handle Items as either *Schema or []*Schema
			if itemsSchema, ok := typeSchema.Items.(*jsonschema.Schema); ok {
				inferredItemsSchema := t.inferSchemaRecursive(itemsSchema, itemDefault)
				if inferredItemsSchema != itemsSchema {
					fmt.Printf("DEBUG: Array items schema was inferred and changed\n")
					typeSchema.Items = inferredItemsSchema
				}
				// If the items schema is an object, recurse into its properties
				if inferredItemsSchema.Type == "object" && inferredItemsSchema.Properties != nil {
					if itemDefaultMap, ok := itemDefault.(map[string]interface{}); ok {
						for propName, propSchema := range inferredItemsSchema.Properties {
							if propDefault, exists := itemDefaultMap[propName]; exists {
								inferredPropSchema := t.inferSchemaRecursive(propSchema, propDefault)
								if inferredPropSchema != propSchema {
									fmt.Printf("DEBUG: Array item property '%s' schema was inferred and changed\n", propName)
									inferredItemsSchema.Properties[propName] = inferredPropSchema
								}
							}
						}
					}
				}
			}
			// For array of schemas (tuple case), skip inference for now
		}
	}

	return typeSchema
}

// hasConsistentStructure checks if a default object has a consistent structure
// that could be inferred as a specific object schema
func (t *TypeInferenceHandler) hasConsistentStructure(obj map[string]interface{}) bool {
	// For now, we'll consider any object with properties as having consistent structure
	// In a more sophisticated implementation, we could analyze multiple default values
	// to ensure consistency across different instances
	return len(obj) > 0
}

// createInferredObjectSchema creates a specific object schema from a default value
func (t *TypeInferenceHandler) createInferredObjectSchema(defaultObj map[string]interface{}, valueSchema interface{}) *jsonschema.Schema {
	fmt.Printf("DEBUG: createInferredObjectSchema called with keys: %v\n", getMapKeys(defaultObj))
	schema := &jsonschema.Schema{
		Type:       "object",
		Properties: make(map[string]*jsonschema.Schema),
		Required:   &[]string{},
	}

	// Convert the value schema to a proper Schema object
	var valueTypeSchema *jsonschema.Schema
	if valueSchemaMap, ok := valueSchema.(map[string]interface{}); ok {
		// This is a simplified conversion - in practice, we'd need more sophisticated handling
		if typeStr, ok := valueSchemaMap["type"].(string); ok {
			valueTypeSchema = &jsonschema.Schema{Type: typeStr}
		}
	} else if valueSchemaPtr, ok := valueSchema.(*jsonschema.Schema); ok {
		valueTypeSchema = valueSchemaPtr
	}

	if valueTypeSchema == nil {
		// Fallback to string type if we can't determine the value schema
		valueTypeSchema = &jsonschema.Schema{Type: "string"}
	}

	// Create properties based on the default object structure
	for key, value := range defaultObj {
		// Analyze the actual value type for better inference
		var propSchema *jsonschema.Schema
		switch v := value.(type) {
		case string:
			propSchema = &jsonschema.Schema{Type: "string"}
		case float64:
			propSchema = &jsonschema.Schema{Type: "number"}
		case bool:
			propSchema = &jsonschema.Schema{Type: "boolean"}
		case map[string]interface{}:
			// Recursively infer schema for nested objects
			propSchema = t.inferSchemaRecursive(&jsonschema.Schema{Type: "object"}, v)
		case []interface{}:
			// Handle arrays
			if len(v) > 0 {
				itemSchema := t.inferSchemaRecursive(&jsonschema.Schema{}, v[0])
				propSchema = &jsonschema.Schema{
					Type:  "array",
					Items: itemSchema,
				}
			} else {
				propSchema = &jsonschema.Schema{Type: "array"}
			}
		default:
			// Fallback to the value schema
			propSchema = valueTypeSchema
		}

		schema.Properties[key] = propSchema
		*schema.Required = append(*schema.Required, key)
	}

	return schema
}

// getMapKeys returns the keys of a map as a slice of strings
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// InferSchemaFromObjectExpression analyzes an object expression and infers schema
func (t *TypeInferenceHandler) InferSchemaFromObjectExpression(objExpr *hclsyntax.ObjectConsExpr) (*jsonschema.Schema, error) {
	schema := &jsonschema.Schema{
		Type:       "object",
		Properties: make(map[string]*jsonschema.Schema),
		Required:   &[]string{},
	}

	for _, item := range objExpr.Items {
		// Get the key
		keyVal, diags := item.KeyExpr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate object key: %v", diags)
		}
		if keyVal.Type() != cty.String {
			return nil, fmt.Errorf("object key must be a string")
		}
		key := keyVal.AsString()

		// Convert the value type
		valueSchema, err := t.converter.ConvertType(item.ValueExpr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert property type for key '%s': %w", key, err)
		}

		schema.Properties[key] = valueSchema
		*schema.Required = append(*schema.Required, key)
	}

	return schema, nil
}
