package validation

import (
	"fmt"
	"os"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
)

// Rule is a validation constraint that stamps itself onto a JSON Schema.
type Rule interface {
	Apply(schema *jsonschema.Schema)
}

// ParserFunc parses a condition into a Rule plus its property path;
// a nil Rule means "not recognised by this parser".
type ParserFunc func(expr hcl.Expression, varName string) (Rule, []string, error)

// ScopedRule pairs a Rule with the path to the property it applies to.
type ScopedRule struct {
	Rule Rule
	Path []string
}

// ruleParsers are tried in order until one recognises a condition;
// add new rule kinds to this list.
var ruleParsers = []ParserFunc{
	parseAllTrueRule,
	parseLengthRule,
	parseRegexRule,
	parseEnumRule,
	parseOrChainRule,
	parseRangeRule,
}

// ExtractValidationRules extracts the validation rules from a variable's blocks.
func ExtractValidationRules(blocks hcl.Blocks, varName string) ([]ScopedRule, error) {
	var rules []ScopedRule
	for _, block := range blocks {
		if block.Type != "validation" {
			continue
		}
		content, diags := block.Body.Content(&hcl.BodySchema{Attributes: []hcl.AttributeSchema{{Name: "condition"}, {Name: "error_message"}}})
		condition := content.Attributes["condition"]
		if diags.HasErrors() || condition == nil {
			continue
		}
		matched := false
		for _, parse := range ruleParsers {
			rule, path, err := parse(condition.Expr, varName)
			if err != nil {
				return nil, err
			}
			if rule != nil {
				rules = append(rules, ScopedRule{Rule: rule, Path: path})
				matched = true
				break
			}
		}
		if !matched {
			// A skipped condition makes the generated schema validate less
			// than the .tf file demands, so emit a warning instead of ignoring it.
			condRange := condition.Expr.Range()
			fmt.Fprintf(os.Stderr,
				"tfschema: warning: variable %q: skipped validation condition at %s:%d (no rule parser understands it)\n",
				varName, condRange.Filename, condRange.Start.Line)
		}
	}
	return rules, nil
}
