package teststore

import (
	"backend/internal/api/v1/models"
)

type GroupRepository struct {
	store *Store
}

func (r *GroupRepository) Create(group *models.Group) error {
	panic("implement me")
}

func (r *GroupRepository) GetAllGroups(limit, offset int) ([]models.Group, error) {
	panic("implement me")
}

func (r *GroupRepository) Find(i int) (*models.Group, error) {
	panic("implement me")
}

func (r *GroupRepository) FindByName(s string) (*models.Group, error) {
	panic("implement me")
}

func (r *GroupRepository) Update(i int, group *models.UpdateGroup) error {
	panic("implement me")
}

func (r *GroupRepository) Delete(i int) error {
	panic("implement me")
}

func (r *GroupRepository) IsGroupExist(groupID int) (bool, error) {
	panic("implement me")
}

func (r *GroupRepository) AddGroupMember(userID, groupID int, inviterID int) error {
	panic("implement me")
}

func (r *GroupRepository) IsUserGroupMember(userID, groupID int) (bool, error) {
	panic("implement me")
}

func (r *GroupRepository) GetGroupsUserMemberOf(userID int) ([]models.Group, error) {
	panic("implement me")
}

func (r *GroupRepository) GetGroupMembers(groupID int) ([]models.User, error) {
	panic("implement me")
}

func (r *GroupRepository) GetMembersCount(groupID int) (int, error) {
	panic("implement me")
}

func (r *GroupRepository) GetMemberRoles(userID, groupID int) ([]models.Role, error) {
	panic("implement me")
}

func (r *GroupRepository) GetRolePermissions(roleID int) ([]models.Permission, error) {
	panic("implement me")
}

func (r *GroupRepository) GetRole(roleID int) (*models.Role, error) {
	panic("implement me")
}

func (r *GroupRepository) GetRoleByName(roleName string) (*models.Role, error) {
	panic("implement me")
}

func (r *GroupRepository) GetGroupInvite(groupID int) (*models.GroupInvite, error) {
	panic("implement me")
}

func (r *GroupRepository) GetGroupInviteByHash(inviteHash string) (*models.GroupInvite, error) {
	panic("implement me")
}

func (r *GroupRepository) AddGroupInviteHash(invite *models.GroupInvite) error {
	panic("implement me")
}

func (r *GroupRepository) DeleteGroupInvites(groupID int) error {
	panic("implement me")
}

func (r *GroupRepository) DeleteGroupInviteByHash(hash string) error {
	panic("implement me")
}
