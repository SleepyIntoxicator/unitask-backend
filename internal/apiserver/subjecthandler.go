package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func (s *server) handleSubjectCreate() http.HandlerFunc {
	type request struct {
		SubjectName string `json:"subject_name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}

		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		subject := &models.Subject{
			Name: req.SubjectName,
		}
		err = s.services.Subject().Create(r.Context(), subject)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, subject)
	}
}

func (s *server) handleSubjects() http.HandlerFunc {
	type response struct {
		Total    int              `json:"total"`
		Subjects []models.Subject `json:"subjects"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		subjects, err := s.services.Subject().GetAllSubjects(r.Context(), limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Total:    len(subjects),
			Subjects: subjects,
		})
	}
}

func (s *server) handleSubject() http.HandlerFunc {
	type response struct {
		Subject *models.Subject `json:"subject"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		subjectID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		subject, err := s.services.Subject().Find(r.Context(), subjectID)
		if err != nil && err != service.ErrSubjectNotFound {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		} else if err == service.ErrSubjectNotFound {
			s.error(w, r, http.StatusNotFound, err)
			return
		}

		res := response{
			Subject: subject,
		}
		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleDeleteSubject() http.HandlerFunc {
	type response struct {
		Subject *models.Subject `json:"subject"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		subjectID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		subject, err := s.services.Subject().Delete(r.Context(), subjectID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Subject: subject,
		})
	}
}
