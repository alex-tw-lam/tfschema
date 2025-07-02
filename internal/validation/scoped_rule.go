package validation

// ScopedRule pairs a Rule with the path to the property it applies to.
type ScopedRule struct {
	Rule Rule
	Path []string
}
