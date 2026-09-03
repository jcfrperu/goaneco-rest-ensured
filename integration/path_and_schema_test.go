package integration_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestIntegration_JsonPathAndXmlPath(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	// 1. JSONPath queries
	respJSON := rest.Given().
		BaseURI(ts.URL).
		Get("/json/store")

	must.NoError(respJSON.Err())
	is.Equal(http.StatusOK, respJSON.StatusCode())

	jp := respJSON.JsonPath()
	is.Equal("Nigel Rees", jp.GetString("store.book.0.author"))
	is.Equal(8.95, jp.GetFloat64("store.book.0.price"))
	is.Equal("red", jp.GetString("store.bicycle.color"))

	// 2. XMLPath queries (XPath 1.0)
	respXML := rest.Given().
		BaseURI(ts.URL).
		Get("/xml/store")

	must.NoError(respXML.Err())
	is.Equal(http.StatusOK, respXML.StatusCode())

	xp, err := respXML.XmlPath()
	must.NoError(err)
	is.Equal("Nigel Rees", xp.GetString("//book[@category='reference']/author"))
	is.Equal("Sword of Honour", xp.GetString("//book[@category='fiction']/title"))
	is.Equal("red", xp.GetString("//bicycle/color"))
}

func TestIntegration_JsonSchemaValidation(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	ts := integrationServer

	userSchemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"total": {"type": "integer"},
			"users": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "integer"},
						"name": {"type": "string"},
						"email": {"type": "string"}
					},
					"required": ["id", "name", "email"]
				}
			}
		},
		"required": ["total", "users"]
	}`

	tempDir := t.TempDir()
	schemaFilePath := filepath.Join(tempDir, "users_schema.json")
	err := os.WriteFile(schemaFilePath, []byte(userSchemaJSON), 0600)
	must.NoError(err)

	// Validate response from /json/users against JSON Schema (inline and file)
	valid := rest.Given().
		BaseURI(ts.URL).
		Get("/json/users").
		Then().
		StatusCode(http.StatusOK).
		BodyMatchesSchema(userSchemaJSON).
		BodyMatchesSchemaFile(schemaFilePath)

	must.False(valid.HasFailures())
	is.Empty(valid.Failures())
}
