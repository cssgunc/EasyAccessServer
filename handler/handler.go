package handler

import (
	"net/http"
	"fmt"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	h.setUpApp()

	r.Route("/", func(r chi.Router) {
		// set up routes
		r.Get("/", h.hello)
		r.Post("/user", h.authUser)
		r.Get("/matches", h.getMatches)
		r.Get("/colleges", h.getColleges)
		//r.Patch("/updateUser", h.updateUser)
	})
	r.Get("/health", h.health)
	

	h.Handler = r
	return &h, nil
}

func (h *Handler) hello(w http.ResponseWriter, r *http.Request) {
	// Add any DB, Redis, or server pings here to have a full health check.
	fmt.Fprintln(w, "Hello World!")
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	// Add any DB, Redis, or server pings here to have a full health check.
	render.JSON(w, r, struct {
		Health string `json:"health"`
	}{
		Health: "OK",
	})
}
