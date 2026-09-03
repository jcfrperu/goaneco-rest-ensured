# goaneco-rest-ensured

一个流畅、易读的 Go 语言 REST API 测试库 — 灵感来源于 [REST Assured](https://github.com/rest-assured/rest-assured)（Java）。

本项目为开源项目，采用 [MIT 许可证](LICENSE) 发布。最初主要为个人使用而构建，
但欢迎各类想法、建议和改进 — 能让项目更有价值的贡献始终值得考虑。
你也可以自由地 fork 本项目，将其带往你想要的任何方向。

写出像与 API 对话一样可读的测试：

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

## 安装

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## 快速开始

每个测试都遵循相同的三阶段模式：

```
Given()  →  配置请求（URL、请求头、参数、请求体）
When()   →  选择 HTTP 方法和路径
Then()   →  声明期望的结果
```

一个完整的 Petstore API 测试示例：

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
        Port(0). // 使用 0 以省略默认端口
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## 构建请求

所有配置都在 `Given()` 和 `When()` 之间进行，方法可以按任意顺序链式调用。

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
    BodyObject(myStruct). // 自动序列化为 JSON
    When().
    Get("/users/{userId}")
```

| 方法 | 说明 |
|---|---|
| `BaseURI(url)` | 设置基础 URL |
| `Port(n)` | 覆盖端口（使用 `0` 省略） |
| `Header(key, value)` | 添加请求头 |
| `Accept(contentType)` | 设置 Accept 请求头 |
| `ContentType(ct)` | 设置 Content-Type 请求头 |
| `QueryParam(key, values...)` | 添加查询参数 |
| `PathParam(key, value)` | 替换 URL 路径中的 `{key}` |
| `BodyObject(obj)` | 将结构体序列化为 JSON 并设为请求体 |
| `Body(string)` | 设置原始字符串请求体 |
| `BodyBytes([]byte)` | 设置原始字节请求体 |
| `FormParam(key, value)` | 添加表单参数 |
| `MultiPart(name, data)` | 添加 multipart 字段 |
| `MultiPartNamed(name, filename, data, mime)` | 添加命名的 multipart 文件 |
| `MultiPartFile(name, filepath)` | 从磁盘添加 multipart 文件 |
| `Cookie(name, value)` | 添加 cookie |
| `Auth().Basic(user, pass)` | HTTP Basic 认证 |
| `Auth().OAuth2(token)` | Bearer Token 认证 |

---

## HTTP 方法

在 `When()` 之后选择 HTTP 动词：

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

也可以省略 `When()` — 库支持两种风格：

```go
petstore().BodyObject(pet).Post("/pet")        // 简短形式
petstore().BodyObject(pet).When().Post("/pet") // 显式形式
```

---

## 断言

在 `Then()` 和 `AssertAllNoFail(t)` 之间链式添加断言。所有断言会被收集 — 任何失败都会在 `AssertAllNoFail` 处停止测试。

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

| 方法 | 检查内容 |
|---|---|
| `StatusCode(n)` | HTTP 状态码 |
| `ContentType(ct)` | Content-Type 请求头 |
| `Header(key, value)` | 响应头的值 |
| `Body(path, value)` | 通过 gjson 路径访问 JSON 字段 |
| `BodyLength(path, matcher)` | JSON 数组的长度 |
| `BodyMatches(path, matcher)` | 字段与自定义 matcher 匹配 |
| `BodyMatchesSchema(schema)` | 完整的 JSON Schema 验证 |
| `BodyMatchesSchemaFile(path)` | 从文件加载 JSON Schema |
| `Time(matcher)` | 响应时间断言 |

### JSON 路径语法

路径遵循 [gjson](https://github.com/tidwall/gjson) 语法：

```go
Body("name", "Fluffy")          // 顶级字段
Body("category.name", "Dogs")   // 嵌套字段
Body("tags.0.name", "cute")     // 数组索引
Body("0.status", "available")   // 根数组
```

### Matchers

使用 matcher 进行数值和比较断言：

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

## 提取值

使用 `JsonPath()` 或 `XmlPath()` 从响应中读取值 — 例如将一个请求中生成的 ID 传递给下一个请求：

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// 在后续请求中使用 petID
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## 可复用的 Spec

在多个测试中共享通用配置，无需重复编写。

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// 在多个请求中复用
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// 将相同的契约应用于多个接口
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## 实战示例

`examples/` 目录包含 20 个完整的测试场景，针对公开的 Petstore API，从简单到高级依次排列：

| # | 演示内容 |
|---|---|
| 00 | 查询参数、Accept 请求头、JSON 数组断言 |
| 01 | 最简单的 GET 请求 |
| 02 | 带结构体请求体的 POST（自动序列化为 JSON） |
| 03 | 路径参数、字段级请求体断言 |
| 04 | 使用 PUT 更新资源 |
| 05 | 带自定义请求头的 DELETE |
| 06 | POST 并从响应中提取生成的 ID |
| 07 | 在后续 GET 中使用提取的 ID |
| 08 | 创建-然后-删除模式 |
| 09 | 使用无冲突的唯一名称 POST 用户 |
| 10 | 多个查询参数（登录接口） |
| 11 | 通过路径参数 GET 用户 |
| 12 | 使用 PUT 更新用户 |
| 13 | 通过用户名 DELETE 用户 |
| 14 | Multipart 文件上传 |
| 15 | 同一查询参数的多个值 |
| 16 | 负面测试 — 期望 404 |
| 17 | 在两个请求中复用 `RequestSpec` |
| 18 | 在两个请求中复用 `ResponseSpec` |
| 19 | 完整 CRUD 生命周期：创建 → 读取 → 更新 → 删除 → 验证 404 |

运行所有示例：

```bash
go test ./examples/... -v -timeout 60s
```

运行单个示例：

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## 一个测试中的完整 CRUD

这就是 `goaneco-rest-ensured` 的最佳表现 — 用简洁可读的 Go 代码完成完整的生命周期：

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // 创建
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // 读取
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // 更新
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // 删除
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // 验证已删除
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## 配置

全局默认值可以在测试套件开始时统一设置：

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## 过滤器

过滤器允许拦截每个请求/响应 — 适用于 cookie、CSRF 令牌、计时和认证流程：

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

内置过滤器：

| 过滤器 | 用途 |
|---|---|
| `NewCookieFilter()` | 在请求间持久化 cookie（类似浏览器会话） |
| `NewSessionFilter()` | 共享会话状态 |
| `NewTimingFilter()` | 测量并断言响应时间 |
| `NewCsrfFilter()` | 自动处理 CSRF 令牌 |
| `NewFormAuthFilter(url, user, pass)` | 通过 HTML 登录表单进行认证 |

---

## 运行测试

```bash
# 单元测试和集成测试
go test ./...

# 仅运行 Petstore 示例
go test ./examples/... -v -timeout 60s

# 运行单个示例
go test ./examples/... -run TestPetstore_05 -v
```

---

## 许可证

MIT 许可证 — 详情见 [LICENSE](LICENSE)。

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)