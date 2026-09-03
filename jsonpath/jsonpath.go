package jsonpath

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/tidwall/gjson"
)

// ErrInvalidPath is returned when a requested JSON path is malformed.
var ErrInvalidPath = errors.New("invalid json path")

// ErrNilJsonPath is returned when an operation is attempted on a nil JsonPath receiver.
var ErrNilJsonPath = errors.New("nil JsonPath instance")

// JsonPath provides standalone JSON parsing and query capabilities.
type JsonPath struct {
	raw    string
	config *Config
}

// From parses a JSON string and returns a JsonPath queryable instance.
func From(jsonStr string) *JsonPath {
	return &JsonPath{
		raw:    jsonStr,
		config: DefaultConfig(),
	}
}

// FromBytes parses JSON bytes and returns a JsonPath instance.
func FromBytes(data []byte) *JsonPath {
	return From(string(data))
}

// FromReader reads JSON from an io.Reader and returns a JsonPath instance.
func FromReader(r io.Reader) (*JsonPath, error) {
	if r == nil {
		return nil, errors.New("nil reader provided")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading json data: %w", err)
	}
	return FromBytes(data), nil
}

// Using returns a new JsonPath instance using the specified configuration.
func (jp *JsonPath) Using(cfg *Config) *JsonPath {
	if jp == nil {
		return nil
	}
	return &JsonPath{
		raw:    jp.raw,
		config: cfg,
	}
}

// Get evaluates a path expression against the JSON document.
func (jp *JsonPath) Get(path string) gjson.Result {
	if jp == nil || jp.raw == "" {
		return gjson.Result{}
	}
	return gjson.Get(jp.raw, path)
}

// GetString extracts a string at the specified path.
func (jp *JsonPath) GetString(path string) string {
	return jp.Get(path).String()
}

// GetInt extracts an int value at the specified path.
func (jp *JsonPath) GetInt(path string) int {
	return int(jp.Get(path).Int())
}

// GetInt64 extracts an int64 value at the specified path.
func (jp *JsonPath) GetInt64(path string) int64 {
	return jp.Get(path).Int()
}

// GetFloat extracts a float64 value at the specified path.
func (jp *JsonPath) GetFloat(path string) float64 {
	return jp.Get(path).Float()
}

// GetFloat64 extracts a float64 value at the specified path (alias for GetFloat).
func (jp *JsonPath) GetFloat64(path string) float64 {
	return jp.GetFloat(path)
}

// GetBool extracts a boolean value at the specified path.
func (jp *JsonPath) GetBool(path string) bool {
	return jp.Get(path).Bool()
}

// GetList extracts a list of gjson.Result elements at the specified path.
func (jp *JsonPath) GetList(path string) []gjson.Result {
	res := jp.Get(path)
	if !res.IsArray() {
		return []gjson.Result{}
	}
	return res.Array()
}

// GetStringList extracts a string slice at the specified path.
func (jp *JsonPath) GetStringList(path string) []string {
	res := jp.Get(path)
	if !res.IsArray() {
		return []string{}
	}
	arr := res.Array()
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		result = append(result, item.String())
	}
	return result
}

// GetIntList extracts an int slice at the specified path.
func (jp *JsonPath) GetIntList(path string) []int {
	res := jp.Get(path)
	if !res.IsArray() {
		return []int{}
	}
	arr := res.Array()
	result := make([]int, 0, len(arr))
	for _, item := range arr {
		result = append(result, int(item.Int()))
	}
	return result
}

// GetFloatList extracts a float64 slice at the specified path.
func (jp *JsonPath) GetFloatList(path string) []float64 {
	res := jp.Get(path)
	if !res.IsArray() {
		return []float64{}
	}
	arr := res.Array()
	result := make([]float64, 0, len(arr))
	for _, item := range arr {
		result = append(result, item.Float())
	}
	return result
}

// GetMap extracts a key-value map of gjson.Result at the specified path.
func (jp *JsonPath) GetMap(path string) map[string]gjson.Result {
	res := jp.Get(path)
	if !res.IsObject() {
		return make(map[string]gjson.Result)
	}
	return res.Map()
}

// GetObject deserializes the JSON value at path into the target pointer.
func (jp *JsonPath) GetObject(path string, target any) error {
	if jp == nil {
		return ErrNilJsonPath
	}
	res := jp.Get(path)
	if !res.Exists() {
		return fmt.Errorf("%w: path %q does not exist", ErrInvalidPath, path)
	}
	if err := json.Unmarshal([]byte(res.Raw), target); err != nil {
		return fmt.Errorf("unmarshaling json path %q: %w", path, err)
	}
	return nil
}

// Exists checks if the specified path exists in the JSON document.
func (jp *JsonPath) Exists(path string) bool {
	if jp == nil || jp.raw == "" {
		return false
	}
	return jp.Get(path).Exists()
}

// Pretty returns a pretty-printed JSON representation of the document.
func (jp *JsonPath) Pretty() string {
	if jp == nil || jp.raw == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(jp.raw), "", "  "); err != nil {
		return jp.raw
	}
	return buf.String()
}

// GetObjectTyped deserializes the value at path into type T.
func GetObjectTyped[T any](jp *JsonPath, path string) (T, error) {
	var target T
	if jp == nil {
		return target, ErrNilJsonPath
	}
	if err := jp.GetObject(path, &target); err != nil {
		return target, err
	}
	return target, nil
}
