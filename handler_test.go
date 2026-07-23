package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// The body is a pull request event. Happy path. We expect a status 201
func TestEventHandler_HandleValidPayload(t *testing.T) {
	c := make(chan PullRequestEvent, 1)
	eh := EventHandler{channel: c}

	rr, err := request("test/fixtures/hook-pull-request-fulfilled.json", eh.Handle())
	if err != nil {
		t.Fatal(err)
	}

	event := <-c
	if &event.Repository == nil {
		t.Error("repository must be defined")
	}

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}

// The body is a pull request event but state is not set to MERGED. We expect a status 422
func TestEventHandler_HandleUnsupportedState(t *testing.T) {
	c := make(chan PullRequestEvent, 1)
	eh := EventHandler{channel: c}

	rr, err := request("test/fixtures/hook-pull-request-created.json", eh.Handle())
	if err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusUnprocessableEntity {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnprocessableEntity)
	}
}

// The body is a push event (not a pull request event). We expect a status 400
func TestEventHandler_HandleUnsupportedEventType(t *testing.T) {
	c := make(chan PullRequestEvent, 1)
	eh := EventHandler{channel: c}

	rr, err := request("test/fixtures/hook-pr-merged-develop.json", eh.Handle())
	if err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// The body contains some random model that we cannot deserialize. We expect a status 400.
func TestEventHandler_HandleBadRequest(t *testing.T) {
	c := make(chan PullRequestEvent, 1)
	eh := EventHandler{channel: c}

	rr, err := request("test/fixtures/hook-bad-request.json", eh.Handle())
	if err != nil {
		t.Fatal(err)
	}

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

// The query token matches the configured one. We expect the next handler to run.
func TestEventHandler_CheckTokenValid(t *testing.T) {
	rr, called := tokenRequest("secret", "/hook?token=secret")

	if !called {
		t.Error("next handler must be called when the token matches")
	}

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// The query token differs from the configured one. We expect a status 403.
func TestEventHandler_CheckTokenInvalid(t *testing.T) {
	rr, called := tokenRequest("secret", "/hook?token=wrong")

	if called {
		t.Error("next handler must not be called when the token differs")
	}

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}

// A token is configured but the query has none. We expect a status 403.
func TestEventHandler_CheckTokenMissing(t *testing.T) {
	rr, called := tokenRequest("secret", "/hook")

	if called {
		t.Error("next handler must not be called when the token is missing")
	}

	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusForbidden)
	}
}

// No token is configured, so the endpoint is open. Documents the default.
func TestEventHandler_CheckTokenNotConfigured(t *testing.T) {
	_, called := tokenRequest("", "/hook")

	if !called {
		t.Error("next handler must be called when no token is configured")
	}
}

func tokenRequest(token, target string) (*httptest.ResponseRecorder, bool) {
	called := false
	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
	})

	eh := EventHandler{channel: make(chan PullRequestEvent, 1)}
	req := httptest.NewRequest("POST", target, nil)

	rr := httptest.NewRecorder()
	eh.CheckToken(token, next).ServeHTTP(rr, req)

	return rr, called
}

func request(filename string, hf http.Handler) (*httptest.ResponseRecorder, error) {
	file, _ := os.Open(filename)
	req, err := http.NewRequest("POST", "/hook", file)
	if err != nil {
		return nil, err
	}

	rr := httptest.NewRecorder()
	hf.ServeHTTP(rr, req)

	return rr, nil
}
