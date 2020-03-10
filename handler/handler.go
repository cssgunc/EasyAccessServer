package handler

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	firestore "cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

// Config is the config for the handler.
type Config struct {
	Logger *logrus.Logger
}

// Handler is the global handler for the api.
type Handler struct {
	http.Handler

	logger *logrus.Logger
}

func isValidConfig(c Config) error {
	if c.Logger == nil {
		return errors.New("logger cannot be nil")
	}
	return nil
}

var app *firebase.App
var client *firestore.Client

func (h *Handler) setUpApp() {
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

// New returns a new handler.
func New(c Config) (*Handler, error) {
	if err := isValidConfig(c); err != nil {
		return nil, errors.Wrap(err, "invalid handler config")
	}

	h := Handler{
		logger: c.Logger,
	}

	r := chi.NewRouter()

	// Middleware set up
	r.Use(middleware.DefaultCompress)
	r.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	h.setUpApp()
	r.Route("/", func(r chi.Router) {
		// set up routes
		r.Post("/user", h.AuthUser)
		r.Patch("/updateUser", h.updateUser)
		r.Post("/collegeMatches", h.getMatches)
		r.Get("/majors", h.collegeMajors)
		r.Get("/updateSelectivityScores", h.updateSelectivityScores)
		r.Get("/updateSchoolNeedMet", h.updateSchoolNeedMet)
		r.Post("/addUserInfo", h.addUserInfo)
		r.Get("/pastMatches", h.getPastMatches)
		r.Get("/test", h.testOtherFunc)
		r.Get("/updateMajor", h.updateMajorInfo)
	})
	r.Get("/health", h.health)

	h.Handler = r
	return &h, nil
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	// Add any DB, Redis, or server pings here to have a full health check.
	render.JSON(w, r, struct {
		Health string `json:"health"`
	}{
		Health: "OK",
	})
}
