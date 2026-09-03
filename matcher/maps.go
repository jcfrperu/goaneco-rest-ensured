package matcher

import (
	"fmt"
	"reflect"
)

// toMapStringAny normalizes map types into map[string]any.
func toMapStringAny(v any) (map[string]any, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]any); ok {
		return m, true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Map {
		return nil, false
	}

	result := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		result[fmt.Sprintf("%v", iter.Key().Interface())] = iter.Value().Interface()
	}
	return result, true
}

type hasEntryMatcher struct {
	keyMatcher   Matcher[string]
	valueMatcher AnyMatcher
}

// HasEntry matches if a map contains a key-value entry satisfying both matchers.
func HasEntry(keyMatcher Matcher[string], valueMatcher AnyMatcher) AnyMatcher {
	return &hasEntryMatcher{keyMatcher: keyMatcher, valueMatcher: valueMatcher}
}

// HasEntryValues matches if a map contains the exact key-value pair.
func HasEntryValues(key string, value any) AnyMatcher {
	return HasEntry(EqualTo(key), EqualTo[any](value))
}

func (m *hasEntryMatcher) Matches(actual any) bool {
	mObj, ok := toMapStringAny(actual)
	if !ok {
		return false
	}
	for k, v := range mObj {
		if m.keyMatcher.Matches(k) && m.valueMatcher.Matches(v) {
			return true
		}
	}
	return false
}

func (m *hasEntryMatcher) Describe() string {
	return fmt.Sprintf("a map containing {%s: %s}", m.keyMatcher.Describe(), m.valueMatcher.Describe())
}

func (m *hasEntryMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

type hasKeyMatcher struct {
	keyMatcher Matcher[string]
}

// HasKey matches if a map contains a key satisfying the string matcher.
func HasKey(m Matcher[string]) AnyMatcher {
	return &hasKeyMatcher{keyMatcher: m}
}

// HasKeyValue matches if a map contains the specified key string.
func HasKeyValue(key string) AnyMatcher {
	return HasKey(EqualTo(key))
}

func (m *hasKeyMatcher) Matches(actual any) bool {
	mObj, ok := toMapStringAny(actual)
	if !ok {
		return false
	}
	for k := range mObj {
		if m.keyMatcher.Matches(k) {
			return true
		}
	}
	return false
}

func (m *hasKeyMatcher) Describe() string {
	return fmt.Sprintf("a map containing key %s", m.keyMatcher.Describe())
}

func (m *hasKeyMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}

type hasValueMatcher struct {
	valueMatcher AnyMatcher
}

// HasValue matches if a map contains at least one value satisfying the matcher.
func HasValue(m AnyMatcher) AnyMatcher {
	return &hasValueMatcher{valueMatcher: m}
}

// HasValueValue matches if a map contains at least one value equal to value.
func HasValueValue(value any) AnyMatcher {
	return HasValue(EqualTo[any](value))
}

func (m *hasValueMatcher) Matches(actual any) bool {
	mObj, ok := toMapStringAny(actual)
	if !ok {
		return false
	}
	for _, v := range mObj {
		if m.valueMatcher.Matches(v) {
			return true
		}
	}
	return false
}

func (m *hasValueMatcher) Describe() string {
	return fmt.Sprintf("a map containing value %s", m.valueMatcher.Describe())
}

func (m *hasValueMatcher) DescribeMismatch(actual any) string {
	return fmt.Sprintf("was <%v>", actual)
}
