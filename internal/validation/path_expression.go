package validation

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// extractPathFromExpression returns the property path of the first reference
// to the variable in the expression, e.g. ["settings", "port"] for
// var.settings.port. Literal-only expressions return nil, which is not an error.
func extractPathFromExpression(expr hcl.Expression, varName string) ([]string, error) {
	for _, trav := range expr.Variables() {
		rootName := trav.RootName()
		if rootName == varName {
			// Bare variable name: loop variable inside a for-expression.
			return toPathSegments(trav[1:]), nil
		}
		if rootName == "var" {
			// var.<name>... reference: keep the path after the variable name.
			rel := trav.SimpleSplit().Rel
			if len(rel) > 0 {
				if attr, ok := rel[0].(hcl.TraverseAttr); ok && attr.Name == varName {
					return toPathSegments(rel[1:]), nil
				}
			}
		}
		if rootName == "self" || rootName == "each" {
			// self.x / each.value.x reference: keep the path after the root.
			return toPathSegments(trav[1:]), nil
		}
	}
	return nil, nil // no reference to this variable
}

// toPathSegments turns HCL traversers into path segments: attribute names stay
// as-is, numeric indices become "[N]", string keys become the "[*]" wildcard.
func toPathSegments(traversers []hcl.Traverser) []string {
	var path []string
	for _, traverser := range traversers {
		switch t := traverser.(type) {
		case hcl.TraverseAttr:
			path = append(path, t.Name)
		case hcl.TraverseIndex:
			if t.Key.Type().Equals(cty.Number) {
				if intVal, accuracy := t.Key.AsBigFloat().Int64(); accuracy == 0 {
					path = append(path, fmt.Sprintf("[%d]", intVal))
					continue
				}
			}
			path = append(path, "[*]")
		}
	}
	return path
}
