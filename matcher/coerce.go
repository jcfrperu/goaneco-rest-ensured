package matcher

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// toFloat64 converts any numeric type (or json.Number/numeric string) to float64.
// Returns (value, true) if conversion succeeded, or (0, false) otherwise.
func toFloat64(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uintptr:
		return float64(val), true
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// toAnySlice normalizes a slice or array of any type into []any.
// Returns (slice, true) if v is a slice/array, or (nil, false) otherwise.
func toAnySlice(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}
	if s, ok := v.([]any); ok {
		return s, true
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}

	length := rv.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = rv.Index(i).Interface()
	}
	return result, true
}

// toStringSlice converts a collection or slice into []string.
func toStringSlice(v any) ([]string, bool) {
	if v == nil {
		return nil, false
	}
	if s, ok := v.([]string); ok {
		return s, true
	}

	slice, ok := toAnySlice(v)
	if !ok {
		return nil, false
	}

	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = fmt.Sprintf("%v", item)
	}
	return result, true
}

// toComparable normalizes two values so they can be accurately compared.
// It resolves differences between gjson numeric types (float64) and Go literals (int, int64, etc.).
// Returns (normalizedActual, normalizedExpected, true) on success.
func toComparable(actual, expected any) (any, any, bool) {
	if actual == nil && expected == nil {
		return nil, nil, true
	}

	// If both are strings, compare as strings directly
	if actStr, ok := actual.(string); ok {
		if expStr, ok := expected.(string); ok {
			return actStr, expStr, true
		}
	}

	// If both are bools, compare as bools directly
	if actBool, ok := actual.(bool); ok {
		if expBool, ok := expected.(bool); ok {
			return actBool, expBool, true
		}
	}

	// Try numeric coercion (handles float64, int, int64, json.Number, and mixed numeric types).
	// Note: a numeric string such as "123" is coerced to float64, so EqualTo[any]("123")
	// matches the number 123.  Use typed matchers (EqualTo[string]) when you need to
	// distinguish between the string "123" and the number 123.
	actNum, actOk := toFloat64(actual)
	expNum, expOk := toFloat64(expected)
	if actOk && expOk {
		return actNum, expNum, true
	}

	// Slices comparison check
	actSlice, actIsSlice := toAnySlice(actual)
	expSlice, expIsSlice := toAnySlice(expected)
	if actIsSlice && expIsSlice {
		return actSlice, expSlice, true
	}

	// Maps comparison check
	if actMap, ok := actual.(map[string]any); ok {
		if expMap, ok := expected.(map[string]any); ok {
			return actMap, expMap, true
		}
	}

	return actual, expected, false
}
