package converter

import (
	"testing"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// variableAttrs parses a single-variable Terraform snippet and returns the
// variable block's attributes.
func variableAttrs(t *testing.T, source string) map[string]*hcl.Attribute {
	t.Helper()

	file, diags := hclparse.NewParser().ParseHCL([]byte(source), "test.hcl")
	require.False(t, diags.HasErrors(), "unexpected diagnostics on parse")

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variable", LabelNames: []string{"name"}}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on content")

	varContent, diags := content.Blocks[0].Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "type"}, {Name: "description"}, {Name: "default"}, {Name: "sensitive"}, {Name: "nullable"},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on variable content")
	return varContent.Attributes
}

func TestApplyAttributesBasics(t *testing.T) {
	attrs := variableAttrs(t, `
variable "my_string" {
  type        = string
  default     = "abc"
  description = "A test variable."
}
`)
	schema := &jsonschema.Schema{Type: "string"}
	require.NoError(t, applyAttributes(schema, attrs))

	assert.Equal(t, "A test variable.", schema.Description)
	assert.Equal(t, "abc", schema.Default)
	assert.Nil(t, schema.Sensitive) // not marked sensitive
}

func TestApplyAttributesSensitive(t *testing.T) {
	attrs := variableAttrs(t, `
variable "my_secret" {
  type      = string
  sensitive = true
}
`)
	schema := &jsonschema.Schema{Type: "string"}
	require.NoError(t, applyAttributes(schema, attrs))

	require.NotNil(t, schema.Sensitive)
	assert.True(t, *schema.Sensitive)
}

// TestApplyAttributesNullable pins the terraschema-compatible anyOf rewrite.
func TestApplyAttributesNullable(t *testing.T) {
	attrs := variableAttrs(t, `
variable "my_nullable" {
  type     = string
  nullable = true
}
`)
	schema := &jsonschema.Schema{Type: "string"}
	require.NoError(t, applyAttributes(schema, attrs))

	assert.Empty(t, schema.Type) // replaced by the anyOf choice
	assert.Equal(t, "Select a type", schema.Title)
	require.Len(t, schema.AnyOf, 2)
	assert.Equal(t, "null", schema.AnyOf[0].Type)
	assert.Equal(t, "string", schema.AnyOf[1].Type)
}

// TestFindTargetSchema navigates a schema the way validation rules do:
// property names, "[N]" tuple indices and "*" wildcards.
func TestFindTargetSchema(t *testing.T) {
	port := &jsonschema.Schema{Type: "number"}
	stringItem := &jsonschema.Schema{Type: "string"}
	numberItem := &jsonschema.Schema{Type: "number"}
	variableSchema := &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"port":  port,
			"tuple": {Type: "array", Items: &jsonschema.Items{Tuple: []*jsonschema.Schema{stringItem, numberItem}}},
			"tags":  {Type: "array", Items: &jsonschema.Items{Single: &jsonschema.Schema{Type: "string"}}},
			"meta":  {Type: "object", AdditionalProperties: &jsonschema.AdditionalProperties{ValueSchema: &jsonschema.Schema{Type: "string"}}},
		},
	}

	target, err := findTargetSchema("my_var", variableSchema, []string{"port"})
	require.NoError(t, err)
	assert.Same(t, port, target)

	target, err = findTargetSchema("my_var", variableSchema, []string{"tuple", "[1]"})
	require.NoError(t, err)
	assert.Same(t, numberItem, target) // tuple positions have their own schemas

	target, err = findTargetSchema("my_var", variableSchema, []string{"tags", "*"})
	require.NoError(t, err)
	assert.Equal(t, "string", target.Type) // wildcard lands on the item schema

	target, err = findTargetSchema("my_var", variableSchema, []string{"meta", "*"})
	require.NoError(t, err)
	assert.Equal(t, "string", target.Type) // map wildcard lands on the value schema

	_, err = findTargetSchema("my_var", variableSchema, []string{"missing"})
	assert.Error(t, err) // unknown property names are an error, not silence
}
