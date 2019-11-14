package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	firestore "cloud.google.com/go/firestore"

	"github.com/stretchr/testify/assert"
)

func TestAuthUser(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		description        string
		url                string
		method             string
		body               string
		expectedStatusCode int
	}{
		{
			description:        "Incorrect http method",
			url:                "/user",
			method:             "GET",
			expectedStatusCode: 405,
		},
		{
			description:        "Missing body",
			url:                "/user",
			method:             "POST",
			body:               "{test;",
			expectedStatusCode: 500,
		},
		{
			description:        "Incorrect idToken",
			url:                "/user",
			method:             "POST",
			body:               `"h48hd9398sg43"`,
			expectedStatusCode: 404,
		},
		{
			description:        "Correct idToken",
			url:                "/user",
			method:             "POST",
			body:               `"xLwd4c1WjKaxG3Vf3GDVMXMTLFE3"`,
			expectedStatusCode: 200,
		},
	}

	for _, tc := range tests {
		body := bytes.NewReader([]byte(tc.body))
		request, err := http.NewRequest(tc.method, tc.url, body)
		assert.NoError(err)
		response := httptest.NewRecorder()
		Router().ServeHTTP(response, request)
		t.Log("XXX", response.Body)
		assert.Equal(tc.expectedStatusCode, response.Code, tc.description)
	}
}

func TestUpdateUser(t *testing.T) {
	var update []firestore.Update
	update = append(update, firestore.Update{
		Path:  "name",
		Value: "test",
	})
	info := updateInfo{
		UID:  "xLwd4c1WjKaxG3Vf3GDVMXMTLFE3",
		Info: update,
	}
	badInfo := updateInfo{
		UID:  "xLwd4c1WjKaVf3GDVMXM3",
		Info: update,
	}
	assert := assert.New(t)
	tests := []struct {
		description        string
		url                string
		method             string
		body               updateInfo
		expectedStatusCode int
	}{
		{
			description:        "Incorrect http method",
			url:                "/updateUser",
			method:             "GET",
			expectedStatusCode: 405,
		},
		{
			description:        "Missing body",
			url:                "/updateUser",
			method:             "PATCH",
			body:               badInfo,
			expectedStatusCode: 404,
		},
		{
			description:        "Correct idToken",
			url:                "/updateUser",
			method:             "PATCH",
			body:               info,
			expectedStatusCode: 200,
		},
	}

	for _, tc := range tests {
		bodyJSON, err := json.Marshal(tc.body)
		body := bytes.NewReader([]byte(bodyJSON))
		request, err := http.NewRequest(tc.method, tc.url, body)
		assert.NoError(err)
		response := httptest.NewRecorder()
		Router().ServeHTTP(response, request)
		t.Log("XXX", response.Body)
		assert.Equal(tc.expectedStatusCode, response.Code, tc.description)
	}
}

func TestGetMatches(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		description        string
		url                string
		method             string
		body               string
		expectedStatusCode int
	}{
		{
			description:        "Incorrect http method",
			url:                "/matches",
			method:             "POST",
			expectedStatusCode: 405,
		},
		{
			description:        "Bad body",
			url:                "/matches",
			method:             "GET",
			body:               "{test;",
			expectedStatusCode: 500,
		},
		{
			description:        "Incorrect uuid",
			url:                "/matches",
			method:             "GET",
			body:               `"h48hd9398sg43"`,
			expectedStatusCode: 404,
		},
		{
			description:        "Correct idToken",
			url:                "/matches",
			method:             "GET",
			body:               `"xLwd4c1WjKaxG3Vf3GDVMXMTLFE3"`,
			expectedStatusCode: 200,
		},
	}

	for _, tc := range tests {
		body := bytes.NewReader([]byte(tc.body))
		request, err := http.NewRequest(tc.method, tc.url, body)
		assert.NoError(err)
		response := httptest.NewRecorder()
		Router().ServeHTTP(response, request)
		t.Log("XXX", tc.description, response.Body)
		assert.Equal(tc.expectedStatusCode, response.Code, tc.description)
	}
}

func TestGetColleges(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		description        string
		url                string
		method             string
		body               string
		expectedStatusCode int
	}{
		{
			description:        "Incorrect http method",
			url:                "/colleges",
			method:             "POST",
			expectedStatusCode: 405,
		},
		{
			description:        "Correct",
			url:                "/colleges",
			method:             "GET",
			expectedStatusCode: 200,
		},
	}

	for _, tc := range tests {
		body := bytes.NewReader([]byte(tc.body))
		request, err := http.NewRequest(tc.method, tc.url, body)
		assert.NoError(err)
		response := httptest.NewRecorder()
		Router().ServeHTTP(response, request)
		t.Log("XXX", tc.description, response.Body)
		assert.Equal(tc.expectedStatusCode, response.Code, tc.description)
	}
}
