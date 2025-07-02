package validation

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// PathExpression represents a traversal path within a Terraform variable.
type PathExpression struct {
	Segments []PathSegment
}

// PathSegment is a single part of a PathExpression.
type PathSegment struct {
	Name     string
	IndexKey *hcl.TraverseIndex
}

// PathExpressionHandler handles extraction of property paths from complex HCL expressions
type PathExpressionHandler struct{}

// NewPathExpressionHandler creates a new path expression handler
func NewPathExpressionHandler() *PathExpressionHandler {
	return &PathExpressionHandler{}
}

// ExtractPathFromExpression extracts a property path from an expression, handling complex traversals
func (p *PathExpressionHandler) ExtractPathFromExpression(expr hcl.Expression, varName string) ([]string, error) {
	var vars []hcl.Traversal
	switch e := expr.(type) {
	case *hclsyntax.FunctionCallExpr, *hclsyntax.BinaryOpExpr, *hclsyntax.ScopeTraversalExpr:
		vars = e.Variables()
	default:
		// This can happen for literal expressions that are not paths, which is not an error.
		return nil, nil
	}

	for _, trav := range vars {
		rootName := trav.RootName()

		// Handle loop variables, where the root name matches the varName directly.
		if rootName == varName {
			return p.extractComplexPath(trav[1:]), nil
		}

		if rootName == "var" {
			// e.g., var.foo or var.foo.bar or var.foo[0].bar
			path := trav.SimpleSplit().Rel
			if len(path) > 0 {
				firstTraverser := path[0]
				if attr, ok := firstTraverser.(hcl.TraverseAttr); ok && attr.Name == varName {
					// This is a reference to the variable itself, e.g. `var.foo` inside `variable "foo"`.
					// Extract the rest of the path, a TBD.
					return p.extractComplexPath(path[1:]), nil
				}
			}
		}

		if rootName == "self" {
			// e.g., self.bar or self.array[0].prop
			return p.extractComplexPath(trav[1:]), nil
		}

		// also need to handle `each`
		if rootName == "each" {
			// each.value.name
			return p.extractComplexPath(trav[1:]), nil
		}
	}

	// This can happen for literal expressions that are not paths, which is not an error.
	return nil, nil
}

// extractComplexPath extracts a path from traversers, handling both attributes and indices
func (p *PathExpressionHandler) extractComplexPath(traversers []hcl.Traverser) []string {
	var path []string

	for _, traverser := range traversers {
		switch t := traverser.(type) {
		case hcl.TraverseAttr:
			// Regular attribute access: .property
			path = append(path, t.Name)
		case hcl.TraverseIndex:
			// Array/map access: [index] or ["key"]
			// For tuple validation, we need to capture the specific index

			// Get the actual index value if it's a number
			indexValue := ""
			if t.Key.Type().Equals(cty.Number) {
				if bigFloat := t.Key.AsBigFloat(); bigFloat != nil {
					if intVal, accuracy := bigFloat.Int64(); accuracy == 0 {
						indexValue = fmt.Sprintf("[%d]", intVal)
					}
				}
			}

			if indexValue == "" {
				// Fallback for non-numeric indices (like string keys)
				indexValue = "[*]"
			}

			// Add the index as a separate path segment
			path = append(path, indexValue)
		default:
			// For any other traversal types, we'll skip them
			// This handles cases like TraverseSplat (*) which we don't need for validation
			continue
		}
	}

	return path
}

// IsIndexedPath checks if a path contains array indexing
func (p *PathExpressionHandler) IsIndexedPath(path []string) bool {
	for _, segment := range path {
		if len(segment) > 3 && segment[len(segment)-3:] == "[*]" {
			return true
		}
	}
	return false
}

// GetBasePath removes indexing markers from a path
func (p *PathExpressionHandler) GetBasePath(path []string) []string {
	var basePath []string
	for _, segment := range path {
		if len(segment) > 3 && segment[len(segment)-3:] == "[*]" {
			basePath = append(basePath, segment[:len(segment)-3])
		} else {
			basePath = append(basePath, segment)
		}
	}
	return basePath
}
