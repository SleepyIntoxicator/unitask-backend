package apiserver

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/service"
	"context"
	"time"
)

type UserResponseInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	//TODO: Avatar
}

type GroupResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Group struct {
	ID         int               `json:"id"`
	FullName   string            `json:"full_name"`
	University models.University `json:"university"`

	CourseNumber int       `json:"course_number"`
	StartYear    string    `json:"start_year"`
	CreatedAt    time.Time `json:"created_at"`
}

type FullGroupInfo struct {
	Group             Group               `json:"group"`
	GroupMembersCount int                 `json:"group_members_count"`
	Members           []UserResponseInfo  `json:"members"`
	GroupHaveInvite   bool                `json:"group_have_invite"`
	Invite            *models.GroupInvite `json:"invite"`
}

type ShortTaskInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID          int  `json:"id"`
	TypeID      int  `json:"type_id,omitempty" `
	IsGroupTask bool `json:"is_group_task"`
	IsLocalTask bool `json:"is_local_task"`

	Name    string    `json:"name"`
	Content string    `json:"content"`
	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`

	Groups     []GroupResponse    `json:"groups"`
	Users      []UserResponseInfo `json:"users"`
	Subject    models.Subject     `json:"subject"`
	ParentTask *ShortTaskInfo     `json:"parent_task"`
	Subtasks   []ShortTaskInfo    `json:"subtasks"`
	PrevTasks  []ShortTaskInfo    `json:"prev_tasks"`
	NextTasks  []ShortTaskInfo    `json:"next_tasks"`
	AddedByID  UserResponseInfo   `json:"added_by"`

	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatesCount int       `json:"updates_count"`
	Views        int       `json:"watches"`
}

func (s *server) GetFullInfoOfGroup(group models.Group) (FullGroupInfo, error) {
	newGroupInfo := FullGroupInfo{}
	newGroupInfo.Group = Group{
		ID:           group.ID,
		FullName:     group.CustomName,
		CourseNumber: group.CourseNumber,
		StartYear:    group.StartYear,
		CreatedAt:    group.CreatedAt,
	}

	university, err := s.services.University().Find(group.UniversityID)
	if err != nil {
		return newGroupInfo, err
	}
	newGroupInfo.Group.University = *university

	//Adding members to response groups
	members, err := s.services.Group().GetGroupMembers(group.ID)
	if err != nil && err != service.ErrGroupHaveNoMembers {
		return newGroupInfo, err
	}

	if len(members) > 0 {
		for _, member := range members {
			u := UserResponseInfo{
				ID:       member.ID,
				Username: member.Login,
				FullName: member.FullName,
			}
			newGroupInfo.Members = append(newGroupInfo.Members, u)
		}
	}
	newGroupInfo.GroupMembersCount = len(newGroupInfo.Members)

	invite, err := s.services.Group().GetInviteLink(group.ID)
	if err != nil && err != service.ErrGroupHaveNotInvitation {
		return newGroupInfo, err
	}
	if invite == nil {
		newGroupInfo.GroupHaveInvite = false
	} else {
		newGroupInfo.GroupHaveInvite = true
		newGroupInfo.Invite = invite
	}

	return newGroupInfo, nil
}

func (s *server) GetUserResponseInfo(user models.User) UserResponseInfo {
	return UserResponseInfo{
		ID:       user.ID,
		Username: user.Login,
		FullName: user.FullName,
	}
}

func (s *server) GetDetailOfTask(ctx context.Context, t models.Task) (Task, error) {
	newTask := Task{
		ID:          t.ID,
		TypeID:      t.TypeID,
		IsGroupTask: t.IsGroupTask,
		IsLocalTask: t.IsLocalTask,

		Name:    t.Name,
		Content: t.Content,
		StartAt: t.StartAt,
		EndAt:   t.EndAt,

		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.LastUpdatedAt,
		UpdatesCount: t.UpdatesCount,
		Views:        t.Views,
	}

	for _, gID := range t.GroupsID {
		gr, err := s.services.Group().Find(gID)
		if err != nil {
			return Task{}, err
		}
		newGroup := GroupResponse{
			ID:   gr.ID,
			Name: gr.CustomName,
		}
		newTask.Groups = append(newTask.Groups, newGroup)
	}
	for _, uID := range t.UsersID {
		u, err := s.services.User().Find(uID)
		if err != nil {
			return Task{}, err

		}
		newUser := UserResponseInfo{
			ID:       u.ID,
			Username: u.Login,
			FullName: u.FullName,
		}
		newTask.Users = append(newTask.Users, newUser)
	}
	sub, err := s.services.Subject().Find(t.SubjectID)
	if err != nil && err != service.ErrSubjectNotFound {
		return Task{}, err

	} else if sub != nil {
		newTask.Subject = *sub
	}

	parent, err := s.services.Task().Find(ctx, t.ParentTaskID)
	if err != nil && err != service.ErrTaskNotFound {
		return Task{}, err

	} else if parent != nil {
		newTask.ParentTask = &ShortTaskInfo{
			ID:   parent.ID,
			Name: parent.Name,
		}
	}

	for _, sID := range t.SubtasksIDs {
		subt, err := s.services.Task().Find(ctx, sID)
		if err != nil {
			return Task{}, err

		}
		newSubtask := ShortTaskInfo{
			ID:   subt.ID,
			Name: subt.Name,
		}
		newTask.Subtasks = append(newTask.Subtasks, newSubtask)
	}
	for _, pID := range t.PrevTasksIDs {
		pr, err := s.services.Task().Find(ctx, pID)
		if err != nil {
			return Task{}, err

		}
		newPrevTask := ShortTaskInfo{
			ID:   pr.ID,
			Name: pr.Name,
		}
		newTask.PrevTasks = append(newTask.PrevTasks, newPrevTask)
	}
	for _, nID := range t.NextTasksIDs {
		nx, err := s.services.Task().Find(ctx, nID)
		if err != nil {
			return Task{}, err

		}
		newNextTask := ShortTaskInfo{
			ID:   nx.ID,
			Name: nx.Name,
		}
		newTask.NextTasks = append(newTask.NextTasks, newNextTask)
	}

	u, err := s.services.User().Find(t.AddedByID)
	if err != nil {
		return Task{}, err

	}

	newTask.AddedByID = UserResponseInfo{
		ID:       u.ID,
		Username: u.Login,
		FullName: u.FullName,
	}

	return newTask, nil
}
