# goaneco-rest-ensured

مكتبة اختبار REST API سلسة وسهلة القراءة للغة Go — مستوحاة من [REST Assured](https://github.com/rest-assured/rest-assured) (Java).

هذا المشروع مفتوح المصدر، مُرخَّص بموجب [رخصة MIT](LICENSE). تم تطويره في الأساس للاستخدام الشخصي،
لكن الأفكار والاقتراحات والتحسينات مرحّب بها — المساهمات التي تجعل المشروع أكثر فائدةً تُؤخذ دائماً
بعين الاعتبار. أنت أيضاً حر في عمل نسخة (fork) وأخذه في أي اتجاه تريده.

اكتب اختبارات تُقرأ كمحادثة مع الـ API:

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

## التثبيت

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## البداية السريعة

كل اختبار يتبع نفس النمط الثلاثي المراحل:

```
Given()  →  إعداد الطلب (URL، الترويسات، المعاملات، الجسم)
When()   →  اختيار طريقة HTTP والمسار
Then()   →  تعريف ما تتوقعه
```

مثال كامل لاختبار Petstore API:

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
        Port(0). // استخدم 0 لتجنب إضافة المنفذ الافتراضي
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## بناء الطلب

كل الإعداد يحدث بين `Given()` و `When()`. يمكن ربط الدوال بأي ترتيب.

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
    BodyObject(myStruct). // يُحوَّل تلقائياً إلى JSON
    When().
    Get("/users/{userId}")
```

| الدالة | الوصف |
|---|---|
| `BaseURI(url)` | تعيين الـ URL الأساسي |
| `Port(n)` | تخصيص المنفذ (استخدم `0` لحذفه) |
| `Header(key, value)` | إضافة ترويسة للطلب |
| `Accept(contentType)` | تعيين ترويسة Accept |
| `ContentType(ct)` | تعيين ترويسة Content-Type |
| `QueryParam(key, values...)` | إضافة معاملات الاستعلام |
| `PathParam(key, value)` | استبدال `{key}` في مسار الـ URL |
| `BodyObject(obj)` | تحويل struct إلى JSON وتعيينه كجسم للطلب |
| `Body(string)` | تعيين جسم نصي خام |
| `BodyBytes([]byte)` | تعيين جسم بايت خام |
| `FormParam(key, value)` | إضافة معامل نموذج |
| `MultiPart(name, data)` | إضافة حقل multipart |
| `MultiPartNamed(name, filename, data, mime)` | إضافة ملف multipart مُسمَّى |
| `MultiPartFile(name, filepath)` | إضافة ملف multipart من القرص |
| `Cookie(name, value)` | إضافة cookie |
| `Auth().Basic(user, pass)` | مصادقة HTTP Basic |
| `Auth().OAuth2(token)` | مصادقة Bearer token |

---

## طرق HTTP

بعد `When()`، اختر الفعل المناسب:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

يمكن أيضاً تخطي `When()` — المكتبة تقبل كلا الأسلوبين:

```go
petstore().BodyObject(pet).Post("/pet")        // الصيغة المختصرة
petstore().BodyObject(pet).When().Post("/pet") // الصيغة الصريحة
```

---

## التحقق من النتائج

اربط التحققات بين `Then()` و `AssertAllNoFail(t)`. تُجمَع جميعها — عند أول فشل يتوقف الاختبار في `AssertAllNoFail`.

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

| الدالة | تتحقق من |
|---|---|
| `StatusCode(n)` | كود حالة HTTP |
| `ContentType(ct)` | ترويسة Content-Type |
| `Header(key, value)` | قيمة ترويسة الاستجابة |
| `Body(path, value)` | حقل JSON عبر مسار gjson |
| `BodyLength(path, matcher)` | طول مصفوفة JSON |
| `BodyMatches(path, matcher)` | مطابقة الحقل لـ matcher مخصص |
| `BodyMatchesSchema(schema)` | التحقق الكامل من JSON Schema |
| `BodyMatchesSchemaFile(path)` | JSON Schema من ملف |
| `Time(matcher)` | التحقق من زمن الاستجابة |

### صياغة مسار JSON

المسارات تتبع صياغة [gjson](https://github.com/tidwall/gjson):

```go
Body("name", "Fluffy")          // حقل في المستوى الأعلى
Body("category.name", "Dogs")   // حقل متداخل
Body("tags.0.name", "cute")     // فهرس مصفوفة
Body("0.status", "available")   // مصفوفة جذر
```

### Matchers

استخدم الـ matchers للتحقق العددي والمقارن:

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

## استخراج القيم

استخدم `JsonPath()` أو `XmlPath()` لقراءة القيم من الاستجابة — مثلاً لتمرير معرّف من طلب إلى آخر:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// استخدام petID في الطلب التالي
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## المواصفات القابلة للإعادة

شارك الإعداد المشترك بين اختبارات عديدة دون تكرار.

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// إعادة الاستخدام في عدة طلبات
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// تطبيق نفس العقد على عدة نقاط نهاية
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## أمثلة عملية

يحتوي مجلد `examples/` على 20 سيناريو اختبار كامل ضد Petstore API العام، منظّمة من البسيط إلى المتقدم:

| # | ما يُظهره |
|---|---|
| 00 | معاملات الاستعلام، ترويسة Accept، التحقق من مصفوفات JSON |
| 01 | طلب GET بسيط |
| 02 | POST بجسم struct (يُحوَّل تلقائياً إلى JSON) |
| 03 | معاملات المسار، التحقق من حقول الجسم |
| 04 | PUT لتحديث مورد |
| 05 | DELETE مع ترويسة طلب مخصصة |
| 06 | POST واستخراج المعرّف المُولَّد من الاستجابة |
| 07 | استخدام المعرّف المستخرج في GET لاحق |
| 08 | نمط الإنشاء والحذف |
| 09 | POST لمستخدم باسم فريد بلا تعارضات |
| 10 | معاملات استعلام متعددة (نقطة نهاية تسجيل الدخول) |
| 11 | GET لمستخدم عبر معامل المسار |
| 12 | PUT لتحديث مستخدم |
| 13 | DELETE لمستخدم عبر اسم المستخدم |
| 14 | رفع ملف multipart |
| 15 | قيم متعددة لنفس معامل الاستعلام |
| 16 | اختبار سلبي — توقّع 404 |
| 17 | `RequestSpec` قابل للإعادة لطلبين |
| 18 | `ResponseSpec` قابل للإعادة لطلبين |
| 19 | دورة CRUD كاملة: إنشاء → قراءة → تحديث → حذف → التحقق من 404 |

تشغيل جميع الأمثلة:

```bash
go test ./examples/... -v -timeout 60s
```

تشغيل مثال واحد:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## CRUD كامل في اختبار واحد

هكذا يبدو `goaneco-rest-ensured` في أفضل حالاته — دورة حياة كاملة بكود Go نظيف وسهل القراءة:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // إنشاء
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // قراءة
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // تحديث
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // حذف
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // التحقق من الحذف
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## الإعداد العام

يمكن تعيين الإعدادات الافتراضية العامة مرة واحدة في بداية مجموعة الاختبارات:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## الفلاتر

تتيح الفلاتر اعتراض كل طلب/استجابة — مفيدة للـ cookies وتوكنات CSRF وقياس الوقت وتدفقات المصادقة:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

الفلاتر المضمّنة:

| الفلتر | الغرض |
|---|---|
| `NewCookieFilter()` | استمرارية الـ cookies بين الطلبات (مثل جلسة المتصفح) |
| `NewSessionFilter()` | مشاركة حالة الجلسة |
| `NewTimingFilter()` | قياس وقت الاستجابة والتحقق منه |
| `NewCsrfFilter()` | التعامل التلقائي مع توكنات CSRF |
| `NewFormAuthFilter(url, user, pass)` | المصادقة عبر نموذج تسجيل دخول HTML |

---

## تشغيل الاختبارات

```bash
# اختبارات الوحدة والتكامل
go test ./...

# أمثلة Petstore فقط
go test ./examples/... -v -timeout 60s

# مثال واحد محدد
go test ./examples/... -run TestPetstore_05 -v
```

---

## الرخصة

رخصة MIT — انظر [LICENSE](LICENSE) للتفاصيل.

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)