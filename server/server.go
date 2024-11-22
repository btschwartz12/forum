package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"

	"github.com/btschwartz12/forum/repo"
	"github.com/btschwartz12/forum/server/api"
	"github.com/btschwartz12/forum/server/assets"
)

type Server struct {
	router          *chi.Mux
	rpo             *repo.Repo
	logger          *zap.SugaredLogger
	store           *sessions.CookieStore
	slackWebhookUrl string
	public          bool
}

func (s *Server) Init(
	logger *zap.SugaredLogger,
	varDir string,
	public bool,
	authToken string,
	sessionKey string,
	slackWebhookUrl string,
) error {

	var err error
	s.rpo, err = repo.NewRepo(logger, varDir)
	if err != nil {
		return fmt.Errorf("error creating repo: %w", err)
	}

	s.logger = logger
	s.slackWebhookUrl = slackWebhookUrl
	s.public = public
	s.store = sessions.NewCookieStore([]byte(sessionKey))
	s.store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   31536000,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	}
	s.router = chi.NewRouter()

	s.router.Get("/favicon.ico", http.RedirectHandler("/static/img/favicon.ico", http.StatusMovedPermanently).ServeHTTP)
	s.router.Handle("/static/*", staticHandler(http.FileServer(http.FS(assets.Static)), "/"))
	s.router.Get("/image/{filename}", s.serveImageHandler)

	s.router.Get("/login", s.loginPageHandler)
	s.router.Post("/login", s.loginHandler)

	if public {
		s.router.Post("/signup", s.signupHandler)
	}

	s.router.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)
		r.Get("/", s.home)
		r.Post("/upload", s.uploadPostHandler)
		r.Put("/update/{post_id}", s.updatePostHandler)
		r.Delete("/delete/{post_id}", s.deletePostHandler)
	})

	apiServer := &api.ApiServer{}
	err = apiServer.Init(logger, s.rpo, "/api", authToken)
	if err != nil {
		return fmt.Errorf("error initializing api server: %w", err)
	}
	s.router.Mount("/api", apiServer.GetRouter())

	return nil
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func staticHandler(fs http.Handler, strip string) http.Handler {
	return http.StripPrefix(strip, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		fs.ServeHTTP(w, r)
	}))
}

func getIP(r *http.Request) (string, error) {
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}
