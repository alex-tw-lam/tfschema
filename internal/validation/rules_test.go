package validation

import (
	"io"
	"os"
	"testing"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseCondition extracts the first `condition = ...` expression of a
// single-variable Terraform snippet and parses it with the given rule parser.
func parseCondition(t *testing.T, source, varName string, parse ParserFunc) (Rule, []string) {
	t.Helper()

	file, diags := hclparse.NewParser().ParseHCL([]byte(source), "test.hcl")
	require.False(t, diags.HasErrors(), "unexpected diagnostics on parse")

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variable", LabelNames: []string{"name"}}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on content")

	varContent, diags := content.Blocks[0].Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{{Name: "type"}, {Name: "default"}, {Name: "description"}},
		Blocks:     []hcl.BlockHeaderSchema{{Type: "validation"}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on variable content")

	validationContent, diags := varContent.Blocks[0].Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{{Name: "condition"}, {Name: "error_message"}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on validation content")

	rule, path, err := parse(validationContent.Attributes["condition"].Expr, varName)
	require.NoError(t, err)
	require.NotNil(t, rule)
	return rule, path
}

// applyTo stamps a rule onto a fresh schema of the given type.
func applyTo(t *testing.T, rule Rule, typeName string) *jsonschema.Schema {
	t.Helper()
	schema := &jsonschema.Schema{Type: typeName}
	rule.Apply(schema)
	return schema
}

func TestParseLengthRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_string" {
  type = string
  validation {
    condition     = length(var.my_string) > 5
    error_message = "The string must be longer than 5 characters."
  }
}
`, "my_string", parseLengthRule)

	schema := applyTo(t, rule, "string")
	assert.Equal(t, 6, *schema.MinLength) // > 5 means at least 6
	assert.Nil(t, schema.MaxLength)
}

func TestParseCompoundLengthRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_list" {
  type = list(string)
  validation {
    condition     = length(var.my_list) >= 1 && length(var.my_list) <= 5
    error_message = "The list must have between 1 and 5 items."
  }
}
`, "my_list", parseLengthRule)

	schema := applyTo(t, rule, "array")
	assert.Equal(t, 1, *schema.MinItems)
	assert.Equal(t, 5, *schema.MaxItems)
}

func TestParseRegexRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_string" {
  type = string
  validation {
    condition     = can(regex("^[a-zA-Z0-9]*$", var.my_string))
    error_message = "The string must be alphanumeric."
  }
}
`, "my_string", parseRegexRule)

	schema := applyTo(t, rule, "string")
	assert.Equal(t, "^[a-zA-Z0-9]*$", schema.Pattern)
}

func TestParseEnumRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_env" {
  type = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.my_env)
    error_message = "Unknown environment."
  }
}
`, "my_env", parseEnumRule)

	schema := applyTo(t, rule, "string")
	assert.Equal(t, []interface{}{"dev", "staging", "prod"}, schema.Enum)
}

func TestParseRangeRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_port" {
  type = number
  validation {
    condition     = var.my_port >= 1 && var.my_port <= 65535
    error_message = "Port must be between 1 and 65535."
  }
}
`, "my_port", parseRangeRule)

	schema := applyTo(t, rule, "number")
	assert.Equal(t, float64(1), *schema.Minimum)
	assert.Equal(t, float64(65535), *schema.Maximum)
}

func TestParseOrChainRule(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_env" {
  type = string
  validation {
    condition     = var.my_env == "dev" || var.my_env == "prod"
    error_message = "dev or prod only."
  }
}
`, "my_env", parseOrChainRule)

	schema := applyTo(t, rule, "string")
	assert.Equal(t, []interface{}{"dev", "prod"}, schema.Enum)
}

// TestMergeRangeRules pins the merge semantics: a bound set by both sides
// takes the right side's value; a bound set by one side survives.
func TestMergeRangeRules(t *testing.T) {
	zero, one, five, ten := float64(0), float64(1), float64(5), float64(10)

	// e.g. `var.x >= 10 && var.x >= 1`: the right side wins.
	merged := mergeRangeRules(&RangeRule{Minimum: &ten}, &RangeRule{Minimum: &one})
	assert.Equal(t, one, *merged.Minimum)

	// e.g. `var.x > 0 && var.x <= 5`: disjoint bounds combine.
	merged = mergeRangeRules(&RangeRule{ExclusiveMinimum: &zero}, &RangeRule{Maximum: &five})
	assert.Equal(t, zero, *merged.ExclusiveMinimum)
	assert.Equal(t, five, *merged.Maximum)
	assert.Nil(t, merged.Minimum)
}

// TestParseRangeRuleReversed covers value-first comparisons (3 < var.x).
func TestParseRangeRuleReversed(t *testing.T) {
	rule, _ := parseCondition(t, `
variable "my_number" {
  type = number
  validation {
    condition     = 3 < var.my_number
    error_message = "Must be greater than 3."
  }
}
`, "my_number", parseRangeRule)

	schema := applyTo(t, rule, "number")
	assert.Equal(t, float64(3), *schema.ExclusiveMinimum)
}

func TestParseAllTrueRule(t *testing.T) {
	rule, path := parseCondition(t, `
variable "my_numbers" {
  type = list(number)
  validation {
    condition     = alltrue([for n in var.my_numbers : n > 0])
    error_message = "All numbers must be positive."
  }
}
`, "my_numbers", parseAllTrueRule)

	// The "*" wildcard marks "apply to every element of the collection".
	assert.Equal(t, []string{"*"}, path)

	schema := applyTo(t, rule, "number")
	assert.Equal(t, float64(0), *schema.ExclusiveMinimum)
}

// TestUnmatchedConditionFallsThrough checks that a condition no parser
// recognises yields no rule and prints a warning to stderr.
func TestUnmatchedConditionFallsThrough(t *testing.T) {
	file, diags := hclparse.NewParser().ParseHCL([]byte(`
variable "my_string" {
  type = string
  validation {
    condition     = upper(var.my_string) == "HELLO"
    error_message = "No parser understands upper()."
  }
}
`), "test.hcl")
	require.False(t, diags.HasErrors(), "unexpected diagnostics on parse")

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "variable", LabelNames: []string{"name"}}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on content")

	varContent, diags := content.Blocks[0].Body.Content(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{{Name: "type"}},
		Blocks:     []hcl.BlockHeaderSchema{{Type: "validation"}},
	})
	require.False(t, diags.HasErrors(), "unexpected diagnostics on variable content")

	// Capture stderr while the rules are extracted.
	stderr := os.Stderr
	pipeReader, pipeWriter, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = pipeWriter

	rules, err := ExtractValidationRules(varContent.Blocks, "my_string")

	pipeWriter.Close()
	os.Stderr = stderr
	captured, _ := io.ReadAll(pipeReader)

	require.NoError(t, err)
	assert.Empty(t, rules)
	assert.Contains(t, string(captured), "warning: variable \"my_string\"")
}
