package matcher

import (
	"fmt"
	"reflect"
)

// equalToMatcher implements equality comparison with coercion fallback.
type equalToMatcher[T any] struct {
	expected T
}

// EqualTo creates a matcher that matches when actual equals expected.
// For interface/any types, applies numeric coercion as a fallback after reflect.DeepEqual fails.
// Note: numeric strings are coerced to float64, so EqualTo[any]("123") matches int(123).
// Use typed matchers (EqualTo[string]) when you need to distinguish the string "123" from 123.
func EqualTo[T any](expected T) Matcher[T] {
	return &equalToMatcher[T]{expected: expected}
}

func (m *equalToMatcher[T]) Matches(actual T) bool {
	if reflect.DeepEqual(actual, m.expected) {
		return true
	}

	normAct, normExp, ok := toComparable(actual, m.expected)
	if ok {
		if reflect.DeepEqual(normAct, normExp) {
			return true
		}
	}
	return false
}

func (m *equalToMatcher[T]) Describe() string {
	return fmt.Sprintf("<equal to %v>", m.expected)
}

func (m *equalToMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

// Is is syntactic sugar for EqualTo(value).
func Is[T any](value T) Matcher[T] {
	return EqualTo(value)
}

// IsMatcher returns the given matcher unchanged, providing readable fluent syntax.
func IsMatcher[T any](m Matcher[T]) Matcher[T] {
	return m
}

// notMatcher inverts the outcome of an inner matcher.
type notMatcher[T any] struct {
	inner Matcher[T]
}

// Not creates a matcher that matches when the inner matcher does not match.
func Not[T any](m Matcher[T]) Matcher[T] {
	return &notMatcher[T]{inner: m}
}

// NotValue creates a matcher that matches when actual is not equal to value.
func NotValue[T any](value T) Matcher[T] {
	return Not(EqualTo(value))
}

func (m *notMatcher[T]) Matches(actual T) bool {
	if m.inner == nil {
		return false
	}
	return !m.inner.Matches(actual)
}

func (m *notMatcher[T]) Describe() string {
	if m.inner == nil {
		return "not (<nil>)"
	}
	return fmt.Sprintf("not (%s)", m.inner.Describe())
}

func (m *notMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

// anythingMatcher matches any value unconditionally.
type anythingMatcher struct {
	description string
}

// Anything creates a matcher that always returns true.
// Optionally accepts a custom description string.
func Anything(desc ...string) Matcher[any] {
	d := "ANYTHING"
	if len(desc) > 0 && desc[0] != "" {
		d = desc[0]
	}
	return &anythingMatcher{description: d}
}

func (m *anythingMatcher) Matches(_ any) bool {
	return true
}

func (m *anythingMatcher) Describe() string {
	return m.description
}

func (m *anythingMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

// nullValueMatcher matches nil values.
type nullValueMatcher struct{}

// NullValue creates a matcher that matches nil or null values.
func NullValue() Matcher[any] {
	return &nullValueMatcher{}
}

func (m *nullValueMatcher) Matches(actual any) bool {
	if actual == nil {
		return true
	}
	rv := reflect.ValueOf(actual)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func (m *nullValueMatcher) Describe() string {
	return "nil"
}

func (m *nullValueMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

// notNullValueMatcher matches non-nil values.
type notNullValueMatcher struct{}

// NotNullValue creates a matcher that matches non-nil values.
func NotNullValue() Matcher[any] {
	return &notNullValueMatcher{}
}

func (m *notNullValueMatcher) Matches(actual any) bool {
	if actual == nil {
		return false
	}
	rv := reflect.ValueOf(actual)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !rv.IsNil()
	default:
		return true
	}
}

func (m *notNullValueMatcher) Describe() string {
	return "not nil"
}

func (m *notNullValueMatcher) DescribeMismatch(actual any) string {
	if actual == nil {
		return "was <nil>"
	}
	return fmt.Sprintf("was <%v>", actual)
}
