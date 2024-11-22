package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/btschwartz12/forum/repo"
	"github.com/go-chi/chi/v5"
)

// createUserHandler godoc
// @Summary Create a user
// @Description Create a user
// @Tags users
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param is_admin formData bool false "Is Admin"
// @Router /api/users [post]
// @Security Bearer
// @Success 200 {string} string "user id"
func (s *ApiServer) createUserHandler(w http.ResponseWriter, r *http.Request) {
	user := repo.User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		IsAdmin:  r.FormValue("is_admin") == "true",
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	id, err := s.rpo.InsertUser(r.Context(), &user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error inserting user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", id)
}

// updateUserHandler godoc
// @Summary Update a user
// @Description Update a user
// @Tags users
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param id formData int true "ID"
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param is_admin formData bool false "Is Admin"
// @Router /api/users [put]
// @Security Bearer
// @Success 200 {string} string "user id"
func (s *ApiServer) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	n, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}
	user := repo.User{
		ID:       n,
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		IsAdmin:  r.FormValue("is_admin") == "true",
	}

	if user.Username == "" || user.Password == "" {
		http.Error(w, "id, username, and password are required", http.StatusBadRequest)
		return
	}

	err = s.rpo.UpdateUser(r.Context(), &user)
	if err != nil {
		http.Error(w, fmt.Sprintf("error updating user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%d", user.ID)
}

// getAllUsersHandler godoc
// @Summary Get all users
// @Description Get all users
// @Tags users
// @Produce json
// @Router /api/users [get]
// @Security Bearer
// @Success 200 {array} repo.User
func (s *ApiServer) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.rpo.GetAllUsers(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting users: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := json.MarshalIndent(users, "", "\t")
	if err != nil {
		http.Error(w, fmt.Sprintf("error marshalling users: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// deleteUserHandler godoc
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Router /api/users/{username} [delete]
// @Param username path string true "Username"
// @Security Bearer
// @Success 200
func (s *ApiServer) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	err := s.rpo.DeleteUser(r.Context(), username)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
