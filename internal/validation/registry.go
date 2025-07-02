package validation

import (
	"log"
	"sort"

	"github.com/hashicorp/hcl/v2"
)

// ParserFunc defines the signature for validation rule parsers.
type ParserFunc func(expr hcl.Expression, varName string) (Rule, []string, error)

// prioritizedParser holds a parser function with its priority
type prioritizedParser struct {
	parser   ParserFunc
	priority int
}

// Global registry for validation rule parsers
var parsers []prioritizedParser

// RegisterRuleParser registers a validation rule parser with default priority 0.
func RegisterRuleParser(parser ParserFunc) {
	RegisterRuleParserWithPriority(parser, 0)
}

// RegisterRuleParserWithPriority registers a validation rule parser with a specific priority.
// Higher priority parsers are executed first.
func RegisterRuleParserWithPriority(parser ParserFunc, priority int) {
	parsers = append(parsers, prioritizedParser{parser: parser, priority: priority})
	log.Printf("Registered validation rule parser with priority %d", priority)
}

// GetParsers returns all registered parsers sorted by priority (highest first).
func GetParsers() []ParserFunc {
	// Sort by priority (highest first)
	sorted := make([]prioritizedParser, len(parsers))
	copy(sorted, parsers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].priority > sorted[j].priority
	})

	// Extract just the parser functions
	result := make([]ParserFunc, len(sorted))
	for i, p := range sorted {
		result[i] = p.parser
	}
	return result
}
