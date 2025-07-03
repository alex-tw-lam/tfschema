package validation

import (
	"fmt"

	"github.com/alex-tw-lam/tfschema/internal/jsonschema"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func init() {
	RegisterRuleParserWithPriority(parseEnumRule, 0)
}

// EnumRule represents an enum validation rule.
type EnumRule struct {
	Values []interface{}
}

// Apply applies the enum validation rule to a JSON schema.
func (r *EnumRule) Apply(schema *jsonschema.Schema) error {
	schema.Enum = r.Values
	return nil
}

func parseEnumRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "contains" {
		return nil, nil, nil // Not a 'contains' function call.
	}

	if len(call.Args) != 2 {
		return nil, nil, fmt.Errorf("'contains' function for enum validation must have two arguments")
	}

	path, err := pathHandler.ExtractPathFromExpression(call.Args[1], varName)
	if err != nil {
		return nil, nil, err
	}

	listExpr, ok := call.Args[0].(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, nil, fmt.Errorf("first argument to 'contains' for enum validation must be a literal list")
	}

	var values []interface{}
	for _, itemExpr := range listExpr.Exprs {
		val, diags := itemExpr.Value(nil)
		if diags.HasErrors() {
			return nil, nil, fmt.Errorf("failed to evaluate enum value: %s", diags.Error())
		}
		switch val.Type() {
		case cty.String:
			values = append(values, val.AsString())
		case cty.Number:
			bf, _ := val.AsBigFloat().Float64()
			values = append(values, bf)
		case cty.Bool:
			values = append(values, val.True())
		default:
			return nil, nil, fmt.Errorf("unsupported type in enum validation: %s", val.Type().FriendlyName())
		}
	}

	rule := &EnumRule{Values: values}
	return rule, path, nil
}
