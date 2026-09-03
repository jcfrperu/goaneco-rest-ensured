package rest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/jcfrperu/goaneco-rest-ensured/rest"
)

func ExampleGiven() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "Hello World", "status": 200}`))
	}))
	defer ts.Close()

	resp := rest.Given().
		BaseURI(ts.URL).
		When().
		Get("/hello").
		Then().
		StatusCode(200).
		Body("message", "Hello World").
		Extract()

	fmt.Println(resp.PathString("message"))
	// Output:
	// Hello World
}

func ExampleRequestSpecBuilder() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "authenticated"}`))
	}))
	defer ts.Close()

	spec := rest.NewRequestSpecBuilder().
		SetBaseURI(ts.URL).
		SetContentType(rest.ContentTypeJSON).
		SetAuth(&rest.OAuth2Scheme{AccessToken: "sample-token"}).
		Build()

	resp := rest.Given().
		Spec(spec).
		Get("/auth").
		Then().
		StatusCode(200).
		Extract()

	fmt.Println(resp.PathString("status"))
	// Output:
	// authenticated
}

func ExampleResponseSpecBuilder() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code": 200, "status": "active"}`))
	}))
	defer ts.Close()

	respSpec := rest.NewResponseSpecBuilder().
		ExpectStatusCode(200).
		ExpectContentType(rest.ContentTypeJSON).
		ExpectBody("status", "active").
		Build()

	rest.Given().
		BaseURI(ts.URL).
		Get("/status").
		Then().
		Spec(respSpec)

	fmt.Println("Validation Passed")
	// Output:
	// Validation Passed
}

func ExampleTimingFilter() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	var durationRecorded bool
	timing := &rest.TimingFilter{
		OnComplete: func(d time.Duration) {
			durationRecorded = d >= 0
		},
	}

	rest.Given().
		BaseURI(ts.URL).
		Filter(timing).
		Get("/ping").
		Then().
		StatusCode(200)

	fmt.Println("Recorded:", durationRecorded)
	// Output:
	// Recorded: true
}
