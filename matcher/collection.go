package matcher

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

type hasItemMatcher struct {
	itemMatcher AnyMatcher
}

// HasItem matches when a collection contains at least one element satisfying the matcher.
func HasItem(m AnyMatcher) AnyMatcher {
	return &hasItemMatcher{itemMatcher: m}
}

// HasItemValue matches when a collection contains at least one element equal to value.
func HasItemValue(value any) AnyMatcher {
	return HasItem(EqualTo[any](value))
}

func (m *hasItemMatcher) Matches(actual any) bool {
	slice, ok := toAnySlice(actual)
	if !ok {
		return false
	}
	for _, item := range slice {
		if m.itemMatcher.Matches(item) {
			return true
		}
	}
	return false
}

func (m *hasItemMatcher) Describe() string {
	return fmt.Sprintf("a collection containing %s", m.itemMatcher.Describe())
}

func (m *hasItemMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

type hasItemsMatcher struct {
	matchers []AnyMatcher
}

// HasItems matches when a collection contains ALL of the specified values in any order.
func HasItems(values ...any) AnyMatcher {
	matchers := make([]AnyMatcher, len(values))
	for i, val := range values {
		matchers[i] = EqualTo[any](val)
	}
	return HasItemsMatching(matchers...)
}

// HasItemsMatching matches when a collection contains elements satisfying all provided matchers.
func HasItemsMatching(matchers ...AnyMatcher) AnyMatcher {
	return &hasItemsMatcher{matchers: matchers}
}

func (m *hasItemsMatcher) Matches(actual any) bool {
	slice, ok := toAnySlice(actual)
	if !ok {
		return false
	}
	for _, matcher := range m.matchers {
		matched := false
		for _, item := range slice {
			if matcher.Matches(item) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (m *hasItemsMatcher) Describe() string {
	parts := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		parts[i] = matcher.Describe()
	}
	return fmt.Sprintf("a collection containing items [%s]", strings.Join(parts, ", "))
}

func (m *hasItemsMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

type hasSizeMatcher struct {
	sizeMatcher Matcher[int]
}

// HasSize matches when a collection, string, or map has length equal to size.
func HasSize(size int) AnyMatcher {
	return HasSizeMatcher(EqualTo(size))
}

// HasSizeMatcher matches when a collection, string, or map satisfies an int matcher.
func HasSizeMatcher(m Matcher[int]) AnyMatcher {
	return &hasSizeMatcher{sizeMatcher: m}
}

func (m *hasSizeMatcher) Matches(actual any) bool {
	if actual == nil {
		return m.sizeMatcher.Matches(0)
	}
	if str, ok := actual.(string); ok {
		return m.sizeMatcher.Matches(utf8.RuneCountInString(str))
	}
	if mObj, ok := actual.(map[string]any); ok {
		return m.sizeMatcher.Matches(len(mObj))
	}
	if slice, ok := toAnySlice(actual); ok {
		return m.sizeMatcher.Matches(len(slice))
	}
	rv := reflect.ValueOf(actual)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return m.sizeMatcher.Matches(rv.Len())
	}
	return false
}

func (m *hasSizeMatcher) Describe() string {
	return fmt.Sprintf("a collection with size %s", m.sizeMatcher.Describe())
}

func (m *hasSizeMatcher) DescribeMismatch(actual any) string {
	if slice, ok := toAnySlice(actual); ok {
		return fmt.Sprintf("collection size was <%d>", len(slice))
	}
	return fmt.Sprintf("was <%v>", actual)
}

type containsMatcher struct {
	matchers []AnyMatcher
}

// Contains matches when a collection contains exactly these values in strict order.
func Contains(values ...any) AnyMatcher {
	matchers := make([]AnyMatcher, len(values))
	for i, v := range values {
		matchers[i] = EqualTo[any](v)
	}
	return ContainsMatchers(matchers...)
}

// ContainsMatchers matches when a collection elements satisfy matchers in strict order.
func ContainsMatchers(matchers ...AnyMatcher) AnyMatcher {
	return &containsMatcher{matchers: matchers}
}

func (m *containsMatcher) Matches(actual any) bool {
	slice, ok := toAnySlice(actual)
	if !ok || len(slice) != len(m.matchers) {
		return false
	}
	for i, matcher := range m.matchers {
		if !matcher.Matches(slice[i]) {
			return false
		}
	}
	return true
}

func (m *containsMatcher) Describe() string {
	parts := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		parts[i] = matcher.Describe()
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

func (m *containsMatcher) DescribeMismatch(actual any) string {
	slice, ok := toAnySlice(actual)
	if !ok {
		return fmt.Sprintf("was not a collection <%v>", actual)
	}
	if len(slice) != len(m.matchers) {
		return fmt.Sprintf("collection size was <%d> instead of <%d>", len(slice), len(m.matchers))
	}
	for i, matcher := range m.matchers {
		if !matcher.Matches(slice[i]) {
			return fmt.Sprintf("item at index %d: %s", i, matcher.DescribeMismatch(slice[i]))
		}
	}
	return fmt.Sprintf("was <%v>", actual)
}

type containsInAnyOrderMatcher struct {
	matchers []AnyMatcher
}

// ContainsInAnyOrder matches when a collection contains exactly these values in any order.
func ContainsInAnyOrder(values ...any) AnyMatcher {
	matchers := make([]AnyMatcher, len(values))
	for i, v := range values {
		matchers[i] = EqualTo[any](v)
	}
	return ContainsInAnyOrderMatchers(matchers...)
}

// ContainsInAnyOrderMatchers matches when collection elements satisfy matchers in any order.
func ContainsInAnyOrderMatchers(matchers ...AnyMatcher) AnyMatcher {
	return &containsInAnyOrderMatcher{matchers: matchers}
}

func (m *containsInAnyOrderMatcher) Matches(actual any) bool {
	slice, ok := toAnySlice(actual)
	if !ok || len(slice) != len(m.matchers) {
		return false
	}

	if len(slice) > maxBacktrackSize {
		return matchInAnyOrderGreedy(m.matchers, slice)
	}
	used := make([]bool, len(slice))
	return matchInAnyOrder(m.matchers, slice, 0, used)
}

// maxBacktrackSize is the largest N for which the exact O(N!) backtracking bijection
// search is used. Collections larger than this threshold fall back to the O(N²) greedy
// algorithm, which trades theoretical completeness for bounded runtime. In practice the
// greedy pass is correct for all collections where matchers do not heavily overlap.
const maxBacktrackSize = 15

// matchInAnyOrder uses backtracking to find a bijection between matchers and items.
// Each matcher must be satisfied by a distinct item; the used bitmap tracks which items
// have been claimed so that one item cannot satisfy two matchers simultaneously.
func matchInAnyOrder(matchers []AnyMatcher, items []any, matcherIdx int, used []bool) bool {
	if matcherIdx == len(matchers) {
		return true
	}
	currentMatcher := matchers[matcherIdx]
	for i, item := range items {
		if !used[i] && currentMatcher.Matches(item) {
			used[i] = true
			if matchInAnyOrder(matchers, items, matcherIdx+1, used) {
				return true
			}
			used[i] = false
		}
	}
	return false
}

// matchInAnyOrderGreedy is an O(N²) fallback for large collections.
// It assigns each matcher to the first unclaimed item that satisfies it.
// This may miss valid bijections when matchers heavily overlap, but prevents O(N!) blowup.
func matchInAnyOrderGreedy(matchers []AnyMatcher, items []any) bool {
	used := make([]bool, len(items))
	for _, m := range matchers {
		found := false
		for i, item := range items {
			if !used[i] && m.Matches(item) {
				used[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (m *containsInAnyOrderMatcher) Describe() string {
	parts := make([]string, len(m.matchers))
	for i, matcher := range m.matchers {
		parts[i] = matcher.Describe()
	}
	return fmt.Sprintf("a collection containing [%s] in any order", strings.Join(parts, ", "))
}

func (m *containsInAnyOrderMatcher) DescribeMismatch(actual any) string {
	slice, ok := toAnySlice(actual)
	if !ok {
		return fmt.Sprintf("was not a collection <%v>", actual)
	}
	if len(slice) != len(m.matchers) {
		return fmt.Sprintf("collection size was <%d> instead of <%d>", len(slice), len(m.matchers))
	}
	return fmt.Sprintf("was <%v>", actual)
}

type emptyMatcher struct{}

// Empty matches an empty collection, map, or string.
func Empty() AnyMatcher {
	return &emptyMatcher{}
}

func (m *emptyMatcher) Matches(actual any) bool {
	if actual == nil {
		return true
	}
	if str, ok := actual.(string); ok {
		return str == ""
	}
	if mObj, ok := actual.(map[string]any); ok {
		return len(mObj) == 0
	}
	if slice, ok := toAnySlice(actual); ok {
		return len(slice) == 0
	}
	rv := reflect.ValueOf(actual)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String, reflect.Chan:
		return rv.Len() == 0
	}
	return false
}

func (m *emptyMatcher) Describe() string {
	return "an empty collection"
}

func (m *emptyMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

type everyItemMatcher struct {
	itemMatcher AnyMatcher
}

// EveryItem matches if every element in the collection satisfies the matcher.
func EveryItem(m AnyMatcher) AnyMatcher {
	return &everyItemMatcher{itemMatcher: m}
}

func (m *everyItemMatcher) Matches(actual any) bool {
	slice, ok := toAnySlice(actual)
	if !ok {
		return false
	}
	for _, item := range slice {
		if !m.itemMatcher.Matches(item) {
			return false
		}
	}
	return true
}

func (m *everyItemMatcher) Describe() string {
	return fmt.Sprintf("every item is %s", m.itemMatcher.Describe())
}

func (m *everyItemMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}
