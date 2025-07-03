package converter

import (
	"fmt"
	"sort"

	"github.com/alex-tw-lam/tfschema/internal/converter/types"
	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Converter converts Terraform variable definitions to JSON Schema
type Converter struct {
	parser                *hclparse.Parser
	defaultParser         *DefaultParser
	attributeProcessor    *AttributeProcessor
	validationProcessor   *ValidationProcessor
	typeInferenceHandler  *TypeInferenceHandler
	typeConverterRegistry *types.TypeConverterRegistry
}

// New creates a new Converter instance
func New() *Converter {
	defaultParser := NewDefaultParser()
	c := &Converter{
		parser:              hclparse.NewParser(),
		defaultParser:       defaultParser,
		validationProcessor: NewValidationProcessor(),
	}
	c.typeInferenceHandler = NewTypeInferenceHandler(c.defaultParser, c)

	// Initialize components
	c.initializeAttributeProcessor(defaultParser)
	c.initializeTypeConverters()

	return c
}

// initializeAttributeProcessor sets up the attribute processor with all default appliers
func (c *Converter) initializeAttributeProcessor(defaultParser *DefaultParser) {
	c.attributeProcessor = NewAttributeProcessor(
		NewDescriptionAttributeApplier(),
		NewDefaultAttributeApplier(defaultParser),
		NewSensitiveAttributeApplier(),
		NewNullableAttributeApplier(),
	)
}

// initializeTypeConverters sets up all type converters with proper dependency injection
func (c *Converter) initializeTypeConverters() {
	c.typeConverterRegistry = types.NewTypeConverterRegistry()
	c.typeConverterRegistry.Register("primitive", types.NewPrimitiveTypeConverter())
	c.typeConverterRegistry.Register("list", types.NewListTypeConverter(c))
	c.typeConverterRegistry.Register("object", types.NewObjectTypeConverter(c))
	c.typeConverterRegistry.Register("map", types.NewMapTypeConverter(c))
	c.typeConverterRegistry.Register("set", types.NewSetConverter(c))
	c.typeConverterRegistry.Register("optional", types.NewOptionalTypeConverter(c))
	c.typeConverterRegistry.Register("tuple", types.NewTupleConverter(c))
}

// ConvertFile converts a single Terraform file to a JSON Schema
func (c *Converter) ConvertFile(filepath string) (*jsonschema.Schema, error) {
	body, err := c.parseHCLFile(filepath)
	if err != nil {
		return nil, err
	}
	return c.convertBody(body)
}

// ConvertString converts a string containing Terraform content to a JSON Schema
func (c *Converter) ConvertString(content string) (*jsonschema.Schema, error) {
	body, err := c.parseHCLString(content, "temp.tf")
	if err != nil {
		return nil, err
	}
	return c.convertBody(body)
}

// parseHCLString parses the given HCL content into an HCL body.
func (c *Converter) parseHCLString(content, filename string) (hcl.Body, error) {
	file, diags := c.parser.ParseHCL([]byte(content), filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL string: %v", diags)
	}
	if file == nil || file.Body == nil {
		return nil, fmt.Errorf("failed to parse HCL string: file or body is nil")
	}
	return file.Body, nil
}

// parseHCLFile parses the given HCL file into an HCL body.
func (c *Converter) parseHCLFile(filepath string) (hcl.Body, error) {
	file, diags := c.parser.ParseHCLFile(filepath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags)
	}
	if file == nil || file.Body == nil {
		return nil, fmt.Errorf("failed to parse HCL file: file or body is nil")
	}
	return file.Body, nil
}

// convertBody converts an HCL body to JSON Schema
func (c *Converter) convertBody(body hcl.Body) (*jsonschema.Schema, error) {
	content, diags := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to get body content: %v", diags)
	}

	rootSchema := &jsonschema.Schema{
		Schema:               "http://json-schema.org/draft-07/schema#",
		Type:                 "object",
		Properties:           make(map[string]*jsonschema.Schema),
		Required:             &[]string{},      // Always include required array (terraschema compatibility)
		AdditionalProperties: &[]bool{true}[0], // Follow terraschema's permissive approach
	}

	if err := c.processVariableBlocks(content.Blocks, rootSchema); err != nil {
		return nil, err
	}

	return rootSchema, nil
}

// processVariableBlocks processes all variable blocks and adds them to the root schema.
func (c *Converter) processVariableBlocks(blocks hcl.Blocks, rootSchema *jsonschema.Schema) error {
	for _, block := range blocks {
		if block.Type == "variable" {
			varName := block.Labels[0]
			content, diags := block.Body.Content(&hcl.BodySchema{
				Attributes: []hcl.AttributeSchema{
					{Name: "type"}, {Name: "description"}, {Name: "default"}, {Name: "sensitive"}, {Name: "nullable"},
				},
				Blocks: []hcl.BlockHeaderSchema{
					{Type: "validation"},
				},
			})
			if diags.HasErrors() {
				return fmt.Errorf("failed to get content for var '%s': %w", varName, diags)
			}

			schema, err := c.convertVariableBlock(block, content)
			if err != nil {
				return fmt.Errorf("failed to convert variable '%s': %w", varName, err)
			}
			rootSchema.Properties[varName] = schema

			// Add to required array only if variable has no default value
			if !c.hasDefaultValue(content) {
				*rootSchema.Required = append(*rootSchema.Required, varName)
			}
		}
	}

	// Sort required fields alphabetically (terraschema compatibility)
	sort.Strings(*rootSchema.Required)

	return nil
}

// hasDefaultValue checks if a variable block has a default value
func (c *Converter) hasDefaultValue(content *hcl.BodyContent) bool {
	_, exists := content.Attributes["default"]
	return exists
}

// convertVariableBlock converts a variable block to a JSON Schema
func (c *Converter) convertVariableBlock(block *hcl.Block, content *hcl.BodyContent) (*jsonschema.Schema, error) {
	schema := &jsonschema.Schema{}
	isAnyType := false

	if typeAttr, exists := content.Attributes["type"]; exists {
		if traversal, ok := typeAttr.Expr.(*hclsyntax.ScopeTraversalExpr); ok {
			if traversal.Traversal.RootName() == "any" {
				isAnyType = true
			}
		}

		var err error
		schema, err = c.ConvertType(typeAttr.Expr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert type: %w", err)
		}
	} else {
		// If no type is specified, it defaults to `any` which we can treat as an empty schema
		// waiting for type inference from the default value later.
	}

	// Apply all variable attributes (description, default, sensitive, nullable etc.)
	if err := c.attributeProcessor.ApplyAttributes(schema, content.Attributes); err != nil {
		return nil, fmt.Errorf("failed to apply variable attributes: %w", err)
	}

	// Extract and apply validation rules
	if err := c.validationProcessor.Process(schema, content.Blocks, block.Labels[0]); err != nil {
		return nil, err
	}

	// Infer type from default value if not explicitly set
	if schema.Type == "" && !isAnyType {
		if defaultValue, exists := c.parseDefault(content.Attributes); exists {
			schema = c.typeInferenceHandler.InferSchemaFromDefault(schema, defaultValue)
		}
	}

	return schema, nil
}

// ConvertType converts an HCL type expression to a JSON Schema
func (c *Converter) ConvertType(expr hcl.Expression) (*jsonschema.Schema, error) {
	var typeName string
	var converter types.TypeConverter
	var err error

	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		// Handles primitive types like `string`, `number`, `bool`, `any`
		typeName = e.Traversal.RootName()
		if typeName == "any" {
			return &jsonschema.Schema{}, nil
		}
		converter, err = c.typeConverterRegistry.Get(typeName)
		if err != nil {
			return nil, fmt.Errorf("unsupported type name: %s", typeName)
		}
		return converter.Convert(e)

	case *hclsyntax.FunctionCallExpr:
		// Handles collection types like `list(string)`, `map(any)`, `set(number)`
		// and the `tuple` type constructor itself.
		typeName = e.Name
		converter, err = c.typeConverterRegistry.Get(typeName)
		if err != nil {
			return nil, fmt.Errorf("unsupported type function: %s", typeName)
		}
		return converter.Convert(e)

	case *hclsyntax.TupleConsExpr:
		// This handles the direct tuple constructor expression, which is what the
		// tuple type function (`tuple(...)`) contains as its argument.
		converter, err = c.typeConverterRegistry.Get("tuple")
		if err != nil {
			return nil, err // Should not happen if tuple converter is registered
		}
		return converter.Convert(e)

	default:
		return nil, fmt.Errorf("unsupported type expression: %T", expr)
	}
}

// IsOptionalType checks if an expression represents an optional type
func (c *Converter) IsOptionalType(expr hcl.Expression) bool {
	if funcExpr, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
		return funcExpr.Name == "optional"
	}
	return false
}

func (c *Converter) parseDefault(attrs map[string]*hcl.Attribute) (interface{}, bool) {
	if attr, exists := attrs["default"]; exists {
		val, err := c.defaultParser.ParseDefaultValue(attr.Expr)
		if err == nil {
			return val, true
		}
	}
	return nil, false
}
