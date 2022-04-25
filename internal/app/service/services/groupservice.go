package services

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/service"
	"back-end/internal/app/store"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

var (
	//TODO: create the getting this from the config (via .env)
	inviteTTL = 24 * time.Hour * 15 // 15 days
)

type GroupService struct {
	service *Service
}

func (s *GroupService) Create(group *models.Group, user *models.User) error {
	if err := group.Validate(); err != nil {
		return err
	}

	_, err := s.service.store.Group().FindByName(group.CustomName)
	if err != nil && err != store.ErrRecordNotFound {
		return err
	}

	if err = s.service.store.Group().Create(group); err != nil {
		return err
	}

	return nil
}

func (s *GroupService) GetAllGroups(limit, offset int) ([]models.Group, error) {
	groups, err := s.service.store.Group().GetAllGroups(limit, offset)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else {
		return groups, nil
	}
}

func (s *GroupService) GetGroupMembers(groupID int) ([]models.User, error) {
	var users []models.User

	users, err := s.service.store.Group().GetGroupMembers(groupID)
	if err == store.ErrRecordNotFound {
		return users, service.ErrGroupHaveNoMembers
	}
	return users, err
}

func (s *GroupService) GetUserPermissions(userID, groupID int) error {
	roles, err := s.service.store.Group().GetMemberRoles(userID, groupID)
	if err != nil {
		return nil
	}

	return errors.New("to do me GetUserPermissions" + strconv.Itoa(len(roles)))
}

// Find returns *models.Group by groupID or nil if group not found.
//	If group was not found, method returns service.ErrGroupNotFound.
//	If an error occurs during the execution of the method, the method returns error.
func (s *GroupService) Find(groupID int) (*models.Group, error) {
	group, err := s.service.store.Group().Find(groupID)
	if err != nil && err != store.ErrNoRowsFound {
		return nil, err
	} else if err == store.ErrNoRowsFound {
		return nil, service.ErrGroupNotFound
	}
	return group, nil
}

func (s *GroupService) FindByName(name string) (*models.Group, error) {
	group, err := s.service.store.Group().FindByName(name)
	if err != nil && err != store.ErrNoRowsFound {
		return nil, err
	} else if err == store.ErrNoRowsFound {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (s *GroupService) Update(groupID int, updGroup *models.UpdateGroup) error {
	err := s.service.store.Group().Update(groupID, updGroup)
	return err
}

func (s *GroupService) Delete(groupID, userID int) error {
	_, err := s.service.store.Group().Find(groupID)
	if err == store.ErrRecordNotFound {
		return service.ErrGroupNotFound
	}
	if err != nil {
		return err
	}

	isUserGroupMember, err := s.IsUserGroupMember(userID, groupID)
	if err != nil {
		return err
	} else if !isUserGroupMember {
		return service.ErrUserIsNotGroupMember
	}

	err = s.service.store.Group().Delete(groupID)
	return err
}

func (s *GroupService) IsUserGroupMember(userID, groupID int) (bool, error) {
	return s.service.store.Group().IsUserGroupMember(userID, groupID)
}

func (s *GroupService) GetGroupMembersCount(groupID int) (int, error) {
	return s.service.store.Group().GetMembersCount(groupID)
}

func (s *GroupService) GetGroupsUserMemberOf(userID int) ([]models.Group, error) {
	_, err := s.service.User().Find(userID)
	if err != nil {
		return nil, err
	}

	groups, err := s.service.store.Group().GetGroupsUserMemberOf(userID)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else if err != nil {
		return groups, service.ErrUserNotMemberOfAnyGroups
	}
	return groups, nil
}

func (s *GroupService) GetMemberRoles(userID, groupID int) ([]int, error) {
	panic("implement me")
}

func (s *GroupService) GetRolePermissions(roleID int) ([]models.Permission, error) {
	panic("implement me")
}

func (s *GroupService) GetRole(roleID int) (*models.Role, error) {
	panic("implement me")
}

func (s *GroupService) GetRoleByName(roleName string) (*models.Role, error) {
	panic("implement me")
}

func (s *GroupService) AddUserToGroupByInvite(userID int, invite string) error {
	//Verifying that the user is exist
	isUserExist, err := s.service.store.User().IsUserExist(userID)
	if err != nil {
		return err
	}
	if !isUserExist {
		return store.ErrUserNotFound
	}

	groupInvite, err := s.service.store.Group().GetGroupInviteByHash(invite)
	if err == store.ErrRecordNotFound {
		return errors.New("invalid invite")
	} else if err != nil {
		return err
	}

	if time.Now().After(groupInvite.ExpiresAt) {
		return errors.New("invite has expired")
	}

	err = s.service.store.Group().AddGroupMember(userID, groupInvite.GroupID, groupInvite.InviterID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupService) GetOrCreateInviteLink(groupID int, inviterID int) (*models.GroupInvite, error) {
	group, err := s.service.store.Group().Find(groupID)
	if err != nil {
		return nil, err
	}

	invite, err := s.service.store.Group().GetGroupInvite(groupID)
	if err == nil {
		if time.Now().After(invite.ExpiresAt) {
			err = s.service.store.Group().DeleteGroupInviteByHash(invite.InviteHash)
			if err != nil {
				return nil, err
			}
		}
		return invite, err
	} else if err != store.ErrRecordNotFound {
		return nil, err
	}

	URLEncoder := md5.New()
	URLEncoder.Write([]byte(strconv.Itoa(group.ID)))
	URLEncoder.Write([]byte(strconv.Itoa(group.UniversityID)))
	URLEncoder.Write([]byte(group.CustomName))

	inviteHash := hex.EncodeToString(URLEncoder.Sum(nil))

	invite = &models.GroupInvite{
		GroupID:    groupID,
		InviterID:  inviterID,
		InviteHash: inviteHash,
		ExpiresAt:  time.Now().Add(inviteTTL),
	}

	err = s.service.store.Group().AddGroupInviteHash(invite)
	if err != nil {
		return nil, err
	}
	//domain-name.com/invite/ + inviteHash
	return invite, nil
}

func (s *GroupService) GetInviteLink(groupID int) (*models.GroupInvite, error) {
	invite, err := s.service.store.Group().GetGroupInvite(groupID)
	if err == nil {
		return invite, nil
	} else if err == store.ErrRecordNotFound {
		return nil, service.ErrGroupHaveNotInvitation
	}
	return nil, err
}
