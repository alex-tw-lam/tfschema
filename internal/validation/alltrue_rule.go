package validation

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func init() {
	RegisterRuleParserWithPriority(parseAllTrueRule, 20)
}

func parseAllTrueRule(expr hcl.Expression, varName string) (Rule, []string, error) {
	log.Printf("[alltrue] called for var=%s, expr=%T", varName, expr)
	call, ok := expr.(*hclsyntax.FunctionCallExpr)
	if !ok || call.Name != "alltrue" {
		return nil, nil, nil // Not an alltrue() call.
	}

	if len(call.Args) != 1 {
		return nil, nil, nil
	}

	log.Printf("[alltrue] call.Name=%s argType=%T", call.Name, call.Args[0])

	// The argument can be a tuple expression wrapping the ForExpr or the ForExpr directly
	var forExpr *hclsyntax.ForExpr
	if tuple, ok := call.Args[0].(*hclsyntax.TupleConsExpr); ok {
		if len(tuple.Exprs) != 1 {
			return nil, nil, nil
		}
		forExpr, ok = tuple.Exprs[0].(*hclsyntax.ForExpr)
		if !ok {
			return nil, nil, nil
		}
	} else if fe, ok := call.Args[0].(*hclsyntax.ForExpr); ok {
		forExpr = fe
	} else {
		return nil, nil, nil
	}

	innerExpr := forExpr.ValExpr

	// Unwrap any parentheses around the inner expression for simpler parsing.
	if paren, ok := innerExpr.(*hclsyntax.ParenthesesExpr); ok {
		innerExpr = paren.Expression
	}

	collectionPath, err := pathHandler.ExtractPathFromExpression(forExpr.CollExpr, varName)
	if err != nil {
		return nil, nil, fmt.Errorf("could not extract collection path from for expression: %w", err)
	}

	// Get a list of all parsers except for this one to avoid recursion.
	otherParsers := make([]ParserFunc, 0)
	for _, p := range GetParsers() {
		if !isThisParser(p, parseAllTrueRule) {
			otherParsers = append(otherParsers, p)
		}
	}

	for _, parser := range otherParsers {
		log.Printf("[alltrue] trying subparser %p (%T)", parser, parser)
		rule, innerPath, err := parser(innerExpr, forExpr.ValVar)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse inner expression in alltrue: %w", err)
		}
		if rule != nil {
			log.Printf("[alltrue] subparser produced rule %#v with innerPath=%v", rule, innerPath)
		}
		if rule != nil {
			fullPath := append(collectionPath, "*")
			fullPath = append(fullPath, innerPath...)
			log.Printf("[alltrue] returning rule for fullPath=%v", fullPath)
			return rule, fullPath, nil
		}
	}
	log.Printf("[alltrue] no subparser matched inner expression for loopVar=%s", forExpr.ValVar)

	return nil, nil, nil
}

func isThisParser(p ParserFunc, self ParserFunc) bool {
	return reflect.ValueOf(p).Pointer() == reflect.ValueOf(self).Pointer()
}
