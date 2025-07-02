package converter

// Conversion driver: parse the .tf file, walk variable blocks, assemble the root schema.

import (
	"fmt"
	"os"
	"sort"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

// ConvertFile converts a single Terraform file to a JSON Schema.
func ConvertFile(filepath string) (*jsonschema.Schema, error) {
	content, err := os.ReadFile(filepath) // #nosec G304 -- reading the file given on the command line is this tool's purpose
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	body, err := parseHCL(content, filepath)
	if err != nil {
		return nil, err
	}
	return convertBody(body)
}

// parseHCL parses HCL content into an HCL body.
func parseHCL(content []byte, filename string) (hcl.Body, error) {
	file, diags := hclparse.NewParser().ParseHCL(content, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %s", diags)
	}
	if file == nil || file.Body == nil {
		return nil, fmt.Errorf("failed to parse HCL: file or body is nil")
	}
	return file.Body, nil
}

// convertBody assembles the root schema from all variable blocks.
func convertBody(body hcl.Body) (*jsonschema.Schema, error) {
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
		Required:             &[]string{}, // always present (terraschema compatibility)
		AdditionalProperties: &jsonschema.AdditionalProperties{AllowAny: true},
	}
	return rootSchema, processVariableBlocks(content.Blocks, rootSchema)
}

// processVariableBlocks adds every variable block to the root schema.
func processVariableBlocks(blocks hcl.Blocks, rootSchema *jsonschema.Schema) error {
	for _, block := range blocks {
		if block.Type != "variable" {
			continue
		}
		varName := block.Labels[0]
		content, diags := block.Body.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "type"}, {Name: "description"}, {Name: "default"}, {Name: "sensitive"}, {Name: "nullable"},
			},
			Blocks: []hcl.BlockHeaderSchema{{Type: "validation"}},
		})
		if diags.HasErrors() {
			return fmt.Errorf("failed to get content for var '%s': %w", varName, diags)
		}
		schema, err := convertVariableBlock(content, varName)
		if err != nil {
			return fmt.Errorf("failed to convert variable '%s': %w", varName, err)
		}
		rootSchema.Properties[varName] = schema
		// A variable with a default is not required in tfvars files.
		if _, hasDefault := content.Attributes["default"]; !hasDefault {
			*rootSchema.Required = append(*rootSchema.Required, varName)
		}
	}
	sort.Strings(*rootSchema.Required) // terraschema compatibility
	return nil
}

// convertVariableBlock converts one variable block: type, then attributes,
// then validation blocks.
func convertVariableBlock(content *hcl.BodyContent, varName string) (*jsonschema.Schema, error) {
	schema := &jsonschema.Schema{}
	if typeAttr, exists := content.Attributes["type"]; exists {
		var err error
		if schema, err = ConvertTypeExpr(typeAttr.Expr); err != nil {
			return nil, fmt.Errorf("failed to convert type: %w", err)
		}
	}
	if err := applyAttributes(schema, content.Attributes); err != nil {
		return nil, err
	}
	return schema, processValidations(schema, content.Blocks, varName)
}
