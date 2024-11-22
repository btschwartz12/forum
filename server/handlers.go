package server

import (
	"errors"
	"html/template"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/btschwartz12/forum/repo"
	"github.com/btschwartz12/forum/server/assets"
	"github.com/btschwartz12/forum/server/slack"
	"github.com/go-chi/chi/v5"
	"github.com/samber/mo"
)

var (
	homeTmpl = template.Must(template.ParseFS(
		assets.Templates,
		"templates/base.html.tmpl",
		"templates/home.html.tmpl",
	))
)

type templateData struct {
	Time         string
	Posts        []repo.Post
	LoggedInUser repo.User
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repo.User)
	posts, err := s.rpo.GetAllPosts(r.Context())
	if err != nil {
		s.logger.Errorw("Error getting all posts", "error", err)
		http.Error(w, "Internal server error OH MY GOD", http.StatusInternalServerError)
		return
	}

	data := templateData{
		Time:         time.Now().Format(time.RFC3339),
		Posts:        posts,
		LoggedInUser: *user,
	}

	if err := homeTmpl.ExecuteTemplate(w, "base.html.tmpl", data); err != nil {
		s.logger.Errorw("Error executing template", "error", err)
	}
}

func (s *Server) uploadPostHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repo.User)
	ip, err := getIP(r)
	if err != nil {
		ip = r.RemoteAddr
	}

	post := &repo.Post{
		Author:    *user,
		Content:   r.FormValue("content"),
		Timestamp: repo.EstTime{Time: time.Now()},
		Ip:        ip,
	}

	if post.Content == "" {
		http.Error(w, "author and content are required", http.StatusBadRequest)
		return
	}

	fileOpt := mo.Option[multipart.File]{}
	headerOpt := mo.Option[*multipart.FileHeader]{}
	file, header, err := r.FormFile("file")
	if err == nil {
		fileOpt = mo.Some(file)
		headerOpt = mo.Some(header)
		defer file.Close()
	}

	_, err = s.rpo.InsertPost(r.Context(), post, fileOpt, headerOpt)
	if err != nil {
		if errors.Is(err, repo.ErrStorageFull) {
			http.Error(w, "Storage full", http.StatusInternalServerError)
			return
		}
		if errors.Is(err, repo.ErrInvalidExtension) {
			http.Error(w, "Invalid file extension", http.StatusBadRequest)
			return
		}
		if errors.Is(err, repo.ErrFileTooLarge) {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}
		s.logger.Errorw("Error inserting post", "error", err)
		http.Error(w, "Error inserting post", http.StatusInternalServerError)
		return
	}

	go slack.SendPostCreatedAlert(s.slackWebhookUrl, s.logger, post)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repo.User)
	ip, err := getIP(r)
	if err != nil {
		ip = r.RemoteAddr
	}
	postId := chi.URLParam(r, "post_id")
	n, err := strconv.ParseInt(postId, 10, 64)
	if err != nil {
		http.Error(w, "post_id must be an integer", http.StatusBadRequest)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	postAuthor, err := s.rpo.GetPostAuthor(r.Context(), n)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("Error getting post author", "error", err)
		http.Error(w, "Error getting post author", http.StatusInternalServerError)
		return
	}
	if !user.IsAdmin && user.ID != postAuthor.ID {
		http.Error(w, "You can only update your own posts", http.StatusUnauthorized)
		return
	}
	err = s.rpo.UpdatePostContent(r.Context(), n, content, ip)
	if err != nil {
		s.logger.Errorw("Error updating post content", "error", err)
		http.Error(w, "Error updating post content", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*repo.User)
	postId := chi.URLParam(r, "post_id")
	n, err := strconv.ParseInt(postId, 10, 64)
	if err != nil {
		http.Error(w, "post_id must be an integer", http.StatusBadRequest)
		return
	}
	postAuthor, err := s.rpo.GetPostAuthor(r.Context(), n)
	if err != nil {
		if errors.Is(err, repo.ErrPostNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		s.logger.Errorw("Error getting post author", "error", err)
		http.Error(w, "Error getting post author", http.StatusInternalServerError)
		return
	}
	if !user.IsAdmin && user.ID != postAuthor.ID {
		http.Error(w, "You can only delete your own posts", http.StatusUnauthorized)
		return
	}
	err = s.rpo.DeletePost(r.Context(), n)
	if err != nil {
		s.logger.Errorw("Error deleting post", "error", err)
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) serveImageHandler(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}

	fullUrl := s.rpo.GetPathForPost(filename)
	http.ServeFile(w, r, fullUrl)
}
