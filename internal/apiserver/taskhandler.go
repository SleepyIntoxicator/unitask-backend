package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func (s *server) handleTasksV2() http.HandlerFunc {
	type response struct {
		Total int    `json:"total"`
		Tasks []Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		tasks, err := s.services.Task().GetAllTasks(r.Context(), limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := response{
			Total: len(tasks),
		}

		for _, t := range tasks {
			newTask, err := s.GetDetailOfTask(r.Context(), t)
			if err != nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			res.Tasks = append(res.Tasks, newTask)
		}

		s.respond(w, r, http.StatusOK, res)
	}
}

//	Remove from prod
func (s *server) handleTasks() http.HandlerFunc {
	type response struct {
		Total int           `json:"total"`
		Tasks []models.Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		tasks, err := s.services.Task().GetAllTasks(r.Context(), limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Total: len(tasks),
			Tasks: tasks})
	}
}

func (s *server) handleGetTask() http.HandlerFunc {
	type request struct {
		Task models.Task `json:"task"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		taskID, err := strconv.Atoi(URLVars["id"])

		tasks, err := s.services.Task().Find(r.Context(), taskID)
		if err == service.ErrTaskNotFound {
			s.error(w, r, http.StatusNotFound, err)
			return
		} else if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, request{Task: *tasks})
	}
}

func (s *server) handleGetAllUserTasks() http.HandlerFunc {
	type request struct {
		Total int           `json:"total"`
		Tasks []models.Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		tasks, err := s.services.Task().GetAllUserTasks(r.Context(), user.ID, limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, request{
			Total: len(tasks),
			Tasks: tasks})
	}
}

func (s *server) handleGetUserLocalTasks() http.HandlerFunc {
	type response struct {
		Total int           `json:"total"`
		Tasks []models.Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		tasks, err := s.services.Task().GetUserLocalTasks(r.Context(), user.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Total: len(tasks),
			Tasks: tasks,
		})
	}
}

func (s *server) handleGetGroupTasks() http.HandlerFunc {
	type response struct {
		Total int           `json:"total"`
		Tasks []models.Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, errors.New("invalid group id type"))
			return
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		tasks, err := s.services.Task().GetTasksOfGroup(r.Context(), groupID, user.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Total: len(tasks),
			Tasks: tasks,
		})
	}
}

func (s *server) handleGetGroupTask() http.HandlerFunc {
	type response struct {
		Task Task `json:"task"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, errors.New("invalid group id type"))
			return
		}
		taskID, err := strconv.Atoi(URLVars["taskId"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, errors.New("invalid task id type"))
			return
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		isUserMember, err := s.services.Group().IsUserGroupMember(user.ID, groupID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		if !isUserMember {
			s.error(w, r, http.StatusForbidden, models.ErrUserIsNotGroupMember)
			return
		}

		tasks, err := s.services.Task().GetTasksOfGroup(r.Context(), groupID, user.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := response{}
		for i, task := range tasks {
			if task.ID == taskID {
				res.Task, err = s.GetDetailOfTask(r.Context(), task)
				if err != nil {
					s.error(w, r, http.StatusInternalServerError, err)
					return
				}
				break
			}
			if i == len(tasks)-1 {
				s.error(w, r, http.StatusNotFound, service.ErrTaskNotFound)
				return
			}
		}

		s.respond(w, r, http.StatusOK, res)

	}
}

func (s *server) handleGetUserTasks() http.HandlerFunc {
	type response struct {
		Total int           `json:"total"`
		Tasks []models.Task `json:"tasks"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		tasks, err := s.services.Task().GetTasksOfUser(r.Context(), user.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{
			Total: len(tasks),
			Tasks: tasks,
		})
	}
}

func (s *server) handleCreateGroupTask() http.HandlerFunc {
	type request struct {
		TypeID int `json:"type_id"`

		Name    string    `json:"name"`
		Content string    `json:"content"`
		StartAt time.Time `json:"start_at"`
		EndAt   time.Time `json:"end_at"`

		GroupsID     []int `json:"group_id"`
		UsersID      []int `json:"user_id"`
		SubjectID    int   `json:"subject_id"`
		ParentTaskID int   `json:"parent_task_id"`
		PrevTasksIDs []int `json:"prev_tasks_ids"`
		NextTasksIDs []int `json:"next_tasks_ids"`
	}
	type response struct {
		Task models.Task `json:"task"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, _ := strconv.Atoi(URLVars["id"])

		req := &request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		task := &models.Task{
			TypeID: req.TypeID,

			Name:    req.Name,
			Content: req.Content,
			StartAt: req.StartAt,
			EndAt:   req.EndAt,

			GroupsID:     append(req.GroupsID, groupID),
			UsersID:      req.UsersID,
			SubjectID:    req.SubjectID,
			ParentTaskID: req.ParentTaskID,
			PrevTasksIDs: req.PrevTasksIDs,
			NextTasksIDs: req.NextTasksIDs,

			AddedByID: user.ID,
		}

		if err := s.services.Task().CreateGroupTask(r.Context(), task); err != nil {
			if err == models.ErrTaskCannotPointToItself {
				s.error(w, r, http.StatusBadRequest, err)
				return
			}
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{Task: *task})
	}
}

func (s *server) handleCreateUserTask() http.HandlerFunc {
	type request struct {
		TypeID int `json:"type_id"`

		Name    string    `json:"name"`
		Content string    `json:"content"`
		StartAt time.Time `json:"start_at"`
		EndAt   time.Time `json:"end_at"`

		SubjectID    int   `json:"subject_id"`
		ParentTaskID int   `json:"parent_task_id"`
		PrevTasksIDs []int `json:"prev_tasks_ids"`
		NextTasksIDs []int `json:"next_tasks_ids"`
	}
	type response struct {
		Task models.Task `json:"task"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		task := &models.Task{
			TypeID: req.TypeID,

			Name:    req.Name,
			Content: req.Content,
			StartAt: req.StartAt,
			EndAt:   req.EndAt,

			UsersID:      []int{user.ID},
			SubjectID:    req.SubjectID,
			ParentTaskID: req.ParentTaskID,
			PrevTasksIDs: req.PrevTasksIDs,
			NextTasksIDs: req.NextTasksIDs,

			AddedByID: user.ID,
		}

		if err := s.services.Task().CreateUserTask(r.Context(), task); err != nil {
			if err == models.ErrTaskCannotPointToItself {
				s.error(w, r, http.StatusBadRequest, err)
				return
			}
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{Task: *task})
	}
}

func (s *server) handleCloneUserTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			Добавить запись в таблицу db.UserTask
			Выставить базовый статус task_in_process
		*/
	}
}

/*		task, err := s.services.Task().GetGroupTaskWithContext(r.Context(), groupID, taskID)
		switch err {
		case nil:
		case models.ErrUserIsNotGroupMember:
			s.error(w, r, http.StatusForbidden, err)
			return
		case service.ErrTaskNotFound:
			s.error(w, r, http.StatusNotFound, err)
			return
		default:
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := response{}
		res.Task, err = s.GetDetailOfTask(*task)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, res)
*/
