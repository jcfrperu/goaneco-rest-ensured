# goaneco-rest-ensured

Плавная и читаемая библиотека для тестирования REST API на Go — вдохновлённая [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

Проект с открытым исходным кодом, распространяется под [лицензией MIT](LICENSE). Создан в первую
очередь для личного использования, но идеи, предложения и улучшения приветствуются — вклад,
который делает проект более полезным, всегда рассматривается. Вы также можете свободно сделать
форк и развивать его в любом направлении.

Пишите тесты, которые читаются как разговор с API:

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

## Установка

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## Быстрый старт

Каждый тест следует одному и тому же трёхфазному шаблону:

```
Given()  →  настройка запроса (URL, заголовки, параметры, тело)
When()   →  выбор HTTP-метода и пути
Then()   →  декларация ожидаемого результата
```

Полный пример тестирования Petstore API:

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
        Port(0). // 0 — порт не добавляется к URL
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## Построение запроса

Вся настройка происходит между `Given()` и `When()`. Методы можно соединять в любом порядке.

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
    BodyObject(myStruct). // автоматически сериализуется в JSON
    When().
    Get("/users/{userId}")
```

| Метод | Описание |
|---|---|
| `BaseURI(url)` | Установить базовый URL |
| `Port(n)` | Переопределить порт (используйте `0`, чтобы не добавлять) |
| `Header(key, value)` | Добавить заголовок запроса |
| `Accept(contentType)` | Установить заголовок Accept |
| `ContentType(ct)` | Установить заголовок Content-Type |
| `QueryParam(key, values...)` | Добавить параметры строки запроса |
| `PathParam(key, value)` | Заменить `{key}` в пути URL |
| `BodyObject(obj)` | Сериализовать структуру в JSON и установить как тело |
| `Body(string)` | Установить тело в виде строки |
| `BodyBytes([]byte)` | Установить тело в виде байтов |
| `FormParam(key, value)` | Добавить параметр формы |
| `MultiPart(name, data)` | Добавить поле multipart |
| `MultiPartNamed(name, filename, data, mime)` | Добавить именованный файл multipart |
| `MultiPartFile(name, filepath)` | Добавить файл multipart с диска |
| `Cookie(name, value)` | Добавить cookie |
| `Auth().Basic(user, pass)` | HTTP Basic аутентификация |
| `Auth().OAuth2(token)` | Аутентификация Bearer-токеном |

---

## HTTP-методы

После `When()` выберите нужный глагол:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

`When()` можно пропустить — библиотека принимает оба стиля:

```go
petstore().BodyObject(pet).Post("/pet")        // краткая форма
petstore().BodyObject(pet).When().Post("/pet") // явная форма
```

---

## Утверждения

Соединяйте утверждения между `Then()` и `AssertAllNoFail(t)`. Все они собираются — при первой ошибке тест останавливается в `AssertAllNoFail`.

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

| Метод | Проверяет |
|---|---|
| `StatusCode(n)` | HTTP код статуса |
| `ContentType(ct)` | Заголовок Content-Type |
| `Header(key, value)` | Значение заголовка ответа |
| `Body(path, value)` | Поле JSON по пути gjson |
| `BodyLength(path, matcher)` | Длину JSON-массива |
| `BodyMatches(path, matcher)` | Соответствие поля пользовательскому матчеру |
| `BodyMatchesSchema(schema)` | Полную валидацию по JSON Schema |
| `BodyMatchesSchemaFile(path)` | JSON Schema из файла |
| `Time(matcher)` | Время ответа |

### Синтаксис пути JSON

Пути следуют синтаксису [gjson](https://github.com/tidwall/gjson):

```go
Body("name", "Fluffy")          // поле верхнего уровня
Body("category.name", "Dogs")   // вложенное поле
Body("tags.0.name", "cute")     // индекс массива
Body("0.status", "available")   // корневой массив
```

### Матчеры

Используйте матчеры для числовых и сравнительных утверждений:

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

## Извлечение значений

Используйте `JsonPath()` или `XmlPath()` для чтения значений из ответа — например, чтобы передать ID из одного запроса в следующий:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// Использовать petID в следующем запросе
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## Переиспользуемые спецификации

Определите общую конфигурацию один раз и используйте её во многих тестах без повторений.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// Переиспользовать в нескольких запросах
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// Применить один и тот же контракт к нескольким эндпоинтам
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## Реальные примеры

Директория `examples/` содержит 20 полных тестовых сценариев против публичного Petstore API, организованных от простого к сложному:

| # | Что демонстрирует |
|---|---|
| 00 | Query-параметры, заголовок Accept, утверждения для JSON-массивов |
| 01 | Минимальный GET-запрос |
| 02 | POST с телом-структурой (автоматически сериализуется в JSON) |
| 03 | Path-параметры, утверждения на уровне поля тела |
| 04 | PUT для обновления ресурса |
| 05 | DELETE с пользовательским заголовком запроса |
| 06 | POST и извлечение сгенерированного ID из ответа |
| 07 | Использование извлечённого ID в последующем GET |
| 08 | Паттерн «создать и удалить» |
| 09 | POST пользователя с уникальным именем без коллизий |
| 10 | Несколько query-параметров (эндпоинт входа) |
| 11 | GET пользователя по path-параметру |
| 12 | PUT для обновления пользователя |
| 13 | DELETE пользователя по имени |
| 14 | Загрузка файла через multipart |
| 15 | Несколько значений для одного query-параметра |
| 16 | Негативный тест — ожидание 404 |
| 17 | Переиспользуемый `RequestSpec` для двух запросов |
| 18 | Переиспользуемый `ResponseSpec` для двух запросов |
| 19 | Полный CRUD-цикл: создание → чтение → обновление → удаление → проверка 404 |

Запустить все примеры:

```bash
go test ./examples/... -v -timeout 60s
```

Запустить один конкретный пример:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## Полный CRUD в одном тесте

Вот как выглядит `goaneco-rest-ensured` в своём лучшем виде — полный жизненный цикл на чистом, читаемом Go:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // СОЗДАНИЕ
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // ЧТЕНИЕ
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // ОБНОВЛЕНИЕ
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // УДАЛЕНИЕ
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // ПРОВЕРКА УДАЛЕНИЯ
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## Конфигурация

Глобальные настройки по умолчанию можно задать один раз в начале набора тестов:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## Фильтры

Фильтры позволяют перехватывать каждый запрос/ответ — полезны для cookies, CSRF-токенов, замера времени и потоков аутентификации:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

Встроенные фильтры:

| Фильтр | Назначение |
|---|---|
| `NewCookieFilter()` | Сохранять cookies между запросами (как браузерная сессия) |
| `NewSessionFilter()` | Разделять состояние сессии |
| `NewTimingFilter()` | Измерять и проверять время ответа |
| `NewCsrfFilter()` | Автоматически обрабатывать CSRF-токены |
| `NewFormAuthFilter(url, user, pass)` | Аутентификация через HTML-форму входа |

---

## Запуск тестов

```bash
# Модульные и интеграционные тесты
go test ./...

# Только примеры Petstore
go test ./examples/... -v -timeout 60s

# Один конкретный пример
go test ./examples/... -run TestPetstore_05 -v
```

---

## Лицензия

Лицензия MIT — подробности в [LICENSE](LICENSE).

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)