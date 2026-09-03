package testserver_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
	"github.com/jcfrperu/goaneco-rest-ensured/testserver"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestServerEndpoints(t *testing.T) {
	t.Parallel()
	ts := testserver.NewTestServer()
	t.Cleanup(ts.Close)

	t.Run("GET /hello returns greeting JSON", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		must := require.New(t)

		resp := rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/hello").
			Then().
			AssertWith(t).
			StatusCode(200).
			ContentType(rest.ContentTypeJSON).
			Body("message", "Hello World").
			Extract().
			Response()

		must.NoError(resp.Err())
		is.Equal(http.StatusOK, resp.StatusCode())
	})

	t.Run("GET /status/{code} returns exact status code", func(t *testing.T) {
		t.Parallel()
		statusCodes := []int{200, 201, 204, 400, 401, 403, 404, 500}
		for _, code := range statusCodes {
			rest.Given().
				BaseURI(ts.URL).
				When().
				Get("/status/{0}", code).
				Then().
				AssertWith(t).
				StatusCode(code)
		}
	})

	t.Run("GET /json/users validates users array", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/json/users").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("total", 2).
			Body("users.0.name", "Alice").
			Body("users.0.role", "admin").
			Body("users.0.address.city", "New York").
			BodyContainsElement("users.0.tags", "lead").
			Body("users.1.name", "Bob").
			Body("users.1.active", false)
	})

	t.Run("GET /json/store validates nested store books and bicycle", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/json/store").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("store.bicycle.color", "red").
			Body("store.bicycle.price", 19.95).
			BodyLength("store.book", 2).
			Body("store.book.0.author", "Nigel Rees").
			Body("store.book.1.title", "Sword of Honour")
	})

	t.Run("POST /echo echoes body and custom headers", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Header("X-Custom-Echo", "ValueEcho").
			Body(`{"test":"echo"}`).
			ContentType(rest.ContentTypeJSON).
			When().
			Post("/echo").
			Then().
			AssertWith(t).
			StatusCode(200).
			Header("X-Custom-Echo", "ValueEcho").
			Body("test", "echo")
	})

	t.Run("POST /form urlencoded form submission", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "john_doe").
			FormParam("action", "login").
			When().
			Post("/form").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("username", "john_doe").
			Body("action", "login")
	})

	t.Run("GET & POST /cookies sets and sends cookies", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Set cookie
		postResp := rest.Given().
			BaseURI(ts.URL).
			QueryParam("name", "auth_session").
			QueryParam("value", "secret-token-777").
			When().
			Post("/cookies").
			Then().
			AssertWith(t).
			StatusCode(200).
			Extract().
			Response()

		authCookie := postResp.Cookie("auth_session")
		is.Equal("secret-token-777", authCookie)

		// Send cookie
		rest.Given().
			BaseURI(ts.URL).
			Cookie("user_pref", "dark_mode").
			When().
			Get("/cookies").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("user_pref", "dark_mode")
	})

	t.Run("POST /auth/basic authentication check", func(t *testing.T) {
		t.Parallel()
		// Valid credentials
		rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("admin", "secret").
			When().
			Post("/auth/basic").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("status", "authenticated").
			Body("user", "admin")

		// Invalid credentials
		rest.Given().
			BaseURI(ts.URL).
			Auth().Basic("admin", "wrong").
			When().
			Post("/auth/basic").
			Then().
			AssertWith(t).
			StatusCode(401)
	})

	t.Run("POST /auth/bearer token check", func(t *testing.T) {
		t.Parallel()
		// Valid bearer token
		rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("secret-token-123").
			When().
			Post("/auth/bearer").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("status", "authenticated")

		// Invalid token
		rest.Given().
			BaseURI(ts.URL).
			Auth().OAuth2("wrong-token").
			When().
			Post("/auth/bearer").
			Then().
			AssertWith(t).
			StatusCode(401)
	})

	t.Run("GET /delay/{ms} validates response time assertions", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/delay/50").
			Then().
			AssertWith(t).
			StatusCode(200).
			TimeLessThan(2 * time.Second)
	})

	t.Run("POST /upload file and form field submission", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			FormParam("tag", "report").
			MultiPartNamed("file1", "document.pdf", []byte("%PDF-1.4..."), "application/pdf").
			When().
			Post("/upload").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("fields.tag", "report").
			BodyContainsElement("files", "document.pdf")
	})

	t.Run("GET /headers inspects received headers", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			Header("X-Request-Trace", "trace-987").
			When().
			Get("/headers").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("X-Request-Trace.0", "trace-987")
	})

	t.Run("GET /xml/store returns valid XML payload", func(t *testing.T) {
		t.Parallel()
		resp := rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/xml/store").
			Then().
			AssertWith(t).
			StatusCode(200).
			ContentType(rest.ContentTypeXML).
			Extract().
			Response()

		is := assert.New(t)
		must := require.New(t)
		xp, err := resp.XmlPath()
		must.NoError(err)
		is.Equal("Nigel Rees", xp.GetString("//book[@category='reference']/author"))
	})

	t.Run("GET /gzip/data and /deflate/data compression endpoints", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/gzip/data").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("compressed", true).
			Body("codec", "gzip")

		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/deflate/data").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("compressed", true).
			Body("codec", "deflate")
	})

	t.Run("Redirect chain /redirect/step1", func(t *testing.T) {
		t.Parallel()
		rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/redirect/step1").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("status", "redirect_complete")
	})

	t.Run("CSRF flow /csrf/page and /csrf/submit", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)
		pageResp := rest.Given().
			BaseURI(ts.URL).
			When().
			Get("/csrf/page").
			Then().
			AssertWith(t).
			StatusCode(200).
			Extract().
			AsString()

		token := rest.ExtractCsrfFromHTML(pageResp, "_csrf")
		is.Equal("token-abc-987", token)

		rest.Given().
			BaseURI(ts.URL).
			FormParam("_csrf", token).
			FormParam("data", "user-input").
			When().
			Post("/csrf/submit").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("status", "csrf_accepted")
	})

	t.Run("Form login /login and /secured session flow", func(t *testing.T) {
		t.Parallel()
		loginResp := rest.Given().
			BaseURI(ts.URL).
			FormParam("username", "john").
			FormParam("password", "doe").
			When().
			Post("/login").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("status", "logged_in").
			Extract().
			Response()

		sessCookie := loginResp.Cookie("JSESSIONID")
		is := assert.New(t)
		is.NotEmpty(sessCookie)

		rest.Given().
			BaseURI(ts.URL).
			Cookie("JSESSIONID", sessCookie).
			When().
			Get("/secured").
			Then().
			AssertWith(t).
			StatusCode(200).
			Body("secret", "classified")
	})
}
