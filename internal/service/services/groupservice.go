package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
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

func (s *GroupService) Create(ctx context.Context, group *models.Group, user *models.User) error {
	if err := group.Validate(); err != nil {
		return err
	}

	_, err := s.service.store.Group().FindByName(ctx, group.CustomName)
	if err != nil && err != store.ErrRecordNotFound {
		return err
	}

	if err = s.service.store.Group().Create(ctx, group); err != nil {
		return err
	}

	return nil
}

func (s *GroupService) GetAllGroups(ctx context.Context, limit, offset int) ([]models.Group, error) {
	groups, err := s.service.store.Group().GetAllGroups(ctx, limit, offset)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else {
		return groups, nil
	}
}

// Find returns *models.Group by groupID or nil if group not found.
//	If group was not found, method returns service.ErrGroupNotFound.
//	If an error occurs during the execution of the method, the method returns error.
func (s *GroupService) Find(ctx context.Context, groupID int) (*models.Group, error) {
	group, err := s.service.store.Group().Find(ctx, groupID)
	if err != nil && err != store.ErrNoRowsFound {
		return nil, err
	} else if err == store.ErrNoRowsFound {
		return nil, service.ErrGroupNotFound
	}
	return group, nil
}

func (s *GroupService) FindByName(ctx context.Context, name string) (*models.Group, error) {
	group, err := s.service.store.Group().FindByName(ctx, name)
	if err != nil && err != store.ErrNoRowsFound {
		return nil, err
	} else if err == store.ErrNoRowsFound {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (s *GroupService) Update(ctx context.Context, groupID int, updGroup *models.UpdateGroup) error {
	err := s.service.store.Group().Update(ctx, groupID, updGroup)
	return err
}

func (s *GroupService) Delete(ctx context.Context, groupID, userID int) error {
	_, err := s.service.store.Group().Find(ctx, groupID)
	if err == store.ErrRecordNotFound {
		return service.ErrGroupNotFound
	}
	if err != nil {
		return err
	}

	isUserGroupMember, err := s.IsUserGroupMember(ctx, userID, groupID)
	if err != nil {
		return err
	} else if !isUserGroupMember {
		return service.ErrUserIsNotGroupMember
	}

	err = s.service.store.Group().Delete(ctx, groupID)
	return err
}

func (s *GroupService) GetGroupMembers(ctx context.Context, groupID int) ([]models.User, error) {
	var users []models.User

	users, err := s.service.store.Group().GetGroupMembers(ctx, groupID)
	if err == store.ErrRecordNotFound {
		return users, service.ErrGroupHaveNoMembers
	}
	return users, err
}

func (s *GroupService) GetGroupsUserMemberOf(ctx context.Context, userID int) ([]models.Group, error) {
	_, err := s.service.User().Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	groups, err := s.service.store.Group().GetGroupsUserMemberOf(ctx, userID)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else if err != nil {
		return groups, service.ErrUserNotMemberOfAnyGroups
	}
	return groups, nil
}

func (s *GroupService) GetUserPermissions(ctx context.Context, userID, groupID int) error {
	roles, err := s.service.store.Group().GetMemberRoles(ctx, userID, groupID)
	if err != nil {
		return nil
	}

	return errors.New("to do me GetUserPermissions" + strconv.Itoa(len(roles)))
}

func (s *GroupService) IsUserGroupMember(ctx context.Context, userID, groupID int) (bool, error) {
	return s.service.store.Group().IsUserGroupMember(ctx, userID, groupID)
}

func (s *GroupService) GetGroupMembersCount(ctx context.Context, groupID int) (int, error) {
	return s.service.store.Group().GetMembersCount(ctx, groupID)
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

func (s *GroupService) AddUserToGroupByInvite(ctx context.Context, userID int, invite string) error {
	//Verifying that the user is exist
	isUserExist, err := s.service.store.User().IsUserExist(ctx, userID)
	if err != nil {
		return err
	}
	if !isUserExist {
		return store.ErrUserNotFound
	}

	groupInvite, err := s.service.store.Group().GetGroupInviteByHash(ctx, invite)
	if err == store.ErrRecordNotFound {
		return errors.New("invalid invite")
	} else if err != nil {
		return err
	}

	if time.Now().After(groupInvite.ExpiresAt) {
		return errors.New("invite has expired")
	}

	err = s.service.store.Group().AddGroupMember(ctx, userID, groupInvite.GroupID, groupInvite.InviterID)
	if err != nil {
		return err
	}

	return nil
}

func (s *GroupService) GetInviteLink(ctx context.Context, groupID int) (*models.GroupInvite, error) {
	invite, err := s.service.store.Group().GetGroupInvite(ctx, groupID)
	if err == nil {
		return invite, nil
	} else if err == store.ErrRecordNotFound {
		return nil, service.ErrGroupHaveNotInvitation
	}
	return nil, err
}

func (s *GroupService) GetOrCreateInviteLink(ctx context.Context, groupID int, inviterID int) (*models.GroupInvite, error) {
	group, err := s.service.store.Group().Find(ctx, groupID)
	if err != nil {
		return nil, err
	}

	invite, err := s.service.store.Group().GetGroupInvite(ctx, groupID)
	if err == nil {
		if time.Now().After(invite.ExpiresAt) {
			err = s.service.store.Group().DeleteGroupInviteByHash(ctx, invite.InviteHash)
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

	err = s.service.store.Group().AddGroupInviteHash(ctx, invite)
	if err != nil {
		return nil, err
	}
	//domain-name.com/invite/ + inviteHash
	return invite, nil
}
