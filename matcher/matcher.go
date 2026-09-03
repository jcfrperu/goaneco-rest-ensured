// Package matcher provides composable, type-safe assertion matchers inspired by Hamcrest.
package matcher

import "fmt"

// Matcher is the core interface for value matching and self-description.
// T represents the type being evaluated.
type Matcher[T any] interface {
	// Matches returns true if actual satisfies this matcher's criteria.
	Matches(actual T) bool

	// Describe returns a human-readable description of what this matcher expects.
	Describe() string

	// DescribeMismatch returns a human-readable explanation of why actual did not match.
	DescribeMismatch(actual T) string
}

// AnyMatcher is a Matcher operating on any (interface{}) values.
// Used wherever the concrete type of the matched value is not known at compile time
// (such as dynamic JSON path queries).
type AnyMatcher = Matcher[any]

// FormatMismatch formats a standard two-line Hamcrest-style error message:
//
//	Expected: <matcher.Describe()>
//	     but: <matcher.DescribeMismatch(actual)>
func FormatMismatch[T any](m Matcher[T], actual T) string {
	if m == nil {
		return fmt.Sprintf("Expected: <nil matcher>\n     but: was <%v>", actual)
	}
	return fmt.Sprintf("Expected: %s\n     but: %s", m.Describe(), m.DescribeMismatch(actual))
}
