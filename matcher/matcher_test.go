package matcher_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/matcher"
)

func TestEqualTo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// int
	is.True(matcher.EqualTo(4).Matches(4))
	is.False(matcher.EqualTo(4).Matches(5))

	// string
	is.True(matcher.EqualTo("hello").Matches("hello"))
	is.False(matcher.EqualTo("hello").Matches("world"))

	// float64
	is.True(matcher.EqualTo(3.14).Matches(3.14))

	// Describe format
	is.Equal("<equal to 4>", matcher.EqualTo(4).Describe())

	// DescribeMismatch format
	is.Equal("was <5>", matcher.EqualTo(4).DescribeMismatch(5))
}

func TestEqualTo_CoercedFromJSON(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// JSON number arrives as float64(4) but user supplies int(4) — must match
	m := matcher.EqualTo[any](4)
	is.True(m.Matches(float64(4))) // gjson returns float64
	is.False(m.Matches(float64(5)))

	// json.Number
	is.True(m.Matches(json.Number("4")))
	is.False(m.Matches(json.Number("5")))

	// bool
	mBool := matcher.EqualTo[any](true)
	is.True(mBool.Matches(true))
	is.False(mBool.Matches(false))

	// string
	mStr := matcher.EqualTo[any]("abc")
	is.True(mStr.Matches("abc"))
	is.False(mStr.Matches("xyz"))

	// slices
	mSlice := matcher.EqualTo[any]([]any{1, 2})
	is.True(mSlice.Matches([]any{1, 2}))
	is.False(mSlice.Matches([]any{1, 3}))

	// maps
	mMap := matcher.EqualTo[any](map[string]any{"a": 1})
	is.True(mMap.Matches(map[string]any{"a": 1}))
	is.False(mMap.Matches(map[string]any{"a": 2}))
}

func TestNot(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Not(matcher.EqualTo(4))
	is.True(m.Matches(5))
	is.False(m.Matches(4))
	is.Contains(m.Describe(), "not")
	is.Equal("was <4>", m.DescribeMismatch(4))

	notVal := matcher.NotValue(4)
	is.True(notVal.Matches(5))
	is.False(notVal.Matches(4))

	// Nil safety
	var nilM matcher.Matcher[int]
	nilNot := matcher.Not(nilM)
	is.False(nilNot.Matches(4))
	is.Equal("not (<nil>)", nilNot.Describe())
}

func TestIs(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.True(matcher.Is(4).Matches(4))
	is.False(matcher.Is(4).Matches(5))

	m := matcher.IsMatcher(matcher.EqualTo(10))
	is.True(m.Matches(10))
	is.False(m.Matches(11))
}

func TestNullValue(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.True(matcher.NullValue().Matches(nil))
	is.False(matcher.NullValue().Matches("something"))
	is.Equal("nil", matcher.NullValue().Describe())
	is.Equal("was <abc>", matcher.NullValue().DescribeMismatch("abc"))

	is.True(matcher.NotNullValue().Matches("something"))
	is.False(matcher.NotNullValue().Matches(nil))
	is.Equal("not nil", matcher.NotNullValue().Describe())
	is.Equal("was <nil>", matcher.NotNullValue().DescribeMismatch(nil))
	is.Equal("was <123>", matcher.NotNullValue().DescribeMismatch(123))

	var ptr *int
	is.True(matcher.NullValue().Matches(ptr))
	is.False(matcher.NotNullValue().Matches(ptr))

	val := 10
	ptr = &val
	is.False(matcher.NullValue().Matches(ptr))
	is.True(matcher.NotNullValue().Matches(ptr))
}

func TestAnything(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Anything()
	is.True(m.Matches(nil))
	is.True(m.Matches(42))
	is.True(m.Matches(""))
	is.True(m.Matches([]any{1, 2, 3}))
	is.Equal("ANYTHING", m.Describe())
	is.Equal("was <42>", m.DescribeMismatch(42))

	custom := matcher.Anything("custom desc")
	is.Equal("custom desc", custom.Describe())
}

func TestAllOf(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.AllOf(matcher.GreaterThan(0), matcher.LessThan(10))
	is.True(m.Matches(5))
	is.False(m.Matches(0))
	is.False(m.Matches(10))
	is.False(m.Matches(-1))
	is.Contains(m.Describe(), "and")
	is.Contains(m.DescribeMismatch(-1), "was <-1>")
}

func TestAnyOf(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.AnyOf(matcher.EqualTo(1), matcher.EqualTo(2), matcher.EqualTo(3))
	is.True(m.Matches(1))
	is.True(m.Matches(2))
	is.True(m.Matches(3))
	is.False(m.Matches(4))
	is.Contains(m.Describe(), "or")
	is.Equal("was <4>", m.DescribeMismatch(4))

	emptyAnyOf := matcher.AnyOf[int]()
	is.False(emptyAnyOf.Matches(1))
}

func TestEither_Or(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Either(matcher.EqualTo("reference")).Or(matcher.EqualTo("fiction"))
	is.True(m.Matches("reference"))
	is.True(m.Matches("fiction"))
	is.False(m.Matches("mystery"))
	is.Contains(m.Describe(), "or")
	is.Equal("was <mystery>", m.DescribeMismatch("mystery"))
}

func TestBoth_And(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Both(matcher.GreaterThan(0)).And(matcher.LessThan(10))
	is.True(m.Matches(5))
	is.False(m.Matches(-1))
	is.False(m.Matches(15))
	is.Contains(m.Describe(), "and")
	is.Equal("was <15>", m.DescribeMismatch(15))
}

func TestGreaterThan(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.GreaterThan(5)
	is.True(m.Matches(6))
	is.False(m.Matches(5))
	is.False(m.Matches(4))
	is.Contains(m.Describe(), "greater than")
	is.Equal("was <4>", m.DescribeMismatch(4))
}

func TestLessThan(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.LessThan(5)
	is.True(m.Matches(4))
	is.False(m.Matches(5))
	is.False(m.Matches(6))
	is.Contains(m.Describe(), "less than")
	is.Equal("was <6>", m.DescribeMismatch(6))
}

func TestGreaterThanOrEqualTo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.GreaterThanOrEqualTo(5)
	is.True(m.Matches(5))
	is.True(m.Matches(6))
	is.False(m.Matches(4))
	is.Contains(m.Describe(), "greater than or equal to")
	is.Equal("was <4>", m.DescribeMismatch(4))
}

func TestLessThanOrEqualTo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.LessThanOrEqualTo(5)
	is.True(m.Matches(5))
	is.True(m.Matches(4))
	is.False(m.Matches(6))
	is.Contains(m.Describe(), "less than or equal to")
	is.Equal("was <6>", m.DescribeMismatch(6))
}

func TestCloseTo(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.CloseTo(3.14, 0.01)
	is.True(m.Matches(3.14))
	is.True(m.Matches(3.145))
	is.False(m.Matches(3.16))
	is.Contains(m.Describe(), "within")
	is.Contains(m.DescribeMismatch(3.16), "differed by")

	// CloseToNum AnyMatcher
	numClose := matcher.CloseToNum(10, 2)
	is.True(numClose.Matches(11))
	is.True(numClose.Matches(12))
	is.False(numClose.Matches(13))
	is.False(numClose.Matches("not-a-number"))
	is.Contains(numClose.Describe(), "within")
	is.Contains(numClose.DescribeMismatch("not-a-number"), "non-numeric")
	is.Contains(numClose.DescribeMismatch(14), "differed by")
}

func TestGreaterThanNum_WithJSONFloat(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// gjson returns float64 for all numbers
	m := matcher.GreaterThanNum(2) // threshold is int
	is.True(m.Matches(float64(3)))
	is.False(m.Matches(float64(2)))
	is.False(m.Matches(float64(1)))
	is.False(m.Matches("non-numeric"))
	is.Equal("a value greater than <2>", m.Describe())
	is.Equal("was <1>", m.DescribeMismatch(1))

	gte := matcher.GreaterThanOrEqualToNum(2)
	is.True(gte.Matches(float64(2)))
	is.True(gte.Matches(float64(3)))
	is.False(gte.Matches(float64(1)))
	is.Equal("a value greater than or equal to <2>", gte.Describe())

	lt := matcher.LessThanNum(5)
	is.True(lt.Matches(float64(4)))
	is.False(lt.Matches(float64(5)))
	is.Equal("a value less than <5>", lt.Describe())

	lte := matcher.LessThanOrEqualToNum(5)
	is.True(lte.Matches(float64(5)))
	is.False(lte.Matches(float64(6)))
	is.Equal("a value less than or equal to <5>", lte.Describe())
}

func TestContainsString(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.ContainsString("foo")
	is.True(m.Matches("foobar"))
	is.False(m.Matches("baz"))
	is.Contains(m.Describe(), "foo")
	is.Equal(`was "baz"`, m.DescribeMismatch("baz"))
}

func TestContainsStringIgnoringCase(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.ContainsStringIgnoringCase("FOO")
	is.True(m.Matches("foobar"))
	is.True(m.Matches("FOOBAR"))
	is.Contains(m.Describe(), "(ignoring case)")
	is.Equal(`was "baz"`, m.DescribeMismatch("baz"))
}

func TestStartsWith(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.StartsWith("foo")
	is.True(m.Matches("foobar"))
	is.False(m.Matches("barfoo"))
	is.Equal(`a string starting with "foo"`, m.Describe())
	is.Equal(`was "barfoo"`, m.DescribeMismatch("barfoo"))
}

func TestEndsWith(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EndsWith("bar")
	is.True(m.Matches("foobar"))
	is.False(m.Matches("barfoo"))
	is.Equal(`a string ending with "bar"`, m.Describe())
	is.Equal(`was "barfoo"`, m.DescribeMismatch("barfoo"))
}

func TestMatchesRegex(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.MatchesRegex(`^\d{3}-\d{4}$`)
	is.True(m.Matches("123-4567"))
	is.False(m.Matches("abc"))
	is.Equal(`a string matching pattern "^\\d{3}-\\d{4}$"`, m.Describe())
	is.Equal(`was "abc"`, m.DescribeMismatch("abc"))
}

func TestEqualToIgnoringCase(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EqualToIgnoringCase("HELLO")
	is.True(m.Matches("hello"))
	is.True(m.Matches("HELLO"))
	is.False(m.Matches("world"))
	is.Equal(`equalToIgnoringCase("HELLO")`, m.Describe())
	is.Equal(`was "world"`, m.DescribeMismatch("world"))
}

func TestEmptyString(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EmptyString()
	is.True(m.Matches(""))
	is.False(m.Matches("x"))
	is.Equal("an empty string", m.Describe())
	is.Equal(`was "x"`, m.DescribeMismatch("x"))
}

func TestEmptyOrNullString(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EmptyOrNullString()
	is.True(m.Matches(nil))
	is.True(m.Matches(""))
	is.False(m.Matches("x"))
	is.False(m.Matches(123))
	is.Equal("an empty or null string", m.Describe())
	is.Equal("was <nil>", m.DescribeMismatch(nil))
	is.Equal("was <x>", m.DescribeMismatch("x"))
}

func TestStringAnyWrappers(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	cont := matcher.ContainsStringAny("bar")
	is.True(cont.Matches("foobar"))
	is.False(cont.Matches(nil))
	is.Equal("was <nil>", cont.DescribeMismatch(nil))

	is.True(matcher.ContainsStringIgnoringCaseAny("BAR").Matches("foobar"))
	is.True(matcher.StartsWithAny("foo").Matches("foobar"))
	is.True(matcher.EndsWithAny("bar").Matches("foobar"))
	is.True(matcher.MatchesRegexAny(`^\d+$`).Matches(12345))
	is.True(matcher.EqualToIgnoringCaseAny("GO").Matches("go"))
}

func TestHasItem(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasItemValue(45)
	is.True(m.Matches([]any{float64(23), float64(45), float64(54)}))
	is.False(m.Matches([]any{float64(1), float64(2)}))
	is.False(m.Matches("not-a-slice"))
	is.Equal("a collection containing <equal to 45>", m.Describe())
	is.Equal("was <[1 2]>", m.DescribeMismatch([]any{1, 2}))

	itemMatcher := matcher.HasItem(matcher.GreaterThanNum(50))
	is.True(itemMatcher.Matches([]any{10, 20, 60}))
	is.False(itemMatcher.Matches([]any{10, 20, 30}))
}

func TestHasItems(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasItems(23, 54)
	is.True(m.Matches([]any{float64(23), float64(54), float64(99)}))
	is.False(m.Matches([]any{float64(23)})) // missing 54
	is.False(m.Matches(123))
	is.Equal("a collection containing items [<equal to 23>, <equal to 54>]", m.Describe())
	is.Equal("was <[23]>", m.DescribeMismatch([]any{23}))
}

func TestHasSize(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasSize(3)
	is.True(m.Matches([]any{1, 2, 3}))
	is.False(m.Matches([]any{1, 2}))
	is.True(m.Matches("abc")) // string length
	is.True(m.Matches(map[string]any{"a": 1, "b": 2, "c": 3}))
	is.Equal("a collection with size <equal to 3>", m.Describe())
	is.Equal("collection size was <2>", m.DescribeMismatch([]any{1, 2}))
	is.Equal("was <abc>", m.DescribeMismatch("abc"))

	// HasSizeMatcher
	sizeGt := matcher.HasSizeMatcher(matcher.GreaterThan(2))
	is.True(sizeGt.Matches([]int{1, 2, 3, 4}))
	is.False(sizeGt.Matches([]int{1}))
	is.False(sizeGt.Matches(nil))
}

func TestContains(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Contains("a", "b", "c")
	is.True(m.Matches([]any{"a", "b", "c"}))
	is.False(m.Matches([]any{"a", "c", "b"})) // order matters
	is.False(m.Matches([]any{"a", "b"}))      // missing c
	is.False(m.Matches("not-a-slice"))
	is.Equal("[<equal to a>, <equal to b>, <equal to c>]", m.Describe())
	is.Contains(m.DescribeMismatch("not-a-slice"), "was not a collection")
	is.Contains(m.DescribeMismatch([]any{"a", "b"}), "collection size was <2> instead of <3>")
	is.Contains(m.DescribeMismatch([]any{"a", "c", "b"}), "item at index 1")

	// nil vs empty slice
	is.False(m.Matches(nil))
	is.False(m.Matches([]any{}))
	is.True(matcher.Contains().Matches([]any{}))
	is.False(matcher.Contains().Matches(nil))
}

func TestContainsInAnyOrder(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.ContainsInAnyOrder("a", "b", "c")
	is.True(m.Matches([]any{"c", "a", "b"})) // any order
	is.True(m.Matches([]any{"a", "b", "c"}))
	is.False(m.Matches([]any{"a", "b"}))           // missing c
	is.False(m.Matches([]any{"a", "b", "c", "d"})) // extra element
	is.False(m.Matches("not-a-slice"))
	is.Equal("a collection containing [<equal to a>, <equal to b>, <equal to c>] in any order", m.Describe())
	is.Contains(m.DescribeMismatch("not-a-slice"), "was not a collection")
	is.Contains(m.DescribeMismatch([]any{"a", "b"}), "collection size was <2> instead of <3>")

	// nil vs empty slice
	is.False(m.Matches(nil))
	is.False(m.Matches([]any{}))
	is.True(matcher.ContainsInAnyOrder().Matches([]any{}))
	is.False(matcher.ContainsInAnyOrder().Matches(nil))
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.Empty()
	is.True(m.Matches([]any{}))
	is.True(m.Matches(""))
	is.False(m.Matches([]any{1}))
	is.True(m.Matches(map[string]any{}))
	is.False(m.Matches(map[string]any{"key": "val"}))
	is.True(m.Matches(nil))
	is.Equal("an empty collection", m.Describe())
	is.Equal("was <[1]>", m.DescribeMismatch([]any{1}))
}

func TestEveryItem(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EveryItem(matcher.ContainsStringAny("go"))
	is.True(m.Matches([]any{"go", "golang", "go-test"}))
	is.False(m.Matches([]any{"go", "java"}))
	is.False(m.Matches("not-a-slice"))
	is.Equal(`every item is a string containing "go"`, m.Describe())
	is.Equal("was <[go java]>", m.DescribeMismatch([]any{"go", "java"}))
}

func TestHasEntry(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasEntryValues("category", "reference")
	obj := map[string]any{"category": "reference", "price": float64(8.95)}
	is.True(m.Matches(obj))
	is.False(m.Matches(map[string]any{"category": "fiction"}))
	is.False(m.Matches(nil))
	is.False(m.Matches("not-a-map"))
	is.Equal("a map containing {<equal to category>: <equal to reference>}", m.Describe())
	is.Equal("was <map[]>", m.DescribeMismatch(map[string]any{}))

	// Map with non-string keys via reflection
	reflectMap := map[int]string{1: "first", 2: "second"}
	reflectMatcher := matcher.HasEntry(matcher.EqualTo("1"), matcher.EqualTo[any]("first"))
	is.True(reflectMatcher.Matches(reflectMap))
}

func TestHasKey(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasKeyValue("category")
	is.True(m.Matches(map[string]any{"category": "reference"}))
	is.False(m.Matches(map[string]any{"price": float64(8.95)}))
	is.False(m.Matches(nil))
	is.Equal("a map containing key <equal to category>", m.Describe())
	is.Equal("was <map[]>", m.DescribeMismatch(map[string]any{}))
}

func TestHasValue(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.HasValueValue("reference")
	is.True(m.Matches(map[string]any{"category": "reference"}))
	is.False(m.Matches(map[string]any{"category": "fiction"}))
	is.False(m.Matches(nil))
	is.Equal("a map containing value <equal to reference>", m.Describe())
	is.Equal("was <map[]>", m.DescribeMismatch(map[string]any{}))
}

func TestFormatMismatch(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	m := matcher.EqualTo(4)
	msg := matcher.FormatMismatch(m, 5)
	is.Contains(msg, "Expected:")
	is.Contains(msg, "but:")
	is.Contains(msg, "4")
	is.Contains(msg, "5")

	nilMsg := matcher.FormatMismatch[int](nil, 5)
	is.Contains(nilMsg, "<nil matcher>")
}

type countingTestMatcher struct {
	calls *int
}

func (m *countingTestMatcher) Matches(_ int) bool {
	*m.calls++
	return true
}

func (m *countingTestMatcher) Describe() string {
	return "counting"
}

func (m *countingTestMatcher) DescribeMismatch(_ int) string {
	return "counting mismatch"
}

type staticBoolMatcher struct {
	val bool
}

func (m *staticBoolMatcher) Matches(_ int) bool {
	return m.val
}

func (m *staticBoolMatcher) Describe() string {
	return "static"
}

func (m *staticBoolMatcher) DescribeMismatch(_ int) string {
	return "static mismatch"
}

func TestAllOf_ShortCircuit(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	calls := 0
	counting := &countingTestMatcher{calls: &calls}
	alwaysFalse := &staticBoolMatcher{val: false}

	m := matcher.AllOf[int](alwaysFalse, counting)
	is.False(m.Matches(1))
	is.Equal(0, calls, "second matcher should not be called after first fails")
}

func TestAnyOf_ShortCircuit(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	calls := 0
	counting := &countingTestMatcher{calls: &calls}
	alwaysTrue := &staticBoolMatcher{val: true}

	m := matcher.AnyOf[int](alwaysTrue, counting)
	is.True(m.Matches(1))
	is.Equal(0, calls, "second matcher should not be called after first succeeds")
}

func TestNumAndStrConvenienceWrappers(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// Num wrapper with float64
	numGt := matcher.Num(matcher.GreaterThan(10.0))
	must.NotNil(numGt)
	is.True(numGt.Matches(15.5))
	is.True(numGt.Matches(15))
	is.False(numGt.Matches(5))
	is.False(numGt.Matches("not-a-number"))
	is.Equal("a value greater than <10>", numGt.Describe())
	is.Equal("was <5>", numGt.DescribeMismatch(5))

	// Num wrapper with int
	numInt := matcher.Num(matcher.GreaterThan(10))
	is.True(numInt.Matches(15))
	is.False(numInt.Matches(5))

	// Num wrapper with int64
	numInt64 := matcher.Num(matcher.GreaterThan(int64(10)))
	is.True(numInt64.Matches(int64(15)))
	is.False(numInt64.Matches(int64(5)))

	// Num wrapper with float32
	numF32 := matcher.Num(matcher.GreaterThan(float32(10.0)))
	is.True(numF32.Matches(float32(15.0)))
	is.False(numF32.Matches(float32(5.0)))

	// Num wrapper with int32, uint, uint64
	numI32 := matcher.Num(matcher.GreaterThan(int32(10)))
	is.True(numI32.Matches(int32(15)))
	is.False(numI32.Matches(int32(5)))

	numUint := matcher.Num(matcher.GreaterThan(uint(10)))
	is.True(numUint.Matches(uint(15)))
	is.False(numUint.Matches(uint(5)))

	numUint64 := matcher.Num(matcher.GreaterThan(uint64(10)))
	is.True(numUint64.Matches(uint64(15)))
	is.False(numUint64.Matches(uint64(5)))

	// Str wrapper
	strCont := matcher.Str(matcher.ContainsString("valid"))
	must.NotNil(strCont)
	is.True(strCont.Matches("a valid string"))
	is.False(strCont.Matches(nil))
	is.Equal(`a string containing "valid"`, strCont.Describe())
	is.Equal(`was "invalid"`, strCont.DescribeMismatch("invalid"))
}

func TestToComparable_StringPrecedence(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// String vs string must compare as strings (not convert "0123" to 123.0)
	m := matcher.EqualTo[any]("123")
	is.True(m.Matches("123"))
	is.False(m.Matches("0123"))

	// Cross-type string and number should coerce
	is.True(m.Matches(123))
	is.True(m.Matches(float64(123)))
}

func TestContainsInAnyOrder_OverlappingMatchers(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Overlapping matchers where greedy search could pick the wrong item first
	// 10 matches both GreaterThan(5) and EqualTo(10)
	// 6 matches GreaterThan(5)
	// Both must find valid assignments [10 -> EqualTo(10), 6 -> GreaterThan(5)]
	m := matcher.ContainsInAnyOrderMatchers(
		matcher.GreaterThanNum(5),
		matcher.EqualTo[any](10),
	)

	is.True(m.Matches([]any{10, 6}))
	is.True(m.Matches([]any{6, 10}))
	is.False(m.Matches([]any{6, 6})) // missing 10
}

func TestCloseTo_NegativeDelta(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Negative delta should be treated as absolute value
	m := matcher.CloseTo(10.0, -1.0)
	is.True(m.Matches(9.5))
	is.True(m.Matches(10.5))
	is.False(m.Matches(12.0))
}

func TestMatchesRegex_InvalidPattern(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// Invalid regex must not panic — matcher fails gracefully
	m := matcher.MatchesRegex(`[invalid`)
	is.False(m.Matches("anything"))
	is.Contains(m.Describe(), "invalid regex")
	is.Contains(m.Describe(), "[invalid")
	is.Contains(m.DescribeMismatch("x"), "invalid regex")
}

func TestCombinableMatcher_MixedModePanics(t *testing.T) {
	t.Parallel()

	t.Run("Or on And chain panics", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		is.Panics(func() {
			matcher.Both(matcher.GreaterThan(0)).And(matcher.LessThan(10)).Or(matcher.EqualTo(99))
		})
	})

	t.Run("And on Or chain panics", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		is.Panics(func() {
			matcher.Either(matcher.EqualTo("a")).Or(matcher.EqualTo("b")).And(matcher.EqualTo("c"))
		})
	})
}
