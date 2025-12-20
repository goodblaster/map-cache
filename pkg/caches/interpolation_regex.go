package caches

import "regexp"

// Pre-compiled regex patterns for interpolation to avoid recompilation overhead
var (
	// InterpolationPattern matches ${{key}} syntax with optional whitespace
	InterpolationPattern = regexp.MustCompile(`\${{\s*([^}]+?)\s*}}`)

	// AggregationPattern matches any()/all() function calls with comparisons
	AggregationPattern = regexp.MustCompile(`\b(any|all)\(\s*\${{\s*([^}]+?)\s*}}\s*([!<>=]=?|==)\s*([^\)]+?)\s*\)`)
)
