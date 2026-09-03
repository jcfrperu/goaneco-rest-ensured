// Package testserver provides an in-process HTTP mock server for integration testing.
package testserver

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"
)

// User represents a user model for test endpoints.
type User struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Role    string   `json:"role"`
	Tags    []string `json:"tags"`
	Address Address  `json:"address"`
	Active  bool     `json:"active"`
}

// Address represents a user address model.
type Address struct {
	City    string `json:"city"`
	Street  string `json:"street"`
	ZipCode string `json:"zipCode"`
}

// NewTestServer creates and starts a new httptest.Server configured with all test routes.
func NewTestServer() *httptest.Server {
	mux := http.NewServeMux()
	var sessionsMu sync.RWMutex
	sessions := make(map[string]string) // sessionID -> username

	// -------------------------------------------------------------------------
	// CORE UTILITIES
	// -------------------------------------------------------------------------

	writeJSON := func(w http.ResponseWriter, v any) {
		data, err := json.Marshal(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}

	writeXML := func(w http.ResponseWriter, xml string) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(xml))
	}

	greetingJSON := func(firstName, lastName string) map[string]string {
		return map[string]string{"greeting": fmt.Sprintf("Greetings %s %s", firstName, lastName)}
	}

	greetingXML := func(firstName, lastName string) string {
		return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<greeting>
  <firstName>%s</firstName>
  <lastName>%s</lastName>
</greeting>`, firstName, lastName)
	}

	// -------------------------------------------------------------------------
	// EXISTING ENDPOINTS (unchanged)
	// -------------------------------------------------------------------------

	// /hello and /hello/
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"message": "Hello World"})
	})
	mux.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"message": "Hello World"})
	})

	// /status/{code}
	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		codeStr := strings.TrimPrefix(r.URL.Path, "/status/")
		code, err := strconv.Atoi(codeStr)
		if err != nil {
			http.Error(w, "invalid status code", http.StatusBadRequest)
			return
		}
		w.WriteHeader(code)
		_, _ = fmt.Fprintf(w, "Status %d", code)
	})

	// /echo
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		for k, v := range r.Header {
			if strings.HasPrefix(k, "X-") {
				w.Header()[k] = v
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	// /headers
	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, r.Header)
	})

	// /cookies — returns incoming cookies as JSON or sets cookie
	mux.HandleFunc("/cookies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			name := r.URL.Query().Get("name")
			val := r.URL.Query().Get("value")
			if name == "" {
				name = "session_id"
				val = "xyz-12345"
			}
			http.SetCookie(w, &http.Cookie{Name: name, Value: val, Path: "/"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("cookie set"))
			return
		}
		cookies := make(map[string]string)
		for _, c := range r.Cookies() {
			cookies[c.Name] = c.Value
		}
		writeJSON(w, cookies)
	})

	// /json/users
	mux.HandleFunc("/json/users", func(w http.ResponseWriter, r *http.Request) {
		users := []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin",
				Tags:    []string{"go", "dev", "lead"},
				Address: Address{City: "New York", Street: "5th Ave", ZipCode: "10001"}, Active: true},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "developer",
				Tags:    []string{"go", "docker"},
				Address: Address{City: "San Francisco", Street: "Market St", ZipCode: "94103"}, Active: false},
		}
		writeJSON(w, map[string]any{"total": 2, "users": users})
	})

	// /json/store — 2-book store (kept for backward compat)
	mux.HandleFunc("/json/store", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"store": map[string]any{
				"book": []map[string]any{
					{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
					{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
				},
				"bicycle": map[string]any{"color": "red", "price": 19.95},
			},
		})
	})

	// /xml/store
	mux.HandleFunc("/xml/store", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<store>
  <book category="reference">
    <author>Nigel Rees</author>
    <title>Sayings of the Century</title>
    <price>8.95</price>
  </book>
  <book category="fiction">
    <author>Evelyn Waugh</author>
    <title>Sword of Honour</title>
    <price>12.99</price>
  </book>
  <bicycle>
    <color>red</color>
    <price>19.95</price>
  </bicycle>
</store>`)
	})

	// /auth/basic
	mux.HandleFunc("/auth/basic", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"status": "authenticated", "user": user})
	})

	// /auth/bearer
	mux.HandleFunc("/auth/bearer", func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") || strings.TrimPrefix(authHeader, "Bearer ") != "secret-token-123" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"status": "authenticated"})
	})

	// /form
	mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		res := make(map[string]string)
		for k, v := range r.PostForm {
			if len(v) > 0 {
				res[k] = v[0]
			}
		}
		writeJSON(w, res)
	})

	// /upload
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "cannot parse multipart form", http.StatusBadRequest)
			return
		}
		result := map[string]any{
			"fields": make(map[string]string),
			"files":  make([]string, 0),
		}
		for k, v := range r.MultipartForm.Value {
			if len(v) > 0 {
				result["fields"].(map[string]string)[k] = v[0]
			}
		}
		for _, fileHeaders := range r.MultipartForm.File {
			for _, fh := range fileHeaders {
				result["files"] = append(result["files"].([]string), fh.Filename)
			}
		}
		writeJSON(w, result)
	})

	// /delay/{ms}
	mux.HandleFunc("/delay/", func(w http.ResponseWriter, r *http.Request) {
		msStr := strings.TrimPrefix(r.URL.Path, "/delay/")
		ms, _ := strconv.Atoi(msStr)
		if ms > 10_000 {
			ms = 10_000
		}
		if ms > 0 {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("done"))
	})

	// /gzip/data
	mux.HandleFunc("/gzip/data", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, _ = gw.Write([]byte(`{"compressed":true,"codec":"gzip","payload":"integration-test"}`))
		_ = gw.Close()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(buf.Bytes())
	})

	// /deflate/data
	mux.HandleFunc("/deflate/data", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		fw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
		_, _ = fw.Write([]byte(`{"compressed":true,"codec":"deflate","payload":"integration-test"}`))
		_ = fw.Close()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "deflate")
		_, _ = w.Write(buf.Bytes())
	})

	// /redirect chain
	mux.HandleFunc("/redirect/step1", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect/step2", http.StatusFound)
	})
	mux.HandleFunc("/redirect/step2", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect/final", http.StatusSeeOther)
	})
	mux.HandleFunc("/redirect/final", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"status": "redirect_complete", "step": "final"})
	})

	// /csrf/page and /csrf/submit
	mux.HandleFunc("/csrf/page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Form</title></head><body>
<form action="/csrf/submit" method="POST">
  <input type="hidden" name="_csrf" value="token-abc-987" />
  <input type="text" name="data" value="some-data" />
</form></body></html>`))
	})

	mux.HandleFunc("/csrf/submit", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.FormValue("_csrf") != "token-abc-987" {
			http.Error(w, "Invalid CSRF Token", http.StatusForbidden)
			return
		}
		writeJSON(w, map[string]string{"status": "csrf_accepted", "data": r.FormValue("data")})
	})

	// /login (GET: return form; POST: validate credentials)
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<form action="/login" method="POST"><input type="hidden" name="_csrf" value="login-token-xyz"/></form>`))
			return
		}
		_ = r.ParseForm()
		if r.FormValue("username") == "john" && r.FormValue("password") == "doe" {
			sessID := "sess-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			sessionsMu.Lock()
			sessions[sessID] = r.FormValue("username")
			sessionsMu.Unlock()
			http.SetCookie(w, &http.Cookie{Name: "JSESSIONID", Value: sessID, Path: "/"})
			writeJSON(w, map[string]string{"status": "logged_in", "user": r.FormValue("username")})
			return
		}
		http.Error(w, "Unauthorized Credentials", http.StatusUnauthorized)
	})

	// /secured — requires JSESSIONID session cookie
	mux.HandleFunc("/secured", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("JSESSIONID")
		if err != nil {
			http.Error(w, "Missing Session Cookie", http.StatusUnauthorized)
			return
		}
		sessionsMu.RLock()
		user, exists := sessions[cookie.Value]
		sessionsMu.RUnlock()
		if !exists {
			http.Error(w, "Invalid Session", http.StatusUnauthorized)
			return
		}
		writeJSON(w, map[string]string{"secret": "classified", "user": user})
	})

	// -------------------------------------------------------------------------
	// JSON GREETING ENDPOINTS  (unlock JSONGetITest)
	// -------------------------------------------------------------------------

	// /greet — GET, POST, DELETE
	mux.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeJSON(w, greetingJSON(firstName, lastName))
	})

	// /greetJSON — full greeting object
	mux.HandleFunc("/greetJSON", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeJSON(w, map[string]any{
			"greeting": map[string]string{"firstName": firstName, "lastName": lastName},
		})
	})

	// /lotto — standard lotto JSON used by many JSONGetITest scenarios
	mux.HandleFunc("/lotto", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"lotto": map[string]any{
				"lottoId":         5,
				"winning-numbers": []int{2, 45, 34, 23, 7, 5, 3},
				"winners": []map[string]any{
					{"winnerId": 23, "numbers": []int{2, 45, 34, 23, 3, 5}},
					{"winnerId": 54, "numbers": []int{52, 3, 12, 11, 18, 22}},
				},
			},
		})
	})

	// /jsonStore — 4-book store (matches Java test expectations)
	mux.HandleFunc("/jsonStore", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"store": map[string]any{
				"book": []map[string]any{
					{"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
					{"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99},
					{"category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99},
					{"category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99},
				},
				"bicycle": map[string]any{"color": "red", "price": 19.95},
			},
		})
	})

	// /jsonWithAtSign — JSON containing @ symbol in key
	mux.HandleFunc("/jsonWithAtSign", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"body": map[string]any{"@id": 10, "content": "some content"},
		})
	})

	// /numbers — numeric values
	mux.HandleFunc("/numbers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{
			"values": map[string]any{"pi": 3.14, "answer": 42},
		})
	})

	// /amount — decimal amount
	mux.HandleFunc("/amount", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"amount": 250.00})
	})

	// /anonymous_list_with_numbers — root-level JSON array
	mux.HandleFunc("/anonymous_list_with_numbers", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []any{100, 50, 31.0})
	})

	// /jsonList — list of persons
	mux.HandleFunc("/jsonList", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []map[string]any{
			{"name": "Anders", "address": map[string]string{"street": "Andersgatan", "zipCode": "12345"}},
			{"name": "Sven", "address": map[string]string{"street": "Svensgatan", "zipCode": "67890"}},
		})
	})

	// /i18n — non-ASCII JSON
	mux.HandleFunc("/i18n", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ön":"Är ån"}`))
	})

	// /utf8-body-json — UTF-8 encoded JSON
	mux.HandleFunc("/utf8-body-json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"value":"啊 ☆"}`))
	})

	// /text-json — text/json content type
	mux.HandleFunc("/text-json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		_, _ = w.Write([]byte(`{"test":true}`))
	})

	// /jsonp — JSONP callback endpoint
	mux.HandleFunc("/jsonp", func(w http.ResponseWriter, r *http.Request) {
		callback := r.URL.Query().Get("callback")
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		body := fmt.Sprintf(`{"greeting":"Greetings %s %s"}`, firstName, lastName)
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = fmt.Fprintf(w, "%s(%s);", callback, body)
	})

	// /malformedJson — intentionally invalid JSON
	mux.HandleFunc("/malformedJson", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"value" "missing":"comma"}`))
	})

	// -------------------------------------------------------------------------
	// XML GREETING ENDPOINTS  (unlock XMLGetITest, XPathITest)
	// -------------------------------------------------------------------------

	// /greetXML — GET, POST, OPTIONS
	mux.HandleFunc("/greetXML", func(w http.ResponseWriter, r *http.Request) {
		var firstName, lastName string
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			_ = r.ParseForm()
			firstName = r.FormValue("firstName")
			lastName = r.FormValue("lastName")
		}
		if firstName == "" {
			firstName = r.URL.Query().Get("firstName")
		}
		if lastName == "" {
			lastName = r.URL.Query().Get("lastName")
		}
		writeXML(w, greetingXML(firstName, lastName))
	})

	// /anotherGreetXML — nested name element
	mux.HandleFunc("/anotherGreetXML", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeXML(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<greeting>
  <name>
    <firstName>%s</firstName>
    <lastName>%s</lastName>
  </name>
</greeting>`, firstName, lastName))
	})

	// /greetXMLAttribute — greeting with attributes
	mux.HandleFunc("/greetXMLAttribute", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeXML(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<greeting>
  <name firstName="%s" lastName="%s"/>
</greeting>`, firstName, lastName))
	})

	// /xmlWithMinusInRoot — hyphenated root element
	mux.HandleFunc("/xmlWithMinusInRoot", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeXML(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<a-greeting>
  <firstName>%s</firstName>
  <lastName>%s</lastName>
</a-greeting>`, firstName, lastName))
	})

	// /xmlWithMinusInChild — hyphenated child elements
	mux.HandleFunc("/xmlWithMinusInChild", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeXML(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<greeting>
  <your-firstName>%s</your-firstName>
  <your-lastName>%s</your-lastName>
</greeting>`, firstName, lastName))
	})

	// /xmlWithUnderscoreInChild — underscored child elements
	mux.HandleFunc("/xmlWithUnderscoreInChild", func(w http.ResponseWriter, r *http.Request) {
		firstName := r.URL.Query().Get("firstName")
		lastName := r.URL.Query().Get("lastName")
		writeXML(w, fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<greeting>
  <your_firstName>%s</your_firstName>
  <your_lastName>%s</your_lastName>
</greeting>`, firstName, lastName))
	})

	// /utf8-body-xml — UTF-8 XML
	mux.HandleFunc("/utf8-body-xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><value>啊 ☆</value>`))
	})

	// /shopping — XML shopping categories
	mux.HandleFunc("/shopping", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<?xml version="1.0" encoding="UTF-8"?>
<shopping>
  <category type="groceries">
    <item>Chocolate</item>
    <item>Coffee</item>
  </category>
  <category type="supplies">
    <item>Paper</item>
    <item>Pens</item>
  </category>
  <category type="present">
    <item>Mango</item>
  </category>
</shopping>`)
	})

	// /videos — formatted XML
	mux.HandleFunc("/videos", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<?xml version="1.0" encoding="UTF-8"?>
<music>
  <title>Bohemian Rhapsody</title>
  <artist>Queen</artist>
</music>`)
	})

	// /videos-not-formatted — compact XML
	mux.HandleFunc("/videos-not-formatted", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<?xml version="1.0" encoding="UTF-8"?><music><title>Bohemian Rhapsody</title><artist>Queen</artist></music>`)
	})

	// /namespace-example — XML with namespace
	mux.HandleFunc("/namespace-example", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<?xml version="1.0" encoding="UTF-8"?>
<ns:bookstore xmlns:ns="http://localhost/">
  <ns:book category="reference">
    <ns:title lang="en">Sayings of the Century</ns:title>
    <ns:author>Nigel Rees</ns:author>
    <ns:price>8.95</ns:price>
  </ns:book>
  <ns:book category="fiction">
    <ns:title lang="en">Sword of Honour</ns:title>
    <ns:author>Evelyn Waugh</ns:author>
    <ns:price>12.99</ns:price>
  </ns:book>
</ns:bookstore>`)
	})

	// /namespace-example2 — SOAP-style namespace
	mux.HandleFunc("/namespace-example2", func(w http.ResponseWriter, r *http.Request) {
		writeXML(w, `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
  <soapenv:Body>
    <ns:greeting xmlns:ns="http://localhost/">Hello</ns:greeting>
  </soapenv:Body>
</soapenv:Envelope>`)
	})

	// /textXML — text/xml content type
	mux.HandleFunc("/textXML", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		_, _ = w.Write([]byte(`<xml>something</xml>`))
	})

	// -------------------------------------------------------------------------
	// MULTI-VALUE PARAMS  (unlock ParamITest, URLEncodingITest)
	// -------------------------------------------------------------------------

	// /multiValueParam — echo multi-value list param
	mux.HandleFunc("/multiValueParam", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		values := r.URL.Query()["list"]
		if len(values) == 0 {
			values = r.PostForm["list"]
		}
		writeJSON(w, map[string]string{"list": strings.Join(values, ",")})
	})

	// /threeMultiValueParam — three multi-value form params
	mux.HandleFunc("/threeMultiValueParam", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		writeJSON(w, map[string]string{
			"list":  strings.Join(r.PostForm["list"], ","),
			"list2": strings.Join(r.PostForm["list2"], ","),
			"list3": strings.Join(r.PostForm["list3"], ","),
		})
	})

	// /noValueParam — params without values
	mux.HandleFunc("/noValueParam", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var parts []string
			for k, v := range r.URL.Query() {
				parts = append(parts, fmt.Sprintf("%s=%s", k, strings.Join(v, "")))
			}
			writeJSON(w, map[string]string{"params": strings.Join(parts, ", ")})
		case http.MethodPost:
			_ = r.ParseForm()
			var parts []string
			for k, v := range r.PostForm {
				parts = append(parts, fmt.Sprintf("%s=%s", k, strings.Join(v, "")))
			}
			writeJSON(w, map[string]string{"params": strings.Join(parts, ", ")})
		default:
			// PUT, PATCH — return OK only if params have no values
			_ = r.ParseForm()
			for _, v := range r.Form {
				for _, val := range v {
					if val != "" {
						http.Error(w, "params must have no values", http.StatusInternalServerError)
						return
					}
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		}
	})

	// -------------------------------------------------------------------------
	// REFLECTION / ECHO  (unlock many body and content-type tests)
	// -------------------------------------------------------------------------

	// /reflect — echoes body with same content-type, mirrors cookies
	mux.HandleFunc("/reflect", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		ct := r.Header.Get("Content-Type")
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		for _, c := range r.Cookies() {
			http.SetCookie(w, c)
		}
		_, _ = w.Write(body)
	})

	// /param-reflect — echoes all form params as JSON
	mux.HandleFunc("/param-reflect", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		params := make(map[string]any)
		for k, v := range r.PostForm {
			if len(v) == 1 {
				params[k] = v[0]
			} else {
				params[k] = v
			}
		}
		writeJSON(w, params)
	})

	// /body — raw body echo (POST, PUT, PATCH, DELETE)
	mux.HandleFunc("/body", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(body)
	})

	// /returnContentTypeAsBody — echoes the request Content-Type header as the response body (text/plain).
	mux.HandleFunc("/returnContentTypeAsBody", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(ct))
	})

	// /uuid — returns a JSON object with a UUID-formatted string field.
	mux.HandleFunc("/uuid", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"id": "e8db9168-8d34-4c9a-bc16-04e6bebe3f8d"})
	})

	// -------------------------------------------------------------------------
	// HEADERS  (unlock HeaderITest)
	// -------------------------------------------------------------------------

	// /multiValueHeader — response with duplicate header names
	mux.HandleFunc("/multiValueHeader", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("MultiHeader", "Value 1")
		w.Header().Add("MultiHeader", "Value 2")
		w.WriteHeader(http.StatusOK)
	})

	// /headersWithValues — returns all request headers as JSON object with arrays
	mux.HandleFunc("/headersWithValues", func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]any)
		for k, v := range r.Header {
			headers[k] = v
		}
		writeJSON(w, headers)
	})

	// /multiHeaderReflect — mirrors all request headers into response headers
	mux.HandleFunc("/multiHeaderReflect", func(w http.ResponseWriter, r *http.Request) {
		for k, vals := range r.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	// /header — returns list of header names
	mux.HandleFunc("/header", func(w http.ResponseWriter, r *http.Request) {
		var names []string
		for k := range r.Header {
			names = append(names, k)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(strings.Join(names, ", ")))
	})

	// -------------------------------------------------------------------------
	// COOKIES  (unlock CookieITest)
	// -------------------------------------------------------------------------

	// /multiCookie — sets 2 cookies with same name, different attributes
	mux.HandleFunc("/multiCookie", func(w http.ResponseWriter, r *http.Request) {
		// Use raw Set-Cookie headers to allow duplicate names with different attrs
		w.Header().Add("Set-Cookie", "cookie1=cookieValue1; Domain=localhost")
		w.Header().Add("Set-Cookie", "cookie1=cookieValue2; Path=/; Max-Age=1234567; Domain=localhost; Secure; Version=1")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("OK"))
	})

	// /setCookies — sets 3 named cookies
	mux.HandleFunc("/setCookies", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "key1", Value: "value1", Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "key2", Value: "value2", Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: "key3", Value: "value3", Path: "/"})
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})

	// /setCommonIdCookies — sets 3 cookies with same name
	mux.HandleFunc("/setCommonIdCookies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "key1=value1; Path=/")
		w.Header().Add("Set-Cookie", "key1=value2; Path=/")
		w.Header().Add("Set-Cookie", "key1=value3; Path=/")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})

	// /multiCookieRequest — returns all incoming request cookies as JSON array
	mux.HandleFunc("/multiCookieRequest", func(w http.ResponseWriter, r *http.Request) {
		var cookies []map[string]string
		for _, c := range r.Cookies() {
			cookies = append(cookies, map[string]string{"name": c.Name, "value": c.Value})
		}
		if cookies == nil {
			cookies = []map[string]string{}
		}
		writeJSON(w, cookies)
	})

	// /cookiesWithValues — returns incoming cookies with all attributes as JSON
	mux.HandleFunc("/cookiesWithValues", func(w http.ResponseWriter, r *http.Request) {
		var result []map[string]any
		for _, c := range r.Cookies() {
			result = append(result, map[string]any{
				"name":    c.Name,
				"value":   c.Value,
				"path":    c.Path,
				"domain":  c.Domain,
				"secure":  c.Secure,
				"maxAge":  c.MaxAge,
				"version": 0,
			})
		}
		if result == nil {
			result = []map[string]any{}
		}
		writeJSON(w, result)
	})

	// /cookie — echoes cookie names for all HTTP methods
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		var names []string
		for _, c := range r.Cookies() {
			names = append(names, c.Name)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(strings.Join(names, ",")))
	})

	// /html_with_cookie — returns HTML and sets JSESSIONID
	mux.HandleFunc("/html_with_cookie", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "JSESSIONID",
			Value: "B3134D534F40968A3805968207273EF5",
			Path:  "/",
		})
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<body>HTML with cookie</body>`))
	})

	// -------------------------------------------------------------------------
	// STATUS CODES  (unlock ErrorMessageITest, LoggingITest)
	// -------------------------------------------------------------------------

	// /409
	mux.HandleFunc("/409", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("ERROR"))
	})

	// /statusCode500
	mux.HandleFunc("/statusCode500", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("An expected error occurred"))
	})

	// /statusCode409WithNoBody
	mux.HandleFunc("/statusCode409WithNoBody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
	})

	// /emptyBody
	mux.HandleFunc("/emptyBody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	// -------------------------------------------------------------------------
	// CONTENT TYPE VARIANTS  (unlock ContentTypeITest)
	// -------------------------------------------------------------------------

	// /textHTML — HTML with title and paragraphs (formatted)
	mux.HandleFunc("/textHTML", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>my title</title></head>
<body>
  <p>paragraph 1</p>
  <p>paragraph 2</p>
</body>
</html>`))
	})

	// /textHTML-not-formatted — compact HTML
	mux.HandleFunc("/textHTML-not-formatted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>my title</title></head><body><p>paragraph 1</p><p>paragraph 2</p></body></html>`))
	})

	// /mimeTypeWithPlusJson — application/something+json
	mux.HandleFunc("/mimeTypeWithPlusJson", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/something+json")
		_, _ = w.Write([]byte(`{"message":"It works"}`))
	})

	// /mimeTypeWithPlusXml — application/something+xml
	mux.HandleFunc("/mimeTypeWithPlusXml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/something+xml")
		_, _ = w.Write([]byte(`<body><message>Custom mime-type ending with +xml</message></body>`))
	})

	// /mimeTypeWithPlusHtml — application/something+html
	mux.HandleFunc("/mimeTypeWithPlusHtml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/something+html")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>my title</title></head><body><p>p1</p><p>p2</p></body></html>`))
	})

	// /noContentTypeJsonCompatible — no Content-Type header, JSON body
	mux.HandleFunc("/noContentTypeJsonCompatible", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":"It works"}`))
	})

	// /contentTypeJsonButBodyIsNotJson — JSON content-type but non-JSON body
	mux.HandleFunc("/contentTypeJsonButBodyIsNotJson", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("This is not JSON"))
	})

	// /contentTypeAsBody — returns the incoming Content-Type as the response body
	mux.HandleFunc("/contentTypeAsBody", func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, ct)
	})

	// /customMimeType — custom non-standard content type
	mux.HandleFunc("/customMimeType", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/something-custom")
		_, _ = w.Write([]byte(`<body><message>Custom mime-type</message></body>`))
	})

	// /customMimeTypeJsonCompatible — vendor JSON type
	mux.HandleFunc("/customMimeTypeJsonCompatible", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.uoml+json")
		_, _ = w.Write([]byte(`{"message":"It works"}`))
	})

	// /rss — RSS feed
	mux.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test RSS</title>
    <item><title>Item 1</title></item>
  </channel>
</rss>`))
	})

	// /contentTypeButNoBody — JSON content-type with empty body
	mux.HandleFunc("/contentTypeButNoBody", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	// -------------------------------------------------------------------------
	// GZIP VARIANTS  (unlock GzipITest)
	// -------------------------------------------------------------------------

	// /gzip-empty-body — gzip-encoded empty body
	mux.HandleFunc("/gzip-empty-body", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_ = gw.Close()
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(buf.Bytes())
	})

	// /gzip-json — gzip-encoded JSON (alias matching Java endpoint name)
	mux.HandleFunc("/gzip-json", func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_, _ = gw.Write([]byte(`{"hello":"Hello Scalatra"}`))
		_ = gw.Close()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(buf.Bytes())
	})

	// -------------------------------------------------------------------------
	// SESSION ID  (unlock SessionIdITest)
	// -------------------------------------------------------------------------

	// /sessionId — issues JSESSIONID if not present; validates if present
	mux.HandleFunc("/sessionId", func(w http.ResponseWriter, r *http.Request) {
		const fixedSession = "1234"
		cookie, err := r.Cookie("jsessionid")
		if err != nil || cookie.Value == "" {
			http.SetCookie(w, &http.Cookie{Name: "jsessionid", Value: fixedSession, Path: "/"})
			writeJSON(w, map[string]string{"status": "session_created", "sessionId": fixedSession})
			return
		}
		if cookie.Value == fixedSession {
			writeJSON(w, map[string]string{"status": "Success"})
			return
		}
		w.WriteHeader(http.StatusConflict)
		writeJSON(w, map[string]string{"status": "invalid_session"})
	})

	// -------------------------------------------------------------------------
	// FORM AUTH / SPRING SECURITY SIMULATION  (unlock CsrfITest, AuthenticationITest)
	// -------------------------------------------------------------------------

	const csrfFormToken = "8adf2ea1-b246-40aa-8e13-a85fb7914341"
	const csrfHeaderToken = "ab8722b1-1f23-4dcf-bf63-fb8b94be4107"

	// /loginPageWithCsrf — GET returns login form with _csrf field
	mux.HandleFunc("/loginPageWithCsrf", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Type", "text/html")
			return
		}
		if r.Method == http.MethodPost {
			_ = r.ParseForm()
			token := r.FormValue("_csrf")
			if token != csrfFormToken {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
			writeJSON(w, map[string]string{"status": "logged_in"})
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><body>
<form method="POST" action="/loginPageWithCsrf">
  <input type="hidden" name="_csrf" value="%s"/>
  <input type="text" name="j_username"/>
  <input type="password" name="j_password"/>
</form></body></html>`, csrfFormToken)
	})

	// /j_spring_security_check — Spring Security form login
	mux.HandleFunc("/j_spring_security_check", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		user := r.FormValue("j_username")
		pass := r.FormValue("j_password")
		if user == "John" && pass == "Doe" {
			http.SetCookie(w, &http.Cookie{Name: "jsessionid", Value: "1234", Path: "/"})
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("OK"))
			return
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})

	// /session-required — requires valid jsessionid cookie
	mux.HandleFunc("/session-required", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jsessionid")
		if err != nil || cookie.Value != "1234" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		writeJSON(w, map[string]string{"message": "session valid"})
	})

	// /formAuth — returns login form and sets jsessionid
	mux.HandleFunc("/formAuth", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "jsessionid", Value: "1234", Path: "/"})
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<form method="POST"><input name="j_username"/><input name="j_password"/></form>`))
	})

	// /pageWithDefaultHeaderCsrf — returns page with CSRF meta header
	mux.HandleFunc("/pageWithDefaultHeaderCsrf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><head>
<meta name="_csrf_header" content="X-CSRF-TOKEN"/>
<meta name="_csrf" content="%s"/>
</head><body></body></html>`, csrfHeaderToken)
	})

	// /pageThatRequireHeaderCsrf — validates CSRF sent as header
	mux.HandleFunc("/pageThatRequireHeaderCsrf", func(w http.ResponseWriter, r *http.Request) {
		headerName := r.URL.Query().Get("headerName")
		if headerName == "" {
			headerName = "X-CSRF-TOKEN"
		}
		token := r.Header.Get(headerName)
		if token != csrfHeaderToken {
			http.Error(w, "Failed", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("OK"))
	})

	// -------------------------------------------------------------------------
	// PATH PARAMETERS  (unlock PathParamITest)
	// Note: uses /greeting/ prefix to avoid routing conflicts.
	// Go integration tests reference these URLs directly.
	// -------------------------------------------------------------------------

	// /greeting/{firstName}/{lastName}
	mux.HandleFunc("/greeting/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/greeting/"), "/")
		switch len(parts) {
		case 2:
			firstName, lastName := parts[0], parts[1]
			writeJSON(w, map[string]string{
				"firstName": firstName,
				"lastName":  lastName,
				"fullName":  firstName + " " + lastName,
			})
		case 3:
			firstName, middleName, lastName := parts[0], parts[1], parts[2]
			writeJSON(w, map[string]string{
				"firstName":  firstName,
				"middleName": middleName,
				"lastName":   lastName,
			})
		default:
			http.NotFound(w, r)
		}
	})

	// -------------------------------------------------------------------------
	// MISCELLANEOUS UTILITY ENDPOINTS
	// -------------------------------------------------------------------------

	// /getWithContent — validates request has body content
	mux.HandleFunc("/getWithContent", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.TrimSpace(string(body)) == "hullo" {
			writeJSON(w, map[string]string{"status": "ok"})
			return
		}
		http.Error(w, "No or incorrect content", http.StatusBadRequest)
	})

	// /requestUrl — returns the request URL as JSON
	mux.HandleFunc("/requestUrl", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"url": r.RequestURI})
	})

	// /something.json — simple value response
	mux.HandleFunc("/something.json", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"value": "something"})
	})

	// -------------------------------------------------------------------------
	// HTTP METHOD VARIANTS  (unlock PutITest, PatchITest, DeleteITest, OptionsITest)
	// -------------------------------------------------------------------------

	// /greetPut — PUT greeting from query or form params
	mux.HandleFunc("/greetPut", func(w http.ResponseWriter, r *http.Request) {
		var firstName, lastName string
		if r.Method == http.MethodPut {
			_ = r.ParseForm()
			firstName = r.FormValue("firstName")
			lastName = r.FormValue("lastName")
		}
		if firstName == "" {
			firstName = r.URL.Query().Get("firstName")
		}
		if lastName == "" {
			lastName = r.URL.Query().Get("lastName")
		}
		writeJSON(w, greetingJSON(firstName, lastName))
	})

	// /greetPatch — PATCH greeting from query or form params
	mux.HandleFunc("/greetPatch", func(w http.ResponseWriter, r *http.Request) {
		var firstName, lastName string
		if r.Method == http.MethodPatch {
			_ = r.ParseForm()
			firstName = r.FormValue("firstName")
			lastName = r.FormValue("lastName")
		}
		if firstName == "" {
			firstName = r.URL.Query().Get("firstName")
		}
		if lastName == "" {
			lastName = r.URL.Query().Get("lastName")
		}
		writeJSON(w, greetingJSON(firstName, lastName))
	})

	// /binaryBody — echoes request body bytes as comma-separated decimal values
	mux.HandleFunc("/binaryBody", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		parts := make([]string, len(body))
		for i, b := range body {
			parts[i] = strconv.Itoa(int(b))
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(strings.Join(parts, ", ")))
	})

	// /jsonGreet — accepts JSON body with firstName/lastName, returns fullName
	mux.HandleFunc("/jsonGreet", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		fullName := payload["firstName"] + " " + payload["lastName"]
		writeJSON(w, map[string]string{"fullName": fullName})
	})

	// /returnBodyAsBody — echoes the request body for any method (used by OptionsITest)
	mux.HandleFunc("/returnBodyAsBody", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(body)
	})

	// -------------------------------------------------------------------------
	// ACCEPT HEADER  (unlock AcceptHeaderITest)
	// -------------------------------------------------------------------------

	// /jsonBodyAcceptHeader — reads JSON body, returns just the message string
	mux.HandleFunc("/jsonBodyAcceptHeader", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(payload["message"]))
	})

	// -------------------------------------------------------------------------
	// BOM SUPPORT  (unlock BomITest)
	// -------------------------------------------------------------------------

	// /xmlWithBom — returns an XML document prefixed with a UTF-8 BOM
	mux.HandleFunc("/xmlWithBom", func(w http.ResponseWriter, r *http.Request) {
		// UTF-8 BOM: 0xEF 0xBB 0xBF
		bom := []byte{0xEF, 0xBB, 0xBF}
		xmlBody := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project>
  <target name="build"/>
</project>`)
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write(append(bom, xmlBody...))
	})

	// -------------------------------------------------------------------------
	// EXTRACT TESTS  (unlock GivenWhenThenExtractITest)
	// -------------------------------------------------------------------------

	// /products — list of 2 products with dimensions
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []map[string]any{
			{
				"id": 2, "name": "An ice sculpture", "price": 12.5,
				"dimensions": map[string]any{"length": 7.0, "width": 12.0, "height": 9.5},
			},
			{
				"id": 3, "name": "A blue mouse", "price": 25.5,
				"dimensions": map[string]any{"length": 3.1, "width": 1.0, "height": 1.0},
			},
		})
	})

	return httptest.NewServer(mux)
}
