package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

type containsStringMatcher struct {
	substring    string
	ignoringCase bool
}

// ContainsString matches when actual string contains substring (case-sensitive).
func ContainsString(substring string) Matcher[string] {
	return &containsStringMatcher{substring: substring, ignoringCase: false}
}

// ContainsStringIgnoringCase matches when actual string contains substring (case-insensitive).
func ContainsStringIgnoringCase(substring string) Matcher[string] {
	return &containsStringMatcher{substring: substring, ignoringCase: true}
}

func (m *containsStringMatcher) Matches(actual string) bool {
	if m.ignoringCase {
		return strings.Contains(strings.ToLower(actual), strings.ToLower(m.substring))
	}
	return strings.Contains(actual, m.substring)
}

func (m *containsStringMatcher) Describe() string {
	if m.ignoringCase {
		return fmt.Sprintf("a string containing (ignoring case) %q", m.substring)
	}
	return fmt.Sprintf("a string containing %q", m.substring)
}

func (m *containsStringMatcher) DescribeMismatch(actual string) string {
	return fmt.Sprintf("was %q", actual)
}

type startsWithMatcher struct {
	prefix string
}

// StartsWith matches when actual string begins with prefix.
func StartsWith(prefix string) Matcher[string] {
	return &startsWithMatcher{prefix: prefix}
}

func (m *startsWithMatcher) Matches(actual string) bool {
	return strings.HasPrefix(actual, m.prefix)
}

func (m *startsWithMatcher) Describe() string {
	return fmt.Sprintf("a string starting with %q", m.prefix)
}

func (m *startsWithMatcher) DescribeMismatch(actual string) string {
	return fmt.Sprintf("was %q", actual)
}

type endsWithMatcher struct {
	suffix string
}

// EndsWith matches when actual string ends with suffix.
func EndsWith(suffix string) Matcher[string] {
	return &endsWithMatcher{suffix: suffix}
}

func (m *endsWithMatcher) Matches(actual string) bool {
	return strings.HasSuffix(actual, m.suffix)
}

func (m *endsWithMatcher) Describe() string {
	return fmt.Sprintf("a string ending with %q", m.suffix)
}

func (m *endsWithMatcher) DescribeMismatch(actual string) string {
	return fmt.Sprintf("was %q", actual)
}

type matchesRegexMatcher struct {
	pattern string
	re      *regexp.Regexp
	compErr error
}

// MatchesRegex matches when actual string matches the regular expression pattern.
// An invalid pattern does not panic — it produces a matcher that always fails and reports
// the compile error via Describe() and DescribeMismatch().
func MatchesRegex(pattern string) Matcher[string] {
	re, err := regexp.Compile(pattern)
	return &matchesRegexMatcher{pattern: pattern, re: re, compErr: err}
}

func (m *matchesRegexMatcher) Matches(actual string) bool {
	if m.compErr != nil {
		return false
	}
	return m.re.MatchString(actual)
}

func (m *matchesRegexMatcher) Describe() string {
	if m.compErr != nil {
		return fmt.Sprintf("invalid regex %q: %v", m.pattern, m.compErr)
	}
	return fmt.Sprintf("a string matching pattern %q", m.pattern)
}

func (m *matchesRegexMatcher) DescribeMismatch(actual string) string {
	if m.compErr != nil {
		return fmt.Sprintf("invalid regex %q: %v", m.pattern, m.compErr)
	}
	return fmt.Sprintf("was %q", actual)
}

type equalToIgnoringCaseMatcher struct {
	expected string
}

// EqualToIgnoringCase matches when actual equals expected, ignoring case distinctions.
func EqualToIgnoringCase(expected string) Matcher[string] {
	return &equalToIgnoringCaseMatcher{expected: expected}
}

func (m *equalToIgnoringCaseMatcher) Matches(actual string) bool {
	return strings.EqualFold(actual, m.expected)
}

func (m *equalToIgnoringCaseMatcher) Describe() string {
	return fmt.Sprintf("equalToIgnoringCase(%q)", m.expected)
}

func (m *equalToIgnoringCaseMatcher) DescribeMismatch(actual string) string {
	return fmt.Sprintf("was %q", actual)
}

type emptyStringMatcher struct{}

// EmptyString matches empty strings ("").
func EmptyString() Matcher[string] {
	return &emptyStringMatcher{}
}

func (m *emptyStringMatcher) Matches(actual string) bool {
	return actual == ""
}

func (m *emptyStringMatcher) Describe() string {
	return "an empty string"
}

func (m *emptyStringMatcher) DescribeMismatch(actual string) string {
	return fmt.Sprintf("was %q", actual)
}

type emptyOrNullStringMatcher struct{}

// EmptyOrNullString matches "" or nil.
func EmptyOrNullString() Matcher[any] {
	return &emptyOrNullStringMatcher{}
}

func (m *emptyOrNullStringMatcher) Matches(actual any) bool {
	if actual == nil {
		return true
	}
	if s, ok := actual.(string); ok {
		return s == ""
	}
	return false
}

func (m *emptyOrNullStringMatcher) Describe() string {
	return "an empty or null string"
}

func (m *emptyOrNullStringMatcher) DescribeMismatch(actual any) string {
	if actual == nil {
		return "was <nil>"
	}
	return fmt.Sprintf("was <%v>", actual)
}

// --- AnyMatcher wrappers for dynamic/JSON string values ---

type anyStringMatcherWrapper struct {
	inner Matcher[string]
}

func (w *anyStringMatcherWrapper) Matches(actual any) bool {
	if actual == nil {
		return false
	}
	return w.inner.Matches(fmt.Sprintf("%v", actual))
}

func (w *anyStringMatcherWrapper) Describe() string {
	return w.inner.Describe()
}

func (w *anyStringMatcherWrapper) DescribeMismatch(actual any) string {
	if actual == nil {
		return "was <nil>"
	}
	return w.inner.DescribeMismatch(fmt.Sprintf("%v", actual))
}

// ContainsStringAny creates an AnyMatcher from ContainsString.
func ContainsStringAny(substring string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: ContainsString(substring)}
}

// ContainsStringIgnoringCaseAny creates an AnyMatcher from ContainsStringIgnoringCase.
func ContainsStringIgnoringCaseAny(substring string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: ContainsStringIgnoringCase(substring)}
}

// StartsWithAny creates an AnyMatcher from StartsWith.
func StartsWithAny(prefix string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: StartsWith(prefix)}
}

// EndsWithAny creates an AnyMatcher from EndsWith.
func EndsWithAny(suffix string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: EndsWith(suffix)}
}

// MatchesRegexAny creates an AnyMatcher from MatchesRegex.
func MatchesRegexAny(pattern string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: MatchesRegex(pattern)}
}

// EqualToIgnoringCaseAny creates an AnyMatcher from EqualToIgnoringCase.
func EqualToIgnoringCaseAny(expected string) AnyMatcher {
	return &anyStringMatcherWrapper{inner: EqualToIgnoringCase(expected)}
}
