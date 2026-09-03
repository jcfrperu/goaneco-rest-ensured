package matcher

import (
	"fmt"
	"strings"
)

// allOfMatcher combines multiple matchers with logical AND (all must match).
type allOfMatcher[T any] struct {
	matchers []Matcher[T]
}

// AllOf creates a matcher that matches when ALL supplied matchers match.
// It short-circuits evaluation on the first failing matcher.
func AllOf[T any](matchers ...Matcher[T]) Matcher[T] {
	return &allOfMatcher[T]{matchers: matchers}
}

func (m *allOfMatcher[T]) Matches(actual T) bool {
	for _, matcher := range m.matchers {
		if matcher != nil && !matcher.Matches(actual) {
			return false
		}
	}
	return true
}

func (m *allOfMatcher[T]) Describe() string {
	parts := make([]string, 0, len(m.matchers))
	for _, matcher := range m.matchers {
		if matcher != nil {
			parts = append(parts, matcher.Describe())
		}
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " and "))
}

func (m *allOfMatcher[T]) DescribeMismatch(actual T) string {
	for _, matcher := range m.matchers {
		if matcher != nil && !matcher.Matches(actual) {
			return matcher.DescribeMismatch(actual)
		}
	}
	return fmt.Sprintf("was <%v>", actual)
}

// anyOfMatcher combines multiple matchers with logical OR (at least one must match).
type anyOfMatcher[T any] struct {
	matchers []Matcher[T]
}

// AnyOf creates a matcher that matches when ANY of the supplied matchers match.
// It short-circuits evaluation on the first matching matcher.
func AnyOf[T any](matchers ...Matcher[T]) Matcher[T] {
	return &anyOfMatcher[T]{matchers: matchers}
}

func (m *anyOfMatcher[T]) Matches(actual T) bool {
	if len(m.matchers) == 0 {
		return false
	}
	for _, matcher := range m.matchers {
		if matcher != nil && matcher.Matches(actual) {
			return true
		}
	}
	return false
}

func (m *anyOfMatcher[T]) Describe() string {
	parts := make([]string, 0, len(m.matchers))
	for _, matcher := range m.matchers {
		if matcher != nil {
			parts = append(parts, matcher.Describe())
		}
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, " or "))
}

func (m *anyOfMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

// CombinableMatcher allows chaining matchers using .Or() and .And().
type CombinableMatcher[T any] struct {
	matchers []Matcher[T]
	mode     string // "and" or "or"
}

// Either begins an either/or chain.
func Either[T any](m Matcher[T]) *CombinableMatcher[T] {
	return &CombinableMatcher[T]{
		matchers: []Matcher[T]{m},
		mode:     "or",
	}
}

// Both begins an and chain.
func Both[T any](m Matcher[T]) *CombinableMatcher[T] {
	return &CombinableMatcher[T]{
		matchers: []Matcher[T]{m},
		mode:     "and",
	}
}

// Or appends another matcher using OR logic.
// Panics if the chain was started with Both/And — mixing And and Or in one chain is not allowed.
// Use AnyOf/AllOf for complex compound expressions.
func (cm *CombinableMatcher[T]) Or(m Matcher[T]) *CombinableMatcher[T] {
	if cm.mode == "and" {
		panic("matcher: cannot call Or() on a Both/And chain — mixing And and Or is not allowed; use AnyOf/AllOf instead")
	}
	cm.matchers = append(cm.matchers, m)
	return cm
}

// And appends another matcher using AND logic.
// Panics if the chain was started with Either/Or — mixing Or and And in one chain is not allowed.
// Use AnyOf/AllOf for complex compound expressions.
func (cm *CombinableMatcher[T]) And(m Matcher[T]) *CombinableMatcher[T] {
	if cm.mode == "or" {
		panic("matcher: cannot call And() on an Either/Or chain — mixing Or and And is not allowed; use AnyOf/AllOf instead")
	}
	cm.matchers = append(cm.matchers, m)
	return cm
}

func (cm *CombinableMatcher[T]) Matches(actual T) bool {
	if cm.mode == "and" {
		for _, m := range cm.matchers {
			if m != nil && !m.Matches(actual) {
				return false
			}
		}
		return true
	}

	// "or" mode
	for _, m := range cm.matchers {
		if m != nil && m.Matches(actual) {
			return true
		}
	}
	return false
}

func (cm *CombinableMatcher[T]) Describe() string {
	parts := make([]string, 0, len(cm.matchers))
	for _, m := range cm.matchers {
		if m != nil {
			parts = append(parts, m.Describe())
		}
	}
	sep := " or "
	if cm.mode == "and" {
		sep = " and "
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, sep))
}

func (cm *CombinableMatcher[T]) DescribeMismatch(actual T) string {
	if cm.mode == "and" {
		for _, m := range cm.matchers {
			if m != nil && !m.Matches(actual) {
				return m.DescribeMismatch(actual)
			}
		}
	}
	return fmt.Sprintf("was <%v>", actual)
}
