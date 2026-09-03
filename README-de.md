# goaneco-rest-ensured

Eine flüssige, gut lesbare REST-API-Testbibliothek für Go — inspiriert von [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

Dieses Projekt ist Open Source und steht unter der [MIT-Lizenz](LICENSE). Es wurde in erster Linie
für den persönlichen Gebrauch entwickelt, aber Ideen, Vorschläge und Verbesserungen sind willkommen —
Beiträge, die das Projekt nützlicher machen, werden stets berücksichtigt. Es steht dir auch frei,
einen Fork zu erstellen und ihn in jede beliebige Richtung weiterzuentwickeln.

Schreibe Tests, die sich wie ein Gespräch mit der API lesen:

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

## Schnellstart

Jeder Test folgt demselben dreiphasigen Muster:

```
Given()  →  Anfrage konfigurieren (URL, Header, Parameter, Body)
When()   →  HTTP-Methode und Pfad wählen
Then()   →  Erwartungen deklarieren
```

Ein vollständiges Beispiel mit der Petstore-API:

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
        Port(0). // 0 verwenden, um den Standard-Port wegzulassen
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## Aufbau einer Anfrage

Die gesamte Konfiguration findet zwischen `Given()` und `When()` statt. Methoden können in beliebiger Reihenfolge verkettet werden.

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
    BodyObject(myStruct). // wird automatisch zu JSON serialisiert
    When().
    Get("/users/{userId}")
```

| Methode | Beschreibung |
|---|---|
| `BaseURI(url)` | Basis-URL setzen |
| `Port(n)` | Port überschreiben (mit `0` weglassen) |
| `Header(key, value)` | Anfrage-Header hinzufügen |
| `Accept(contentType)` | Accept-Header setzen |
| `ContentType(ct)` | Content-Type-Header setzen |
| `QueryParam(key, values...)` | Query-Parameter hinzufügen |
| `PathParam(key, value)` | `{key}` im URL-Pfad ersetzen |
| `BodyObject(obj)` | Struct zu JSON serialisieren und als Body setzen |
| `Body(string)` | Rohen Text-Body setzen |
| `BodyBytes([]byte)` | Rohen Byte-Body setzen |
| `FormParam(key, value)` | Formular-Parameter hinzufügen |
| `MultiPart(name, data)` | Multipart-Feld hinzufügen |
| `MultiPartNamed(name, filename, data, mime)` | Benannte Multipart-Datei hinzufügen |
| `MultiPartFile(name, filepath)` | Multipart-Datei vom Datenträger hinzufügen |
| `Cookie(name, value)` | Cookie hinzufügen |
| `Auth().Basic(user, pass)` | HTTP-Basic-Authentifizierung |
| `Auth().OAuth2(token)` | Bearer-Token-Authentifizierung |

---

## HTTP-Methoden

Nach `When()` die gewünschte Methode wählen:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

`When()` kann auch weggelassen werden — die Bibliothek akzeptiert beide Schreibweisen:

```go
petstore().BodyObject(pet).Post("/pet")        // Kurzform
petstore().BodyObject(pet).When().Post("/pet") // Explizite Form
```

---

## Assertions

Assertions werden zwischen `Then()` und `AssertAllNoFail(t)` verkettet. Alle werden gesammelt — bei einem Fehler stoppt der Test bei `AssertAllNoFail`.

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

| Methode | Prüft |
|---|---|
| `StatusCode(n)` | HTTP-Statuscode |
| `ContentType(ct)` | Content-Type-Header |
| `Header(key, value)` | Wert eines Antwort-Headers |
| `Body(path, value)` | JSON-Feld per gjson-Pfad |
| `BodyLength(path, matcher)` | Länge eines JSON-Arrays |
| `BodyMatches(path, matcher)` | Feld entspricht einem benutzerdefinierten Matcher |
| `BodyMatchesSchema(schema)` | Vollständige JSON-Schema-Validierung |
| `BodyMatchesSchemaFile(path)` | JSON-Schema aus einer Datei |
| `Time(matcher)` | Assertion für die Antwortzeit |

### JSON-Pfad-Syntax

Pfade folgen der [gjson](https://github.com/tidwall/gjson)-Syntax:

```go
Body("name", "Fluffy")          // Feld auf oberster Ebene
Body("category.name", "Dogs")   // Verschachteltes Feld
Body("tags.0.name", "cute")     // Array-Index
Body("0.status", "available")   // Root-Array
```

### Matcher

Matcher für numerische und vergleichende Assertions:

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

## Werte extrahieren

Mit `JsonPath()` oder `XmlPath()` Werte aus der Antwort lesen — z. B. um eine ID von einer Anfrage zur nächsten weiterzugeben:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// petID in einer Folgeanfrage verwenden
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## Wiederverwendbare Specs

Gemeinsame Konfiguration einmal definieren und in vielen Tests nutzen — ohne Wiederholungen.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// In mehreren Anfragen wiederverwenden
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// Denselben Vertrag auf mehrere Endpunkte anwenden
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## Praxisbeispiele

Das Verzeichnis `examples/` enthält 20 vollständige Testszenarien gegen die öffentliche Petstore-API, von einfach bis fortgeschritten:

| # | Was demonstriert wird |
|---|---|
| 00 | Query-Parameter, Accept-Header, Assertions für JSON-Arrays |
| 01 | Minimale GET-Anfrage |
| 02 | POST mit Struct-Body (automatisch zu JSON serialisiert) |
| 03 | Pfad-Parameter, feldbezogene Body-Assertions |
| 04 | PUT zum Aktualisieren einer Ressource |
| 05 | DELETE mit benutzerdefiniertem Anfrage-Header |
| 06 | POST und Extraktion der generierten ID aus der Antwort |
| 07 | Verwendung der extrahierten ID in einem Folge-GET |
| 08 | Erstellen-und-Löschen-Muster |
| 09 | POST eines Benutzers mit eindeutigem, kollisionsfreiem Benutzernamen |
| 10 | Mehrere Query-Parameter (Login-Endpunkt) |
| 11 | GET eines Benutzers per Pfad-Parameter |
| 12 | PUT zum Aktualisieren eines Benutzers |
| 13 | DELETE eines Benutzers per Benutzername |
| 14 | Multipart-Datei-Upload |
| 15 | Mehrere Werte für denselben Query-Parameter |
| 16 | Negativtest — 404 erwarten |
| 17 | Wiederverwendbarer `RequestSpec` für zwei Anfragen |
| 18 | Wiederverwendbarer `ResponseSpec` für zwei Anfragen |
| 19 | Vollständiger CRUD-Lebenszyklus: erstellen → lesen → aktualisieren → löschen → 404 prüfen |

Alle Beispiele ausführen:

```bash
go test ./examples/... -v -timeout 60s
```

Ein einzelnes Beispiel ausführen:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## Vollständiges CRUD in einem Test

So sieht `goaneco-rest-ensured` in seiner besten Form aus — ein vollständiger Lebenszyklus in sauberem, lesbarem Go:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // ERSTELLEN
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // LESEN
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // AKTUALISIEREN
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // LÖSCHEN
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // LÖSCHUNG VERIFIZIEREN
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## Konfiguration

Globale Standardwerte können einmalig zu Beginn einer Test-Suite gesetzt werden:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## Filter

Filter ermöglichen das Abfangen jeder Anfrage/Antwort — nützlich für Cookies, CSRF-Token, Zeitmessung und Authentifizierungsflows:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

Eingebaute Filter:

| Filter | Zweck |
|---|---|
| `NewCookieFilter()` | Cookies über Anfragen hinweg persistieren (wie eine Browser-Sitzung) |
| `NewSessionFilter()` | Sitzungsstatus teilen |
| `NewTimingFilter()` | Antwortzeit messen und prüfen |
| `NewCsrfFilter()` | CSRF-Token automatisch behandeln |
| `NewFormAuthFilter(url, user, pass)` | Über ein HTML-Login-Formular authentifizieren |

---

## Tests ausführen

```bash
# Unit- und Integrationstests
go test ./...

# Nur die Petstore-Beispiele
go test ./examples/... -v -timeout 60s

# Ein einzelnes Beispiel
go test ./examples/... -run TestPetstore_05 -v
```

---

## Lizenz

MIT-Lizenz — Details in [LICENSE](LICENSE).

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)