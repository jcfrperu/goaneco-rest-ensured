# goaneco-rest-ensured

Uma biblioteca de testes de API REST fluente e legível para Go — inspirada no [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

Este projeto é de código aberto, lançado sob a [Licença MIT](LICENSE). Foi desenvolvido principalmente
para uso pessoal, mas ideias, sugestões e melhorias são bem-vindas — contribuições que tornem
o projeto mais útil são sempre consideradas. Você também é livre para fazer um fork e levá-lo na
direção que quiser.

Escreva testes que se leem como uma conversa com a API:

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

## Instalação

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## Início Rápido

Todo teste segue o mesmo padrão de três fases:

```
Given()  →  configure a requisição (URL, cabeçalhos, parâmetros, corpo)
When()   →  escolha o método HTTP e o caminho
Then()   →  declare o que você espera
```

Um exemplo completo testando a API Petstore:

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
        Port(0). // use 0 para omitir a porta padrão
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## Construindo uma Requisição

Toda a configuração acontece entre `Given()` e `When()`. Os métodos podem ser encadeados em qualquer ordem.

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
    BodyObject(myStruct). // serializado para JSON automaticamente
    When().
    Get("/users/{userId}")
```

| Método | Descrição |
|---|---|
| `BaseURI(url)` | Define a URL base |
| `Port(n)` | Sobrescreve a porta (use `0` para omitir) |
| `Header(key, value)` | Adiciona um cabeçalho de requisição |
| `Accept(contentType)` | Define o cabeçalho Accept |
| `ContentType(ct)` | Define o cabeçalho Content-Type |
| `QueryParam(key, values...)` | Adiciona parâmetros de consulta |
| `PathParam(key, value)` | Substitui `{key}` no caminho da URL |
| `BodyObject(obj)` | Serializa uma struct para JSON e define como corpo |
| `Body(string)` | Define um corpo de texto simples |
| `BodyBytes([]byte)` | Define um corpo de bytes brutos |
| `FormParam(key, value)` | Adiciona um parâmetro de formulário |
| `MultiPart(name, data)` | Adiciona um campo multipart |
| `MultiPartNamed(name, filename, data, mime)` | Adiciona um arquivo multipart nomeado |
| `MultiPartFile(name, filepath)` | Adiciona um arquivo multipart do disco |
| `Cookie(name, value)` | Adiciona um cookie |
| `Auth().Basic(user, pass)` | Autenticação HTTP Basic |
| `Auth().OAuth2(token)` | Autenticação com token Bearer |

---

## Métodos HTTP

Após `When()`, escolha o verbo:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

`When()` também pode ser omitido — a biblioteca aceita ambos os estilos:

```go
petstore().BodyObject(pet).Post("/pet")        // forma curta
petstore().BodyObject(pet).When().Post("/pet") // forma explícita
```

---

## Asserções

Encadeie asserções entre `Then()` e `AssertAllNoFail(t)`. Todas são coletadas — se alguma falhar, o teste para em `AssertAllNoFail`.

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
| `StatusCode(n)` | Código de status HTTP |
| `ContentType(ct)` | Cabeçalho Content-Type |
| `Header(key, value)` | Valor de um cabeçalho de resposta |
| `Body(path, value)` | Campo JSON por caminho gjson |
| `BodyLength(path, matcher)` | Comprimento de um array JSON |
| `BodyMatches(path, matcher)` | Campo corresponde a um matcher personalizado |
| `BodyMatchesSchema(schema)` | Validação completa de JSON Schema |
| `BodyMatchesSchemaFile(path)` | JSON Schema a partir de um arquivo |
| `Time(matcher)` | Asserção sobre o tempo de resposta |

### Sintaxe de Caminho JSON

Os caminhos seguem a sintaxe do [gjson](https://github.com/tidwall/gjson):

```go
Body("name", "Fluffy")          // campo de nível superior
Body("category.name", "Dogs")   // campo aninhado
Body("tags.0.name", "cute")     // índice de array
Body("0.status", "available")   // array raiz
```

### Matchers

Use matchers para asserções numéricas e comparativas:

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

## Extraindo Valores

Use `JsonPath()` ou `XmlPath()` para ler valores da resposta — por exemplo, para passar um ID de uma requisição para a próxima:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// Usar petID em uma requisição subsequente
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## Specs Reutilizáveis

Compartilhe a configuração comum entre muitos testes sem se repetir.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// Reutilizar em múltiplas requisições
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// Aplicar o mesmo contrato em múltiplos endpoints
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## Exemplos do Mundo Real

O diretório `examples/` contém 20 cenários de teste completos contra a API pública Petstore, organizados do mais simples ao mais avançado:

| # | O que demonstra |
|---|---|
| 00 | Parâmetros de consulta, cabeçalho Accept, asserções em arrays JSON |
| 01 | Requisição GET mínima |
| 02 | POST com corpo de struct (serializado automaticamente para JSON) |
| 03 | Parâmetros de caminho, asserções no nível de campo do corpo |
| 04 | PUT para atualizar um recurso |
| 05 | DELETE com cabeçalho de requisição personalizado |
| 06 | POST e extração do ID gerado da resposta |
| 07 | Uso do ID extraído em um GET subsequente |
| 08 | Padrão criar-e-deletar |
| 09 | POST de usuário com nome único sem colisões |
| 10 | Múltiplos parâmetros de consulta (endpoint de login) |
| 11 | GET de usuário por parâmetro de caminho |
| 12 | PUT para atualizar um usuário |
| 13 | DELETE de usuário por nome de usuário |
| 14 | Upload de arquivo multipart |
| 15 | Múltiplos valores para o mesmo parâmetro de consulta |
| 16 | Teste negativo — esperar um 404 |
| 17 | `RequestSpec` reutilizável em duas requisições |
| 18 | `ResponseSpec` reutilizável em duas requisições |
| 19 | Ciclo de vida CRUD completo: criar → ler → atualizar → deletar → verificar 404 |

Executar todos os exemplos:

```bash
go test ./examples/... -v -timeout 60s
```

Executar um exemplo específico:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## CRUD Completo em Um Teste

É assim que `goaneco-rest-ensured` se parece no seu melhor — um ciclo de vida completo em Go limpo e legível:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // CRIAR
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // LER
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // ATUALIZAR
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // DELETAR
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // VERIFICAR QUE FOI DELETADO
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## Configuração

Os padrões globais podem ser definidos uma vez no início da suíte de testes:

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

Filtros permitem interceptar cada requisição/resposta — úteis para cookies, tokens CSRF, medição de tempo e fluxos de autenticação:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

Filtros embutidos:

| Filtro | Propósito |
|---|---|
| `NewCookieFilter()` | Persistir cookies entre requisições (como uma sessão de navegador) |
| `NewSessionFilter()` | Compartilhar estado de sessão |
| `NewTimingFilter()` | Medir e verificar o tempo de resposta |
| `NewCsrfFilter()` | Tratar tokens CSRF automaticamente |
| `NewFormAuthFilter(url, user, pass)` | Autenticar via formulário de login HTML |

---

## Executando os Testes

```bash
# Testes unitários e de integração
go test ./...

# Apenas os exemplos do Petstore
go test ./examples/... -v -timeout 60s

# Um exemplo específico
go test ./examples/... -run TestPetstore_05 -v
```

---

## Licença

Licença MIT — veja [LICENSE](LICENSE) para detalhes.

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)