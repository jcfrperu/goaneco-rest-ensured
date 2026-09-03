# goaneco-rest-ensured

Go के लिए एक सरल और पठनीय REST API परीक्षण लाइब्रेरी — [REST Assured](https://github.com/rest-assured/rest-assured) (Java) से प्रेरित।

यह प्रोजेक्ट ओपन सोर्स है, [MIT लाइसेंस](LICENSE) के तहत जारी किया गया है। इसे मुख्यतः
व्यक्तिगत उपयोग के लिए बनाया गया था, लेकिन विचार, सुझाव और सुधार स्वागत योग्य हैं — ऐसा योगदान
जो प्रोजेक्ट को अधिक उपयोगी बनाए, हमेशा विचार किया जाता है। आप इसे fork करके अपनी
पसंद की दिशा में ले जाने के लिए भी स्वतंत्र हैं।

ऐसे टेस्ट लिखें जो API से बातचीत की तरह पढ़े जाएं:

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

## इंस्टॉलेशन

```bash
go get github.com/jcfrperu/goaneco-rest-ensured
```

---

## त्वरित प्रारंभ

हर टेस्ट एक ही तीन-चरण के पैटर्न का पालन करता है:

```
Given()  →  अनुरोध सेट करें (URL, हेडर, पैरामीटर, बॉडी)
When()   →  HTTP मेथड और पाथ चुनें
Then()   →  अपेक्षित परिणाम घोषित करें
```

Petstore API का परीक्षण करने वाला एक पूरा उदाहरण:

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
        Port(0). // डिफ़ॉल्ट पोर्ट न जोड़ने के लिए 0 उपयोग करें
        When().
        Get("/store/inventory").
        Then().
        StatusCode(http.StatusOK).
        ContentType(rest.ContentTypeJSON).
        AssertAllNoFail(t)
}
```

---

## अनुरोध बनाना

सभी सेटअप `Given()` और `When()` के बीच होता है। मेथड्स को किसी भी क्रम में chain किया जा सकता है।

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
    BodyObject(myStruct). // स्वचालित रूप से JSON में परिवर्तित
    When().
    Get("/users/{userId}")
```

| मेथड | विवरण |
|---|---|
| `BaseURI(url)` | बेस URL सेट करें |
| `Port(n)` | पोर्ट ओवरराइड करें (हटाने के लिए `0` उपयोग करें) |
| `Header(key, value)` | अनुरोध हेडर जोड़ें |
| `Accept(contentType)` | Accept हेडर सेट करें |
| `ContentType(ct)` | Content-Type हेडर सेट करें |
| `QueryParam(key, values...)` | क्वेरी पैरामीटर जोड़ें |
| `PathParam(key, value)` | URL पाथ में `{key}` बदलें |
| `BodyObject(obj)` | struct को JSON में serialize करके बॉडी सेट करें |
| `Body(string)` | कच्ची टेक्स्ट बॉडी सेट करें |
| `BodyBytes([]byte)` | कच्ची बाइट्स बॉडी सेट करें |
| `FormParam(key, value)` | फ़ॉर्म पैरामीटर जोड़ें |
| `MultiPart(name, data)` | multipart फ़ील्ड जोड़ें |
| `MultiPartNamed(name, filename, data, mime)` | नामित multipart फ़ाइल जोड़ें |
| `MultiPartFile(name, filepath)` | डिस्क से multipart फ़ाइल जोड़ें |
| `Cookie(name, value)` | cookie जोड़ें |
| `Auth().Basic(user, pass)` | HTTP Basic प्रमाणीकरण |
| `Auth().OAuth2(token)` | Bearer टोकन प्रमाणीकरण |

---

## HTTP मेथड्स

`When()` के बाद, वर्ब चुनें:

```go
.When().Get("/path")
.When().Post("/path")
.When().Put("/path")
.When().Patch("/path")
.When().Delete("/path")
.When().Options("/path")
.When().Head("/path")
```

`When()` को छोड़ा भी जा सकता है — लाइब्रेरी दोनों शैलियाँ स्वीकार करती है:

```go
petstore().BodyObject(pet).Post("/pet")        // संक्षिप्त रूप
petstore().BodyObject(pet).When().Post("/pet") // स्पष्ट रूप
```

---

## Assertions

`Then()` और `AssertAllNoFail(t)` के बीच assertions को chain करें। सभी एकत्र किए जाते हैं — किसी भी विफलता पर टेस्ट `AssertAllNoFail` पर रुकता है।

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

| मेथड | जाँचता है |
|---|---|
| `StatusCode(n)` | HTTP स्टेटस कोड |
| `ContentType(ct)` | Content-Type हेडर |
| `Header(key, value)` | प्रतिक्रिया हेडर का मान |
| `Body(path, value)` | gjson पाथ द्वारा JSON फ़ील्ड |
| `BodyLength(path, matcher)` | JSON array की लंबाई |
| `BodyMatches(path, matcher)` | फ़ील्ड कस्टम matcher से मेल खाती है |
| `BodyMatchesSchema(schema)` | पूर्ण JSON Schema वैलिडेशन |
| `BodyMatchesSchemaFile(path)` | फ़ाइल से JSON Schema |
| `Time(matcher)` | प्रतिक्रिया समय assertion |

### JSON पाथ सिंटैक्स

पाथ [gjson](https://github.com/tidwall/gjson) सिंटैक्स का अनुसरण करते हैं:

```go
Body("name", "Fluffy")          // शीर्ष-स्तरीय फ़ील्ड
Body("category.name", "Dogs")   // नेस्टेड फ़ील्ड
Body("tags.0.name", "cute")     // array इंडेक्स
Body("0.status", "available")   // रूट array
```

### Matchers

संख्यात्मक और तुलनात्मक assertions के लिए matchers उपयोग करें:

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

## मान निकालना

प्रतिक्रिया से मान पढ़ने के लिए `JsonPath()` या `XmlPath()` उपयोग करें — उदाहरण के लिए एक अनुरोध से अगले में ID पास करने के लिए:

```go
created := petstore().BodyObject(pet).Post("/pet")

petID := created.JsonPath().GetInt64("id")
name  := created.JsonPath().GetString("name")

// अगले अनुरोध में petID उपयोग करें
petstore().PathParam("petId", petID).Get("/pet/{petId}").Then()...
```

---

## पुन:उपयोगी Specs

बिना दोहराए कई टेस्ट में सामान्य सेटअप साझा करें।

### Request Spec

```go
sharedSpec := rest.NewRequestSpecBuilder().
    SetBaseURI("https://petstore.swagger.io/v2").
    AddHeader("Accept", "application/json").
    Build()

// कई अनुरोधों में पुन:उपयोग
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "available").Get("/pet/findByStatus")
rest.Given().Spec(sharedSpec).Port(0).QueryParam("status", "sold").Get("/pet/findByStatus")
```

### Response Spec

```go
okJSON := rest.NewResponseSpecBuilder().
    ExpectStatusCode(http.StatusOK).
    ExpectContentType(rest.ContentTypeJSON).
    Build()

// एक ही अनुबंध कई endpoints पर लागू करें
resp1.Then().Spec(okJSON).Body("name", "Pet One").AssertAllNoFail(t)
resp2.Then().Spec(okJSON).Body("name", "Pet Two").AssertAllNoFail(t)
```

---

## वास्तविक उदाहरण

`examples/` डायरेक्टरी में सार्वजनिक Petstore API के विरुद्ध 20 पूर्ण परीक्षण परिदृश्य हैं, सरल से उन्नत तक:

| # | क्या दर्शाता है |
|---|---|
| 00 | क्वेरी पैरामीटर, Accept हेडर, JSON array assertions |
| 01 | न्यूनतम GET अनुरोध |
| 02 | struct बॉडी के साथ POST (स्वचालित JSON serialization) |
| 03 | पाथ पैरामीटर, फ़ील्ड-स्तरीय बॉडी assertions |
| 04 | संसाधन अपडेट करने के लिए PUT |
| 05 | कस्टम हेडर के साथ DELETE |
| 06 | POST और प्रतिक्रिया से उत्पन्न ID निकालना |
| 07 | बाद के GET में निकाले गए ID का उपयोग |
| 08 | बनाओ-और-हटाओ पैटर्न |
| 09 | टकराव-रहित अनूठे नाम के साथ उपयोगकर्ता POST |
| 10 | कई क्वेरी पैरामीटर (login endpoint) |
| 11 | पाथ पैरामीटर द्वारा उपयोगकर्ता GET |
| 12 | उपयोगकर्ता अपडेट करने के लिए PUT |
| 13 | उपयोगकर्ता नाम द्वारा DELETE |
| 14 | Multipart फ़ाइल अपलोड |
| 15 | एक ही क्वेरी पैरामीटर के लिए कई मान |
| 16 | नकारात्मक टेस्ट — 404 की उम्मीद |
| 17 | दो अनुरोधों में पुन:उपयोगी `RequestSpec` |
| 18 | दो अनुरोधों में पुन:उपयोगी `ResponseSpec` |
| 19 | पूर्ण CRUD जीवनचक्र: बनाएं → पढ़ें → अपडेट करें → हटाएं → 404 जाँचें |

सभी उदाहरण चलाएं:

```bash
go test ./examples/... -v -timeout 60s
```

एक विशेष उदाहरण चलाएं:

```bash
go test ./examples/... -run TestPetstore_19 -v
```

---

## एक टेस्ट में पूरा CRUD

यह `goaneco-rest-ensured` अपने सर्वोत्तम रूप में — साफ़ और पठनीय Go में एक पूर्ण जीवनचक्र:

```go
func TestFullPetLifecycle(t *testing.T) {
    api := rest.Given().BaseURI("https://petstore.swagger.io/v2").Port(0)

    // बनाएं
    newPet := Pet{Name: "Lifecycle Dog", Status: "available", PhotoUrls: []string{}}
    created := api.Accept(rest.ContentTypeJSON).BodyObject(newPet).Post("/pet")
    require.NoError(t, created.Err())
    created.Then().StatusCode(http.StatusOK).Body("name", "Lifecycle Dog").AssertAllNoFail(t)

    petID := created.JsonPath().GetInt64("id")

    // पढ़ें
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().
        StatusCode(http.StatusOK).Body("status", "available").AssertAllNoFail(t)

    // अपडेट करें
    updated := Pet{ID: petID, Name: "Lifecycle Dog", Status: "pending", PhotoUrls: []string{}}
    api.Accept(rest.ContentTypeJSON).BodyObject(updated).Put("/pet").Then().
        StatusCode(http.StatusOK).Body("status", "pending").AssertAllNoFail(t)

    // हटाएं
    api.Header("api_key", "special-key").PathParam("petId", petID).
        Delete("/pet/{petId}").Then().StatusCode(http.StatusOK).AssertAllNoFail(t)

    // हटाने की जाँच करें
    api.Accept(rest.ContentTypeJSON).PathParam("petId", petID).
        Get("/pet/{petId}").Then().StatusCode(http.StatusNotFound).AssertAllNoFail(t)
}
```

---

## कॉन्फ़िगरेशन

टेस्ट सूट के शुरू में वैश्विक डिफ़ॉल्ट एक बार सेट किए जा सकते हैं:

```go
rest.SetConfig(rest.Config{
    HTTPClient: rest.HTTPClientConfig{Timeout: 30 * time.Second},
    SSL:        rest.SSLConfig{Insecure: true},
    Redirect:   rest.RedirectConfig{FollowRedirects: true, MaxRedirects: 5},
    Logging:    rest.LoggingConfig{Enabled: true, LogLevel: "info"},
})
```

---

## फ़िल्टर

फ़िल्टर हर अनुरोध/प्रतिक्रिया को इंटरसेप्ट करने की सुविधा देते हैं — cookies, CSRF टोकन, समय मापन और प्रमाणीकरण प्रवाह के लिए उपयोगी:

```go
rest.Given().
    Filter(rest.NewCookieFilter()).
    Filter(rest.NewTimingFilter()).
    Get("/protected/resource")
```

अंतर्निहित फ़िल्टर:

| फ़िल्टर | उद्देश्य |
|---|---|
| `NewCookieFilter()` | अनुरोधों के बीच cookies बनाए रखना (ब्राउज़र सेशन की तरह) |
| `NewSessionFilter()` | सेशन स्थिति साझा करना |
| `NewTimingFilter()` | प्रतिक्रिया समय मापना और जाँचना |
| `NewCsrfFilter()` | CSRF टोकन स्वचालित रूप से संभालना |
| `NewFormAuthFilter(url, user, pass)` | HTML लॉगिन फ़ॉर्म के ज़रिए प्रमाणीकरण |

---

## टेस्ट चलाना

```bash
# Unit और integration टेस्ट
go test ./...

# केवल Petstore उदाहरण
go test ./examples/... -v -timeout 60s

# एक विशेष उदाहरण
go test ./examples/... -run TestPetstore_05 -v
```

---

## लाइसेंस

MIT लाइसेंस — विवरण के लिए [LICENSE](LICENSE) देखें।

[![Go Reference](https://pkg.go.dev/badge/github.com/jcfrperu/goaneco-rest-ensured.svg)](https://pkg.go.dev/github.com/jcfrperu/goaneco-rest-ensured)
[![Go Report Card](https://goreportcard.com/badge/github.com/jcfrperu/goaneco-rest-ensured)](https://goreportcard.com/report/github.com/jcfrperu/goaneco-rest-ensured)