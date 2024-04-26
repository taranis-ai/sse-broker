package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tmaxmax/go-sse"
)

func TestValidateAPIKey(t *testing.T) {
	assert.True(t, validateAPIKey("testkey", "testkey"), "API key validation should return true for matching keys")
	assert.False(t, validateAPIKey("testkey", "wrongkey"), "API key validation should return false for non-matching keys")
}

func TestPublisher(t *testing.T) {
	config := Config{
		APIKey:      "testkey",
		PublishPath: "/publish",
	}

	app := &App{
		Config:    config,
		SSEServer: &sse.Server{},
	}

	// Create a message
	message := Message{
		Data:  "Hello, world!",
		Event: "testEvent",
	}
	messageBody, _ := json.Marshal(message)

	req, err := http.NewRequest("POST", "/publish", bytes.NewBuffer(messageBody))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", "testkey")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.publisher)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)

	expected := ``
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestEventsEndpoint(t *testing.T) {
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.zbgd5BNF1cqQ_prCEqIvBTjSxMS8bDLnJAE_wE-0Cxg"

	config := Config{
		JWTSecretKey: "supersecret",
		SSEPath:      "/events",
		Topics:       []string{"test"},
	}

	app := &App{
		Config: config,
		SSEServer: &sse.Server{
			OnSession: nil,
		},
	}

	app.SSEServer.OnSession = app.onSSESession

	req, err := http.NewRequest("GET", "/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", validToken)

	rr := httptest.NewRecorder()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure resources are cleaned up after the test

	req = req.WithContext(ctx)

	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			app.SSEServer.ServeHTTP(w, r)
		})
		handler.ServeHTTP(rr, req)
	}()

	time.Sleep(100 * time.Millisecond)

	testMessage := Message{
		Data:  "Test message",
		Event: "test",
	}
	app.publishMessage(testMessage)

	time.Sleep(100 * time.Millisecond)

	responseContent := rr.Body.String()

	assert.Equal(t, http.StatusOK, rr.Code, "SSE connection should be authorized with a valid JWT token")
	assert.Contains(t, responseContent, "Test message", "The response content should contain the test message")

	rr.Result().Body.Close()
}
