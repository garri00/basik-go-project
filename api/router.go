package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"basic-go-project/api/handlers"
	md "basic-go-project/api/middleware"
	"basic-go-project/api/routes"
)

const RequestTimeout = 60

var skipPaths = []string{"/ping"}

type Router struct {
	postgresClient *pgxpool.Pool
	log            *zerolog.Logger
}

// NewRouter defines new router instance
func NewRouter(postgresClient *pgxpool.Pool, log *zerolog.Logger) *chi.Mux {
	router := &Router{
		postgresClient: postgresClient,
		log:            log,
	}

	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(md.Logger(router.log, skipPaths))
	r.Use(middleware.Recoverer)

	//Add cors
	md.NewDefaultCors(r)

	// Set timeout for incoming requests
	r.Use(middleware.Timeout(RequestTimeout * time.Second))

	//Set basic hello handler
	r.Get("/hello", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			router.log.Error().Err(fmt.Errorf("error writing response: %w", err)).Send()

			return
		}
		w.WriteHeader(http.StatusOK)
	})

	//health-check returns basic info status of service
	r.Get("/health-check", handlers.HealthCheck)

	//Mounts /api routes to main router
	r.Mount("/api", routes.API(router.postgresClient, router.log))

	return r
}