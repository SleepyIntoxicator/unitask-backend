package apiserver

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

//authorizeAdministrator is the middleware.
//For methods that require admin rules
func (s *server) authorizeAdministrator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqUser, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		// Temporarily: No roles added yet
		if reqUser.Login != "admin" {
			s.respondHTML(w, r, http.StatusForbidden, errForbiddenHTML)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) sendStdoutHandler() http.HandlerFunc {
	type logItem struct {
		Level   log.Level  `json:"level"`
		Time    time.Time  `json:"time"`
		Message string     `json:"message"`
		Data    log.Fields `json:"data"`
	}
	type response struct {
		Logs []logItem `json:"logs"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		msg := s.logrusHook.GetAllItems()
		res := response{}
		for _, m := range msg {
			res.Logs = append(res.Logs, logItem{
				Level:   m.Level,
				Time:    m.Time,
				Message: m.Message,
				Data:    m.Data,
			})

		}
		s.respond(w, r, http.StatusOK, res)
	}
}
