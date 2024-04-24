package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmaxmax/go-sse"
)

func setup() {
	setupConfig()
	sseServer = &sse.Server{}
}

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func TestValidateAPIKey(t *testing.T) {
	assert.True(t, validateAPIKey("supersecret"))
	assert.False(t, validateAPIKey("invalid"))
}

func TestValidateJWT(t *testing.T) {
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.zbgd5BNF1cqQ_prCEqIvBTjSxMS8bDLnJAE_wE-0Cxg"
	invalidToken := "thisIsAnInvalidToken"

	token, err := validateJWT(validToken)
	assert.Nil(t, err)
	assert.NotNil(t, token)

	_, err = validateJWT(invalidToken)
	assert.NotNil(t, err)
}

func TestPublisherHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		contentType    string
		apiKey         string
		body           string
		expectedStatus int
	}{
		{"Valid Request", "POST", "application/json", "supersecret", `{"data":"hello","event":"message"}`, http.StatusOK},
		{"Invalid Method", "GET", "application/json", "supersecret", "", http.StatusMethodNotAllowed},
		{"Invalid Content-Type", "POST", "text/plain", "supersecret", "", http.StatusBadRequest},
		{"Invalid API Key", "POST", "application/json", "wrongkey", "", http.StatusUnauthorized},
		{"Bad JSON", "POST", "application/json", "supersecret", "{bad json}", http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := bytes.NewBufferString(test.body)
			req, _ := http.NewRequest(test.method, "/publish", reqBody)
			req.Header.Set("Content-Type", test.contentType)
			req.Header.Set("X-API-KEY", test.apiKey)
			recorder := httptest.NewRecorder()

			publisher(recorder, req)

			assert.Equal(t, test.expectedStatus, recorder.Code)
		})
	}
}
