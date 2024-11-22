package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/btschwartz12/forum/repo"
	"github.com/btschwartz12/forum/server/api/swagger"
)

type ApiServer struct {
	router *chi.Mux
	logger *zap.SugaredLogger
	rpo    *repo.Repo
	token  string
}

func (s *ApiServer) Init(logger *zap.SugaredLogger, rpo *repo.Repo, prefix string, authToken string) error {
	s.logger = logger
	s.router = chi.NewRouter()
	s.rpo = rpo
	s.token = authToken

	s.router.Get("/", http.RedirectHandler(fmt.Sprintf("%s/swagger/index.html", prefix), http.StatusMovedPermanently).ServeHTTP)
	s.router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swagger.SwaggerJSON)
	})
	s.router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(fmt.Sprintf("%s/swagger.json", prefix))))

	s.router.Group(func(rr chi.Router) {
		rr.Use(s.tokenMiddleware)
		rr.Post("/users", s.createUserHandler)
		rr.Put("/users", s.updateUserHandler)
		rr.Get("/users", s.getAllUsersHandler)
		rr.Delete("/users/{username}", s.deleteUserHandler)
	})

	return nil
}

func (s *ApiServer) GetRouter() chi.Router {
	return s.router
}
