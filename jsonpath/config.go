// Package jsonpath provides a standalone JSON path evaluation engine powered by gjson.
package jsonpath

// Config holds configuration options for JsonPath evaluation.
// gjson handles encoding internally; charset configuration is not needed.
type Config struct{}

// DefaultConfig returns a new Config with standard defaults.
func DefaultConfig() *Config {
	return &Config{}
}
