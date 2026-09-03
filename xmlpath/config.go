// Package xmlpath provides an XML path evaluation engine powered by XPath and antchfx/xmlquery.
package xmlpath

// Config holds configuration options for XmlPath evaluation.
// antchfx/xmlquery handles encoding internally; charset configuration is not needed.
type Config struct{}

// DefaultConfig returns a new Config with standard defaults.
func DefaultConfig() *Config {
	return &Config{}
}
