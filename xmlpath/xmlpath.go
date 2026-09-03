package xmlpath

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

// ErrInvalidXPath is returned when an XPath expression is invalid or matching fails.
var ErrInvalidXPath = errors.New("invalid xpath expression")

// ErrNilXmlPath is returned when an operation is performed on a nil XmlPath receiver.
var ErrNilXmlPath = errors.New("nil XmlPath instance")

// XmlPath provides XML parsing and querying via XPath 1.0 expressions.
type XmlPath struct {
	doc    *xmlquery.Node
	config *Config
}

// From parses an XML string and returns an XmlPath queryable instance.
func From(xmlStr string) (*XmlPath, error) {
	return FromReader(strings.NewReader(xmlStr))
}

// FromBytes parses XML bytes and returns an XmlPath instance.
func FromBytes(data []byte) (*XmlPath, error) {
	return FromReader(bytes.NewReader(data))
}

// FromReader parses XML from an io.Reader and returns an XmlPath instance.
func FromReader(r io.Reader) (*XmlPath, error) {
	if r == nil {
		return nil, errors.New("nil reader provided")
	}

	doc, err := xmlquery.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parsing xml document: %w", err)
	}

	return &XmlPath{
		doc:    doc,
		config: DefaultConfig(),
	}, nil
}

// Using returns a new XmlPath instance using the specified configuration.
func (xp *XmlPath) Using(cfg *Config) *XmlPath {
	if xp == nil {
		return nil
	}
	return &XmlPath{
		doc:    xp.doc,
		config: cfg,
	}
}

// FindNode evaluates the XPath expression and returns the first matching Node or an error.
// Unlike GetNode, it distinguishes between "no match" (nil node, nil error) and an invalid
// expression (nil node, non-nil error), making XPath typos immediately visible.
func (xp *XmlPath) FindNode(expr string) (*xmlquery.Node, error) {
	if xp == nil || xp.doc == nil {
		return nil, ErrNilXmlPath
	}
	node, err := xmlquery.Query(xp.doc, expr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidXPath, err)
	}
	return node, nil
}

// FindNodes evaluates the XPath expression and returns all matching Nodes or an error.
func (xp *XmlPath) FindNodes(expr string) ([]*xmlquery.Node, error) {
	if xp == nil || xp.doc == nil {
		return nil, ErrNilXmlPath
	}
	nodes, err := xmlquery.QueryAll(xp.doc, expr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidXPath, err)
	}
	return nodes, nil
}

// GetNode evaluates the XPath expression and returns the first matching Node, or nil.
// XPath syntax errors are silently discarded; use FindNode when error visibility matters.
func (xp *XmlPath) GetNode(expr string) *xmlquery.Node {
	node, _ := xp.FindNode(expr)
	return node
}

// GetNodes evaluates the XPath expression and returns all matching Nodes.
// XPath syntax errors are silently discarded; use FindNodes when error visibility matters.
func (xp *XmlPath) GetNodes(expr string) []*xmlquery.Node {
	nodes, err := xp.FindNodes(expr)
	if err != nil {
		return []*xmlquery.Node{}
	}
	return nodes
}

// GetString extracts the string/text value of the first matching node at expr.
func (xp *XmlPath) GetString(expr string) string {
	node := xp.GetNode(expr)
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.InnerText())
}

// GetInt extracts an integer at the specified XPath expression.
func (xp *XmlPath) GetInt(expr string) int {
	str := xp.GetString(expr)
	if str == "" {
		return 0
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return val
}

// GetInt64 extracts an int64 at the specified XPath expression.
func (xp *XmlPath) GetInt64(expr string) int64 {
	str := xp.GetString(expr)
	if str == "" {
		return 0
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// GetFloat extracts a float64 at the specified XPath expression.
func (xp *XmlPath) GetFloat(expr string) float64 {
	str := xp.GetString(expr)
	if str == "" {
		return 0.0
	}
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// GetBool extracts a boolean at the specified XPath expression.
func (xp *XmlPath) GetBool(expr string) bool {
	str := strings.ToLower(xp.GetString(expr))
	return str == "true" || str == "1" || str == "yes"
}

// GetStringList extracts a slice of text contents for all nodes matching expr.
func (xp *XmlPath) GetStringList(expr string) []string {
	nodes := xp.GetNodes(expr)
	if len(nodes) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, strings.TrimSpace(node.InnerText()))
	}
	return result
}

// GetIntList extracts a slice of integers for all nodes matching expr.
func (xp *XmlPath) GetIntList(expr string) []int {
	nodes := xp.GetNodes(expr)
	if len(nodes) == 0 {
		return []int{}
	}
	result := make([]int, 0, len(nodes))
	for _, node := range nodes {
		if val, err := strconv.Atoi(strings.TrimSpace(node.InnerText())); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// GetFloatList extracts a slice of float64 for all nodes matching expr.
func (xp *XmlPath) GetFloatList(expr string) []float64 {
	nodes := xp.GetNodes(expr)
	if len(nodes) == 0 {
		return []float64{}
	}
	result := make([]float64, 0, len(nodes))
	for _, node := range nodes {
		if val, err := strconv.ParseFloat(strings.TrimSpace(node.InnerText()), 64); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// Exists checks if any node matches the specified XPath expression.
func (xp *XmlPath) Exists(expr string) bool {
	if xp == nil || xp.doc == nil {
		return false
	}
	return xp.GetNode(expr) != nil
}

// OutputXML returns the raw XML string of the node matching expr.
func (xp *XmlPath) OutputXML(expr string) string {
	node := xp.GetNode(expr)
	if node == nil {
		return ""
	}
	return node.OutputXML(true)
}

// GetObject unmarshals the XML content at expr into target.
func (xp *XmlPath) GetObject(expr string, target any) error {
	if xp == nil {
		return ErrNilXmlPath
	}
	node := xp.GetNode(expr)
	if node == nil {
		return fmt.Errorf("%w: xpath %q matched no nodes", ErrInvalidXPath, expr)
	}
	xmlData := node.OutputXML(true)
	if err := xml.Unmarshal([]byte(xmlData), target); err != nil {
		return fmt.Errorf("unmarshaling xml node %q: %w", expr, err)
	}
	return nil
}

// GetObjectTyped deserializes the XML node at expr into type T.
func GetObjectTyped[T any](xp *XmlPath, expr string) (T, error) {
	var target T
	if xp == nil {
		return target, ErrNilXmlPath
	}
	if err := xp.GetObject(expr, &target); err != nil {
		return target, err
	}
	return target, nil
}
