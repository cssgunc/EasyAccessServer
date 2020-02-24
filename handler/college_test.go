package handler

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	firebase "firebase.google.com/go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

const credEnvVar = "GOOGLE_APPLICATION_CREDENTIALS"

func Router() *chi.Mux {
	h := Handler{
		logger: logrus.New(),
	}
	r := chi.NewRouter()

	// Middleware set up
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.Recoverer)
	setUpApp()
	r.Route("/", func(r chi.Router) {
		// set up routes
		r.Post("/user", h.AuthUser)
		r.Patch("/updateUser", h.updateUser)
	})
	return r
}
func setUpApp() {
	current := os.Getenv(credEnvVar)

	if err := os.Setenv(credEnvVar, "/Users/BaileyFrederick/Downloads/service-account-file.json"); err != nil {
		log.Fatal(err)
	}
	defer os.Setenv(credEnvVar, current)
	ProjectID := os.Getenv("ProjectID")
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}

func TestQueryColleges(t *testing.T) {
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
