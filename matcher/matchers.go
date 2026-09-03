package matcher

import (
	"fmt"
)

// Num wraps any typed numeric matcher into an AnyMatcher, applying automatic numeric coercion.
//
// Example:
//
//	.Body("price", matcher.Num(matcher.GreaterThan(10.0)))
func Num[T Ordered](m Matcher[T]) AnyMatcher {
	return &numAnyWrapper[T]{inner: m}
}

type numAnyWrapper[T Ordered] struct {
	inner Matcher[T]
}

func (w *numAnyWrapper[T]) Matches(actual any) bool {
	if w.inner == nil {
		return false
	}
	// Try type assertion first
	if val, ok := actual.(T); ok {
		return w.inner.Matches(val)
	}

	// Try numeric coercion to float64
	num, ok := toFloat64(actual)
	if !ok {
		return false
	}

	// If inner is for float64
	if mf, ok := any(w.inner).(Matcher[float64]); ok {
		return mf.Matches(num)
	}
	// If inner is for int
	if mi, ok := any(w.inner).(Matcher[int]); ok {
		return mi.Matches(int(num))
	}
	// If inner is for int64
	if mi64, ok := any(w.inner).(Matcher[int64]); ok {
		return mi64.Matches(int64(num))
	}
	// If inner is for int32
	if mi32, ok := any(w.inner).(Matcher[int32]); ok {
		return mi32.Matches(int32(num))
	}
	// If inner is for uint
	if mu, ok := any(w.inner).(Matcher[uint]); ok {
		return mu.Matches(uint(num))
	}
	// If inner is for uint64
	if mu64, ok := any(w.inner).(Matcher[uint64]); ok {
		return mu64.Matches(uint64(num))
	}
	// If inner is for float32
	if mf32, ok := any(w.inner).(Matcher[float32]); ok {
		return mf32.Matches(float32(num))
	}

	return false
}

func (w *numAnyWrapper[T]) Describe() string {
	if w.inner == nil {
		return "<nil>"
	}
	return w.inner.Describe()
}

func (w *numAnyWrapper[T]) DescribeMismatch(actual any) string {
	if w.inner == nil {
		return fmt.Sprintf("was <%v>", actual)
	}
	if val, ok := actual.(T); ok {
		return w.inner.DescribeMismatch(val)
	}
	return fmt.Sprintf("was <%v>", actual)
}

// Str wraps any typed string matcher into an AnyMatcher, formatting actual values to string.
//
// Example:
//
//	.Body("name", matcher.Str(matcher.ContainsString("John")))
func Str(m Matcher[string]) AnyMatcher {
	return &anyStringMatcherWrapper{inner: m}
}
