package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (s *server) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusUnauthorized, errNotAuthenticated)
			return
		}

		s.respond(w, r, http.StatusOK, user)
	}
}

func (s *server) handleUsers() http.HandlerFunc {
	type response struct {
		Total int           `json:"total"`
		Users []models.User `json:"users"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		users, err := s.services.User().GetAllUsers(r.Context(), limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, response{
			Total: len(users),
			Users: users,
		})
	}
}

func (s *server) handleUser() http.HandlerFunc {
	type response struct {
		User models.User `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		userID, err := strconv.Atoi(URLVars["id"])

		user, err := s.services.User().Find(r.Context(), userID)
		if err != nil && err == service.ErrUserNotFound {
			s.error(w, r, http.StatusBadRequest, err)
			return
		} else if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, response{User: *user})
	}
}
