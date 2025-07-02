package validation

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexRuleVisitorImplementsWalker(t *testing.T) {
	var _ hclsyntax.Walker = (*regexRuleVisitor)(nil)
}

func TestRegexRuleVisitorHandlesNodes(t *testing.T) {
	visitor := &regexRuleVisitor{}

	// Test with a FunctionCallExpr
	funcCallNode := &hclsyntax.FunctionCallExpr{
		Name: "can",
	}
	diags := visitor.Enter(funcCallNode)
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

func TestParseRegexRuleWithVisitor(t *testing.T) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL([]byte(`
variable "my_string" {
  type = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9]*$", var.my_string))
    error_message = "The string must be alphanumeric."
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
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "validation"},
		},
		Attributes: []hcl.AttributeSchema{
			{Name: "type"},
			{Name: "default"},
			{Name: "description"},
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

	condAttr, ok := valContentBody.Attributes["condition"]
	require.True(t, ok, "missing condition attribute")

	rule, path, err := parseRegexRule(condAttr.Expr, "my_string")
	require.NoError(t, err)
	require.NotNil(t, rule)

	regexRule, ok := rule.(*RegexRule)
	require.True(t, ok)

	assert.Equal(t, "^[a-zA-Z0-9]*$", regexRule.Pattern)
	assert.Equal(t, []string{}, path)
}
