package validation

import (
	"sort"

	"github.com/hashicorp/hcl/v2"
)

// ParserFunc defines the signature for validation rule parsers.
type ParserFunc func(expr hcl.Expression, varName string) (Rule, []string, error)

// prioritizedParser holds a parser function with its priority and an optional
// name. The name lets composite parsers (such as alltrue) identify and exclude
// themselves when iterating the registry, without resorting to reflect-based
// function-pointer comparison.
type prioritizedParser struct {
	name     string
	parser   ParserFunc
	priority int
}

// Global registry for validation rule parsers
var parsers []prioritizedParser

// RegisterRuleParser registers a validation rule parser with default priority 0.
func RegisterRuleParser(parser ParserFunc) {
	RegisterRuleParserWithName("", parser, 0)
}

// RegisterRuleParserWithPriority registers a validation rule parser with a specific priority.
// Higher priority parsers are executed first.
func RegisterRuleParserWithPriority(parser ParserFunc, priority int) {
	RegisterRuleParserWithName("", parser, priority)
}

// RegisterRuleParserWithName registers a named validation rule parser. The name
// is optional (pass "" for unnamed leaf parsers); named parsers can be excluded
// via GetParsersExcept, which is how composite parsers avoid recursing into
// themselves.
func RegisterRuleParserWithName(name string, parser ParserFunc, priority int) {
	parsers = append(parsers, prioritizedParser{name: name, parser: parser, priority: priority})
}

// sortedParsers returns all registered parsers sorted by priority (highest first).
func sortedParsers() []prioritizedParser {
	sorted := make([]prioritizedParser, len(parsers))
	copy(sorted, parsers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].priority > sorted[j].priority
	})
	return sorted
}

// GetParsers returns all registered parsers sorted by priority (highest first).
func GetParsers() []ParserFunc {
	sorted := sortedParsers()
	result := make([]ParserFunc, len(sorted))
	for i, p := range sorted {
		result[i] = p.parser
	}
	return result
}

// GetParsersExcept returns all registered parsers sorted by priority (highest
// first), excluding any parser registered with the given name. Composite parsers
// use this to iterate their candidate sub-parsers without matching themselves.
func GetParsersExcept(name string) []ParserFunc {
	sorted := sortedParsers()
	result := make([]ParserFunc, 0, len(sorted))
	for _, p := range sorted {
		if p.name == name {
			continue
		}
		result = append(result, p.parser)
	}
	return result
}
