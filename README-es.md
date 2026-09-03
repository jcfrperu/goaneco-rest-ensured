# goaneco-rest-ensured

Una biblioteca de pruebas de API REST fluida y legible para Go — inspirada en [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

Este proyecto es de código abierto, publicado bajo la [Licencia MIT](LICENSE). Fue creado principalmente
para uso personal, pero las ideas, sugerencias y mejoras son bienvenidas — las contribuciones que hagan
el proyecto más útil siempre serán consideradas. También eres libre de hacer un fork y llevarlo en
la dirección que quieras.

Escribe pruebas que se lean como una conversación con la API:

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

## Instalación

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## Inicio Rápido

Cada prueba sigue el mismo patrón de tres fases:

```
Given()  →  configura la solicitud (URL, cabeceras, parámetros, cuerpo)
When()   →  elige el método HTTP y la ruta
Then()   →  declara lo que esperas
```

Un ejemplo completo probando la API Petstore:

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
        Port(0). // usa 0 para omitir el puerto predeterminado
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## Construcción de una Solicitud

Todo el proceso de configuración ocurre entre `Given()` y `When()`. Los métodos pueden encadenarse en cualquier orden.

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
    BodyObject(myStruct). // serializado a JSON automáticamente
    When().
    Get("/users/{userId}")
```

| Método | Descripción |
|---|---|
| `BaseURI(url)` | Establece la URL base |
| `Port(n)` | Sobreescribe el puerto (usa `0` para omitirlo) |
| `Header(key, value)` | Agrega una cabecera de solicitud |
| `Accept(contentType)` | Establece la cabecera Accept |
| `ContentType(ct)` | Establece la cabecera Content-Type |
| `QueryParam(key, values...)` | Agrega parámetros de consulta |
| `PathParam(key, value)` | Reemplaza `{key}` en la ruta de la URL |
| `BodyObject(obj)` | Serializa una estructura a JSON y la establece como cuerpo |
| `Body(string)` | Establece un cuerpo de texto plano |
| `BodyBytes([]byte)` | Establece un cuerpo de bytes sin procesar |
| `FormParam(key, value)` | Agrega un parámetro de formulario |
| `MultiPart(name, data)` | Agrega un campo multipart |
| `MultiPartNamed(name, filename, data, mime)` | Agrega un archivo multipart con nombre |
| `MultiPartFile(name, filepath)` | Agrega un archivo multipart desde disco |
| `Cookie(name, value)` | Agrega una cookie |
| `Auth().Basic(user, pass)` | Autenticación HTTP Basic |
| `Auth().OAuth2(token)` | Autenticación con token Bearer |

---

## Métodos HTTP

Después de `When()`, elige el verbo:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

También puedes omitir `When()` — la biblioteca acepta ambos estilos:

```go
petstore().BodyObject(pet).Post("/pet")        // forma corta
petstore().BodyObject(pet).When().Post("/pet") // forma explícita
```

---

## Aserciones

Encadena aserciones entre `Then()` y `AssertAllNoFail(t)`. Todas se recopilan — si alguna falla, la prueba se detiene en `AssertAllNoFail`.

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

| Método | Verifica |
|---|---|
| `StatusCode(n)` | Código de estado HTTP |
| `ContentType(ct)` | Cabecera Content-Type |
| `Header(key, value)` | Valor de cabecera de respuesta |
| `Body(path, value)` | Campo JSON por ruta gjson |
| `BodyLength(path, matcher)` | Longitud de un arreglo JSON |
| `BodyMatches(path, matcher)` | Campo coincide con un matcher personalizado |
| `BodyMatchesSchema(schema)` | Validación completa de JSON Schema |
| `BodyMatchesSchemaFile(path)` | JSON Schema desde un archivo |
| `Time(matcher)` | Aserción sobre el tiempo de respuesta |

### Sintaxis de Ruta JSON

Las rutas siguen la sintaxis de [gjson](https://github.com/tidwall/gjson):

```go
Body("name", "Fluffy")          // campo de nivel superior
Body("category.name", "Dogs")   // campo anidado
Body("tags.0.name", "cute")     // índice de arreglo
Body("0.status", "available")   // arreglo raíz
```

### Matchers

Usa matchers para aserciones numéricas y comparativas:

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

## Extracción de Valores

Usa `JsonPath()` o `XmlPath()` para leer valores de la respuesta — por ejemplo, para pasar un ID de una solicitud a la siguiente:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// Usar petID en una solicitud de seguimiento
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## Especificaciones Reutilizables

Comparte la configuración común entre muchas pruebas sin repetirte.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// Reutilizar en múltiples solicitudes
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// Aplicar el mismo contrato a múltiples endpoints
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## Ejemplos del Mundo Real

El directorio `examples/` contiene 20 escenarios de prueba completos contra la API pública Petstore, organizados de simple a avanzado:

| # | Qué demuestra |
|---|---|
| 00 | Parámetros de consulta, cabecera Accept, aserciones en arreglos JSON |
| 01 | Solicitud GET mínima |
| 02 | POST con cuerpo de estructura (serializado automáticamente a JSON) |
| 03 | Parámetros de ruta, aserciones a nivel de campo del cuerpo |
| 04 | PUT para actualizar un recurso |
| 05 | DELETE con cabecera de solicitud personalizada |
| 06 | POST y extracción del ID generado de la respuesta |
| 07 | Uso del ID extraído en un GET de seguimiento |
| 08 | Patrón crear-y-eliminar |
| 09 | POST de usuario con nombre único sin colisiones |
| 10 | Múltiples parámetros de consulta (endpoint de login) |
| 11 | GET de usuario por parámetro de ruta |
| 12 | PUT para actualizar un usuario |
| 13 | DELETE de usuario por nombre de usuario |
| 14 | Carga de archivo multipart |
| 15 | Múltiples valores para el mismo parámetro de consulta |
| 16 | Prueba negativa — esperar un 404 |
| 17 | `RequestSpec` reutilizable en dos solicitudes |
| 18 | `ResponseSpec` reutilizable en dos solicitudes |
| 19 | Ciclo de vida CRUD completo: crear → leer → actualizar → eliminar → verificar 404 |

Ejecutar todos los ejemplos:

```bash
go test ./examples/... -v -timeout 60s
```

Ejecutar un ejemplo específico:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## CRUD Completo en Una Prueba

Así es como `goaneco-rest-ensured` luce en su mejor forma — un ciclo de vida completo en Go limpio y legible:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // CREAR
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // LEER
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // ACTUALIZAR
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // ELIMINAR
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // VERIFICAR QUE FUE ELIMINADO
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## Configuración

Los valores predeterminados globales se pueden establecer una vez al inicio de la suite de pruebas:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## Filtros

Los filtros permiten interceptar cada solicitud/respuesta — útiles para cookies, tokens CSRF, medición de tiempos y flujos de autenticación:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

Filtros incorporados:

| Filtro | Propósito |
|---|---|
| `NewCookieFilter()` | Persistir cookies entre solicitudes (como una sesión de navegador) |
| `NewSessionFilter()` | Compartir estado de sesión |
| `NewTimingFilter()` | Medir y verificar el tiempo de respuesta |
| `NewCsrfFilter()` | Manejar tokens CSRF automáticamente |
| `NewFormAuthFilter(url, user, pass)` | Autenticarse mediante un formulario de login HTML |

---

## Ejecución de Pruebas

```bash
# Pruebas unitarias y de integración
go test ./...

# Solo los ejemplos de Petstore
go test ./examples/... -v -timeout 60s

# Un ejemplo específico
go test ./examples/... -run TestPetstore_05 -v
```

---

## Licencia

Licencia MIT — consulta [LICENSE](LICENSE) para más detalles.

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)