package server

import (
	"context"
	"html/template"
	"net/http"

	"github.com/btschwartz12/forum/repo"
	"github.com/btschwartz12/forum/server/assets"
	"github.com/btschwartz12/forum/server/slack"
)

var (
	loginTmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/login.html.tmpl",
		"templates/base.html.tmpl",
	))

	cookieName = "auth"
)

func (s *Server) loginPageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Public bool
	}{
		Public: s.public,
	}

	if err := loginTmpl.ExecuteTemplate(w, "base.html.tmpl", data); err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		s.logger.Errorw("Error executing template", "error", err)
	}
}
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	pass := r.FormValue("password")

	if username == "" || pass == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user, err := s.rpo.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == repo.ErrUserNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("Error getting user", "error", err)
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	if user.Password != pass {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		s.logger.Infow("Invalid password", "username", username, "submitted", pass, "expected", user.Password)
		return
	}

	session, err := s.store.Get(r, cookieName)
	if err != nil {
		s.logger.Errorw("Error getting session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["username"] = user.Username
	if err = session.Save(r, w); err != nil {
		s.logger.Errorw("Error saving session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.store.Get(r, cookieName)
		if err != nil {
			s.logger.Errorw("Error getting session", "error", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		username, ok := session.Values["username"]
		if !ok || username == nil {
			s.logger.Infow("No username in session")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := s.rpo.GetUserByUsername(r.Context(), username.(string))
		if err != nil {
			s.logger.Errorw("Error getting user", "error", err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		s.logger.Infow("user authenticated", "username", user.Username)
		ctx := context.WithValue(r.Context(), "user", user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) signupHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	pass := r.FormValue("password")

	if username == "" || pass == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	user := &repo.User{
		Username: username,
		Password: pass,
		IsAdmin:  false,
	}

	_, err := s.rpo.InsertUser(r.Context(), user)
	if err != nil {
		if err == repo.ErrUserAlreadyExists {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		s.logger.Errorw("Error inserting user", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session, err := s.store.Get(r, cookieName)
	if err != nil {
		s.logger.Errorw("Error getting session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["username"] = user.Username
	if err = session.Save(r, w); err != nil {
		s.logger.Errorw("Error saving session", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	go slack.SendUserCreatedAlert(s.slackWebhookUrl, s.logger, user)

	http.Redirect(w, r, "/", http.StatusFound)
}
