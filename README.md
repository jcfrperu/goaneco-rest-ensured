# goaneco-rest-ensured

A fluent, readable REST API testing library for Go — inspired by [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

This project is open source, released under the [MIT License](LICENSE). It was built primarily
for personal use, but ideas, suggestions, and improvements are welcome — contributions that make
the project more useful are always considered. You are also free to fork it and take it in
whatever direction you like.

Write tests that read like a conversation with the API:

```go
rest.Given().
    BaseURI("https://petstore.swagger.io/v2").
    Accept(rest.ContentTypeJSON).
    QueryParam("status", "available").
    When().
    Get("/pet/findByStatus").
    Then().
    StatusCode(http.StatusOK).
    ContentType(rest.ContentTypeJSON).
    Body("0.status", "available").
    AssertAllNoFail(t)
```

---

## Installation

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## Quick Start

Every test follows the same three-phase pattern:

```
Given()  →  set up your request (URL, headers, params, body)
When()   →  choose the HTTP method and path
Then()   →  declare what you expect
```

Here is a complete example testing the Petstore API:

```go
package mypackage_test

import (
    "net/http"
    "testing"

    "github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func TestGetInventory(t *testing.T) {
    rest.Given().
        BaseURI("https://petstore.swagger.io/v2").
        Port(0). // use 0 to skip the default port
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## Building a Request

All setup happens between `Given()` and `When()`. Methods can be chained in any order.

```go
rest.Given().
    BaseURI("https://api.example.com").
    Port(443).
    Header("Authorization", "Bearer my-token").
    Accept(rest.ContentTypeJSON).
    ContentType(rest.ContentTypeJSON).
    QueryParam("page", "1").
    QueryParam("size", "20").
    PathParam("userId", 42).
    BodyObject(myStruct). // serialized to JSON automatically
    When().
    Get("/users/{userId}")
```

| Method | Description |
|---|---|
| `BaseURI(url)` | Set the base URL |
| `Port(n)` | Override the port (use `0` to omit) |
| `Header(key, value)` | Add a request header |
| `Accept(contentType)` | Set the Accept header |
| `ContentType(ct)` | Set the Content-Type header |
| `QueryParam(key, values...)` | Add query string parameters |
| `PathParam(key, value)` | Replace `{key}` in the URL path |
| `BodyObject(obj)` | Serialize a struct to JSON and set it as the body |
| `Body(string)` | Set a raw string body |
| `BodyBytes([]byte)` | Set a raw bytes body |
| `FormParam(key, value)` | Add a form parameter |
| `MultiPart(name, data)` | Add a multipart field |
| `MultiPartNamed(name, filename, data, mime)` | Add a named multipart file |
| `MultiPartFile(name, filepath)` | Add a multipart file from disk |
| `Cookie(name, value)` | Add a cookie |
| `Auth().Basic(user, pass)` | HTTP Basic authentication |
| `Auth().OAuth2(token)` | Bearer token authentication |

---

## HTTP Methods

After `When()`, choose the verb:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

You can also skip `When()` — the library accepts both styles:

```go
petstore().BodyObject(pet).Post("/pet")        // short form
petstore().BodyObject(pet).When().Post("/pet") // explicit form
```

---

## Assertions

Chain assertions between `Then()` and `AssertAllNoFail(t)`. All assertions are collected — if any fail, the test stops at `AssertAllNoFail`.

```go
response.Then().
    StatusCode(http.StatusOK).
    ContentType(rest.ContentTypeJSON).
    Header("X-Rate-Limit-Limit", "500").
    Body("name", "Fluffy").
    Body("status", "available").
    BodyLength("tags", greaterThanOrEqualTo(1)).
    AssertAllNoFail(t)
```

| Method | Checks |
|---|---|
| `StatusCode(n)` | HTTP status code |
| `ContentType(ct)` | Content-Type header |
| `Header(key, value)` | Response header value |
| `Body(path, value)` | JSON field by gjson path |
| `BodyLength(path, matcher)` | Length of a JSON array |
| `BodyMatches(path, matcher)` | Field matches a custom matcher |
| `BodyMatchesSchema(schema)` | Full JSON Schema validation |
| `BodyMatchesSchemaFile(path)` | JSON Schema from a file |
| `Time(matcher)` | Response time assertion |

### JSON Path Syntax

Paths follow [gjson](https://github.com/tidwall/gjson) syntax:

```go
Body("name", "Fluffy")          // top-level field
Body("category.name", "Dogs")   // nested field
Body("tags.0.name", "cute")     // array index
Body("0.status", "available")   // root array
```

### Matchers

Use matchers for numeric and comparative assertions:

```go
rest.EqualTo(42)
rest.GreaterThan(0)
rest.GreaterThanOrEqualTo(1)
rest.LessThan(100)
rest.Not(rest.EqualTo(0))
rest.HasItems("cat", "dog")
rest.ContainsString("available")
```

---

## Extracting Values

Use `JsonPath()` or `XmlPath()` to read values from the response — for example to pass an ID from one request to the next:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// Use petID in a follow-up request
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## Reusable Specs

Share common setup across many tests without repeating yourself.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// Reuse in multiple requests
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// Apply the same contract to multiple endpoints
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## Real-World Examples

The `examples/` directory contains 20 complete test scenarios against the public Petstore API, organized from simple to advanced:

| # | What it demonstrates |
|---|---|
| 00 | Query parameters, Accept header, JSON array assertions |
| 01 | Minimal GET request |
| 02 | POST with a struct body (auto-serialized to JSON) |
| 03 | Path parameters, field-level body assertions |
| 04 | PUT to update a resource |
| 05 | DELETE with a custom request header |
| 06 | POST and extract the generated ID from the response |
| 07 | Use the extracted ID in a follow-up GET |
| 08 | Create-then-delete pattern |
| 09 | POST a user with a unique, collision-safe username |
| 10 | Multiple query parameters (login endpoint) |
| 11 | GET a user by path parameter |
| 12 | PUT to update a user |
| 13 | DELETE a user by username |
| 14 | Multipart file upload |
| 15 | Multiple values for the same query parameter |
| 16 | Negative test — expect a 404 |
| 17 | Reusable `RequestSpec` across two requests |
| 18 | Reusable `ResponseSpec` across two requests |
| 19 | Full CRUD lifecycle: create → read → update → delete → verify 404 |

Run all examples:

```bash
go test ./examples/... -v -timeout 60s
```

Run a single example:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## Full CRUD in One Test

This is what `goaneco-rest-ensured` looks like at its best — a complete lifecycle in clean, readable Go:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // CREATE
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // READ
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // UPDATE
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // DELETE
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // VERIFY GONE
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## Configuration

Global defaults can be set once at the start of a test suite:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## Filters

Filters let you intercept every request/response — useful for cookies, CSRF tokens, timing, and auth flows:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

Built-in filters:

| Filter | Purpose |
|---|---|
| `NewCookieFilter()` | Persist cookies across requests (like a browser session) |
| `NewSessionFilter()` | Share session state |
| `NewTimingFilter()` | Measure and assert response time |
| `NewCsrfFilter()` | Handle CSRF tokens automatically |
| `NewFormAuthFilter(url, user, pass)` | Authenticate via an HTML login form |

---

## Running the Tests

```bash
# Unit and integration tests
go test ./...

# Only the Petstore examples
go test ./examples/... -v -timeout 60s

# One specific example
go test ./examples/... -run TestPetstore_05 -v
```

---

## License

MIT License — see [LICENSE](LICENSE) for details.

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)
