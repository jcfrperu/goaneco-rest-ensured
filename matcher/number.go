package matcher

import (
	"fmt"
	"math"
)

// Ordered represents all types that support the <, <=, >, >= comparison operators.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type greaterThanMatcher[T Ordered] struct {
	threshold T
}

// GreaterThan matches when actual > threshold.
func GreaterThan[T Ordered](threshold T) Matcher[T] {
	return &greaterThanMatcher[T]{threshold: threshold}
}

func (m *greaterThanMatcher[T]) Matches(actual T) bool {
	return actual > m.threshold
}

func (m *greaterThanMatcher[T]) Describe() string {
	return fmt.Sprintf("a value greater than <%v>", m.threshold)
}

func (m *greaterThanMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

type greaterThanOrEqualToMatcher[T Ordered] struct {
	threshold T
}

// GreaterThanOrEqualTo matches when actual >= threshold.
func GreaterThanOrEqualTo[T Ordered](threshold T) Matcher[T] {
	return &greaterThanOrEqualToMatcher[T]{threshold: threshold}
}

func (m *greaterThanOrEqualToMatcher[T]) Matches(actual T) bool {
	return actual >= m.threshold
}

func (m *greaterThanOrEqualToMatcher[T]) Describe() string {
	return fmt.Sprintf("a value greater than or equal to <%v>", m.threshold)
}

func (m *greaterThanOrEqualToMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

type lessThanMatcher[T Ordered] struct {
	threshold T
}

// LessThan matches when actual < threshold.
func LessThan[T Ordered](threshold T) Matcher[T] {
	return &lessThanMatcher[T]{threshold: threshold}
}

func (m *lessThanMatcher[T]) Matches(actual T) bool {
	return actual < m.threshold
}

func (m *lessThanMatcher[T]) Describe() string {
	return fmt.Sprintf("a value less than <%v>", m.threshold)
}

func (m *lessThanMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

type lessThanOrEqualToMatcher[T Ordered] struct {
	threshold T
}

// LessThanOrEqualTo matches when actual <= threshold.
func LessThanOrEqualTo[T Ordered](threshold T) Matcher[T] {
	return &lessThanOrEqualToMatcher[T]{threshold: threshold}
}

func (m *lessThanOrEqualToMatcher[T]) Matches(actual T) bool {
	return actual <= m.threshold
}

func (m *lessThanOrEqualToMatcher[T]) Describe() string {
	return fmt.Sprintf("a value less than or equal to <%v>", m.threshold)
}

func (m *lessThanOrEqualToMatcher[T]) DescribeMismatch(actual T) string {
	return fmt.Sprintf("was <%v>", actual)
}

type closeToMatcher struct {
	target float64
	delta  float64
}

// CloseTo matches when |actual - target| <= delta for float64 values.
func CloseTo(target, delta float64) Matcher[float64] {
	return &closeToMatcher{target: target, delta: math.Abs(delta)}
}

func (m *closeToMatcher) Matches(actual float64) bool {
	return math.Abs(actual-m.target) <= m.delta
}

func (m *closeToMatcher) Describe() string {
	return fmt.Sprintf("a numeric value within <%v> of <%v>", m.delta, m.target)
}

func (m *closeToMatcher) DescribeMismatch(actual float64) string {
	return fmt.Sprintf("<%v> differed by <%v>", actual, math.Abs(actual-m.target))
}

// --- AnyMatcher wrappers for untyped JSON evaluation ---

type anyNumericMatcher struct {
	threshold float64
	operator  string // ">", ">=", "<", "<="
}

func (m *anyNumericMatcher) Matches(actual any) bool {
	actVal, ok := toFloat64(actual)
	if !ok {
		return false
	}
	switch m.operator {
	case ">":
		return actVal > m.threshold
	case ">=":
		return actVal >= m.threshold
	case "<":
		return actVal < m.threshold
	case "<=":
		return actVal <= m.threshold
	}
	return false
}

func (m *anyNumericMatcher) Describe() string {
	switch m.operator {
	case ">":
		return fmt.Sprintf("a value greater than <%v>", m.threshold)
	case ">=":
		return fmt.Sprintf("a value greater than or equal to <%v>", m.threshold)
	case "<":
		return fmt.Sprintf("a value less than <%v>", m.threshold)
	case "<=":
		return fmt.Sprintf("a value less than or equal to <%v>", m.threshold)
	}
	return "numeric comparison"
}

func (m *anyNumericMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

// GreaterThanNum creates an AnyMatcher for numeric comparisons (coercing JSON numbers).
// Panics if threshold is not a numeric type.
func GreaterThanNum(threshold any) AnyMatcher {
	th, ok := toFloat64(threshold)
	if !ok {
		panic(fmt.Sprintf("matcher: GreaterThanNum requires a numeric threshold, got %T", threshold))
	}
	return &anyNumericMatcher{threshold: th, operator: ">"}
}

// GreaterThanOrEqualToNum creates an AnyMatcher for numeric comparisons.
// Panics if threshold is not a numeric type.
func GreaterThanOrEqualToNum(threshold any) AnyMatcher {
	th, ok := toFloat64(threshold)
	if !ok {
		panic(fmt.Sprintf("matcher: GreaterThanOrEqualToNum requires a numeric threshold, got %T", threshold))
	}
	return &anyNumericMatcher{threshold: th, operator: ">="}
}

// LessThanNum creates an AnyMatcher for numeric comparisons.
// Panics if threshold is not a numeric type.
func LessThanNum(threshold any) AnyMatcher {
	th, ok := toFloat64(threshold)
	if !ok {
		panic(fmt.Sprintf("matcher: LessThanNum requires a numeric threshold, got %T", threshold))
	}
	return &anyNumericMatcher{threshold: th, operator: "<"}
}

// LessThanOrEqualToNum creates an AnyMatcher for numeric comparisons.
// Panics if threshold is not a numeric type.
func LessThanOrEqualToNum(threshold any) AnyMatcher {
	th, ok := toFloat64(threshold)
	if !ok {
		panic(fmt.Sprintf("matcher: LessThanOrEqualToNum requires a numeric threshold, got %T", threshold))
	}
	return &anyNumericMatcher{threshold: th, operator: "<="}
}

type anyCloseToMatcher struct {
	target float64
	delta  float64
}

// CloseToNum creates an AnyMatcher for close-to floating point comparison.
func CloseToNum(target, delta any) AnyMatcher {
	t, _ := toFloat64(target)
	d, _ := toFloat64(delta)
	return &anyCloseToMatcher{target: t, delta: d}
}

func (m *anyCloseToMatcher) Matches(actual any) bool {
	actVal, ok := toFloat64(actual)
	if !ok {
		return false
	}
	return math.Abs(actVal-m.target) <= m.delta
}

func (m *anyCloseToMatcher) Describe() string {
	return fmt.Sprintf("a numeric value within <%v> of <%v>", m.delta, m.target)
}

func (m *anyCloseToMatcher) DescribeMismatch(actual any) string {
	actVal, ok := toFloat64(actual)
	if !ok {
		return fmt.Sprintf("was non-numeric <%v>", actual)
	}
	return fmt.Sprintf("<%v> differed by <%v>", actVal, math.Abs(actVal-m.target))
}
