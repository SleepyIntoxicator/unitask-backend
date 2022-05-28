package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
)

func (s *server) handleGroups() http.HandlerFunc {
	type response struct {
		Total  int             `json:"total"`
		Groups []FullGroupInfo `json:"groups"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset, err := s.getLimitAndOffsetFromQuery(r)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		groups, err := s.services.Group().GetAllGroups(r.Context(), limit, offset)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].ID < groups[j].ID
		})

		res := &response{}

		for _, gr := range groups {
			newResponseGroup, err := s.GetFullInfoOfGroup(gr)
			if err != nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}

			res.Groups = append(res.Groups, newResponseGroup)
			res.Total = len(res.Groups)
		}
		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleGroup() http.HandlerFunc {
	type response struct {
		GroupInfo FullGroupInfo `json:"group"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, errors.New("invalid group id type"))
			return
		}

		group, err := s.services.Group().Find(r.Context(), groupID)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		fullGroupInfo, err := s.GetFullInfoOfGroup(*group)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, &response{
			GroupInfo: fullGroupInfo,
		})
	}
}

func (s *server) handleGroupCreate() http.HandlerFunc {
	type request struct {
		CustomName         string `json:"custom_name"`
		UniversityID       int    `json:"university_id"`
		SpecializationName string `json:"specialization_name"`
		StartYear          string `json:"start_year"`
		CourseNumber       int    `json:"course_number"`
		GroupNumber        string `json:"group_number"`
	}
	type response struct {
		Group models.Group `json:"group"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			if err == errJSONEOF {
				s.error(w, r, http.StatusBadRequest, errJSONParseEOF)
				return
			}
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		creator, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		group := &models.Group{
			CustomName:         req.CustomName,
			UniversityID:       req.UniversityID,
			SpecializationName: req.SpecializationName,
			StartYear:          req.StartYear,
			CourseNumber:       req.CourseNumber,
			GroupNumber:        req.GroupNumber,
		}

		err = s.services.Group().Create(r.Context(), group, creator)
		if err != nil {
			s.error(w, r, http.StatusOK, err)
			return
		}

		s.respond(w, r, http.StatusOK, response{*group})
	}
}

//	Requires: The user must be a member of the group
func (s *server) handleGroupDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, _ := strconv.Atoi(URLVars["id"])

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		err = s.services.Group().Delete(r.Context(), groupID, user.ID)
		if err == service.ErrUserIsNotGroupMember {
			s.error(w, r, http.StatusForbidden, err)
			return
		} else if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleGroupCreateInvitation() http.HandlerFunc {
	type Inviter struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		//TODO: Avatar
	}
	type Group struct {
		ID             int    `json:"id"`
		Name           string `json:"name"`
		UniversityName string `json:"university_name"`
	}
	type response struct {
		InviteHash string  `json:"invite"`
		Group      Group   `json:"group"`
		Inviter    Inviter `json:"inviter"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.errorV2(w, r, 500, models.New(err, 500, "invalid_group_id"))
		}

		inviterID, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		invite, err := s.services.Group().GetOrCreateInviteLink(r.Context(), groupID, inviterID.ID)
		if err != nil {
			s.errorV2(w, r, 500, models.New(err, 500, "invalid_group_id?"))
			return
		}

		group, err := s.services.Group().Find(r.Context(), groupID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		university, err := s.services.University().Find(group.UniversityID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &response{
			InviteHash: invite.InviteHash,
			Inviter: Inviter{
				ID:       inviterID.ID,
				Username: inviterID.Login,
			},
			Group: Group{
				ID:             groupID,
				Name:           group.CustomName,
				UniversityName: university.Name,
			},
		}

		s.respond(w, r, 200, res)
	}
}

func (s *server) handleJoinToGroupWithInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		invite := URLVars["hash"]

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		err = s.services.Group().AddUserToGroupByInvite(r.Context(), user.ID, invite)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusNoContent, nil)
	}
}

func (s *server) handleGroupWhereUserIsMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			NumberOfGroups int            `json:"number_of_groups"`
			Groups         []models.Group `json:"groups"`
		}

		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := response{}
		groups, err := s.services.Group().GetGroupsUserMemberOf(r.Context(), user.ID)
		if err == service.ErrUserNotMemberOfAnyGroups {

		} else if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res.NumberOfGroups = len(groups)
		res.Groups = groups

		s.respond(w, r, http.StatusOK, res)
	}
}

func (s *server) handleGetGroupMembers() http.HandlerFunc {
	type response struct {
		Total int                `json:"total"`
		Users []UserResponseInfo `json:"members"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		URLVars := mux.Vars(r)
		groupID, err := strconv.Atoi(URLVars["id"])
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		members, err := s.services.Group().GetGroupMembers(r.Context(), groupID)
		if err != nil && err != service.ErrGroupHaveNoMembers {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := response{
			Total: len(members),
		}
		for _, m := range members {
			res.Users = append(res.Users, s.GetUserResponseInfo(m))
		}

		s.respond(w, r, http.StatusOK, res)
	}
}
