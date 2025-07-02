package validation

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLengthRuleVisitorImplementsWalker(t *testing.T) {
	var _ hclsyntax.Walker = (*lengthRuleVisitor)(nil)
}

func TestLengthRuleVisitorHandlesNodes(t *testing.T) {
	visitor := &lengthRuleVisitor{}

	// Test with a BinaryOpExpr
	binOpNode := &hclsyntax.BinaryOpExpr{}
	diags := visitor.Enter(binOpNode)
	if diags.HasErrors() {
		t.Errorf("Enter() with BinaryOpExpr returned unexpected error: %v", diags)
	}

	// Test with a FunctionCallExpr
	funcCallNode := &hclsyntax.FunctionCallExpr{
		Name: "length",
	}
	diags = visitor.Enter(funcCallNode)
	if diags.HasErrors() {
		t.Errorf("Enter() with FunctionCallExpr returned unexpected error: %v", diags)
	}

	// Test with a different node to ensure it's gracefully ignored
	otherNode := &hclsyntax.LiteralValueExpr{}
	diags = visitor.Enter(otherNode)
	if diags.HasErrors() {
		t.Errorf("Enter() with LiteralValueExpr returned unexpected error: %v", diags)
	}
}

func TestParseLengthRuleWithVisitor(t *testing.T) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(`
variable "my_string" {
  type = string
  validation {
    condition     = length(var.my_string) > 5
    error_message = "The string must be longer than 5 characters."
  }
}
`), "test.hcl")
	require.False(t, diags.HasErrors(), "unexpected diagnostics on parse")

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on content")

	require.Len(t, content.Blocks, 1, "expected one variable block")
	variableBlock := content.Blocks[0]

	valContent, diags := variableBlock.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "validation"},
		},
		Attributes: []hcl.AttributeSchema{
			{Name: "type"},
			{Name: "default"},
			{Name: "description"},
			{Name: "sensitive"},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on validation content")
	require.Len(t, valContent.Blocks, 1, "expected one validation block")
	validationBlock := valContent.Blocks[0]

	valContentBody, diags := validationBlock.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "condition"},
			{Name: "error_message"},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on validation body content")

	condAttr, ok := valContentBody.Attributes["condition"]
	require.True(t, ok, "missing condition attribute")

	rule, path, err := parseLengthRule(condAttr.Expr, "my_string")
	require.NoError(t, err)
	require.NotNil(t, rule)

	lengthRule, ok := rule.(*LengthRule)
	require.True(t, ok)

	assert.Equal(t, hclsyntax.OpGreaterThan, lengthRule.Operator)
	assert.Equal(t, 5, lengthRule.Value)
	assert.Empty(t, path)
}

func TestParseCompoundLengthRuleWithVisitor(t *testing.T) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(`
variable "my_list" {
  type = list(string)
  validation {
    condition     = length(var.my_list) >= 1 && length(var.my_list) <= 5
    error_message = "The list must have between 1 and 5 items."
  }
}
`), "test.hcl")
	require.False(t, diags.HasErrors(), "unexpected diagnostics on parse")

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on content")

	variableBlock := content.Blocks[0]
	valContent, diags := variableBlock.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "type"},
		},
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "validation"},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on validation content")

	validationBlock := valContent.Blocks[0]
	valContentBody, diags := validationBlock.Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{Name: "condition"},
			{Name: "error_message"},
		},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on validation body content")

	condAttr := valContentBody.Attributes["condition"]
	rule, path, err := parseLengthRule(condAttr.Expr, "my_list")
	require.NoError(t, err)
	require.NotNil(t, rule)

	compoundRule, ok := rule.(*CompoundLengthRule)
	require.True(t, ok)

	assert.Equal(t, 1, *compoundRule.MinValue)
	assert.Equal(t, 5, *compoundRule.MaxValue)
	assert.Empty(t, path)
}
