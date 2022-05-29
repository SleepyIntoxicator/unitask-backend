package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type GroupRepository struct {
	store *Store
}

func (r *GroupRepository) Create(ctx context.Context, group *models.Group) error {
	query := `INSERT INTO public.group (custom_name, university_id, specialization_name, start_year, course_number, group_number, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`

	err := r.store.db.QueryRowContext(ctx, query,
		group.CustomName,
		group.UniversityID,
		group.SpecializationName,
		group.StartYear,
		group.CourseNumber,
		group.GroupNumber,
		time.Now()).Scan(&group.ID, &group.CreatedAt)
	group.CompileFullGroupNameAndCompareCustom()
	return err
}

func (r *GroupRepository) GetAllGroups(ctx context.Context, limit, offset int) ([]models.Group, error) {
	var groups []models.Group

	query := `SELECT * FROM public.group ORDER BY id`
	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	err = r.store.db.SelectContext(ctx, &groups, query)
	if err != nil {
		return nil, store.HandleErrorNoRows(err)
	}

	for i := range groups {
		groups[i].CompileFullGroupNameAndCompareCustom()
	}

	return groups, nil
}

func (r *GroupRepository) Find(ctx context.Context, id int) (*models.Group, error) {
	g := &models.Group{}
	query := `SELECT id, custom_name, university_id, specialization_name, start_year, course_number, group_number, created_at FROM public.group where id = $1`
	err := r.store.db.QueryRowContext(ctx, query, id).Scan(
		&g.ID,
		&g.CustomName,
		&g.UniversityID,
		&g.SpecializationName,
		&g.StartYear,
		&g.CourseNumber,
		&g.GroupNumber,
		&g.CreatedAt)
	g.CompileFullGroupNameAndCompareCustom()
	return g, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) FindByName(ctx context.Context, name string) (*models.Group, error) {
	g := &models.Group{}
	query := `SELECT id, custom_name, university_id, specialization_name, start_year, course_number, group_number, created_at  FROM public.group WHERE custom_name = $1`
	err := r.store.db.QueryRowContext(ctx, query, name).Scan(
		&g.ID,
		&g.CustomName,
		&g.UniversityID,
		&g.SpecializationName,
		&g.StartYear,
		&g.CourseNumber,
		&g.GroupNumber,
		&g.CreatedAt)
	g.CompileFullGroupNameAndCompareCustom()
	return g, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) Update(ctx context.Context, groupID int, up *models.UpdateGroup) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if up.Name != nil {
		setValues = append(setValues, fmt.Sprintf("specialization_name=$%d", argID))
		args = append(args, *up.Name)
		argID++
	}
	if up.UniversityID != nil {
		setValues = append(setValues, fmt.Sprintf("university_id=$%d", argID))
		args = append(args, *up.UniversityID)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE public.group SET %s WHERE id = $%d",
		setQuery, argID)
	args = append(args, groupID)

	fmt.Printf("updateQuery: %s\n", query)
	fmt.Printf("args: %v\n", args)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.store.db.ExecContext(ctx, query, args...)
	return err
}

func (r *GroupRepository) Delete(ctx context.Context, id int) (err error) {
	tx := r.store.db.MustBegin()

	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			netErr := recoverErr.(*pgconn.PgError)
			err = netErr

			logrus.WithFields(logrus.Fields{
				"error": netErr.Error(),
			}).Info("PostgreSQL recover from panic")

			rollbackErr := tx.Rollback()

			if rollbackErr != nil {
				logrus.WithFields(logrus.Fields{
					"error": rollbackErr,
				}).Info("PostgreSQL rollback with error")
			}
		}
	}()

	tx.MustExecContext(ctx, `DELETE FROM taskongroup WHERE group_id = $1`, id)

	tx.MustExecContext(ctx, `DELETE FROM groupmemberroles WHERE group_member_id IN
                                   (SELECT id FROM groupmember WHERE group_id = $1)`, id)

	tx.MustExecContext(ctx, `DELETE FROM groupmember WHERE group_id = $1`, id)

	tx.MustExecContext(ctx, `DELETE FROM groupinvitehashes WHERE group_id = $1`, id)

	tx.MustExecContext(ctx, `DELETE FROM "group" WHERE id = $1`, id)

	err = tx.Commit()

	return store.HandleIgnoreErrorNoRows(err)
}

func (r *GroupRepository) IsGroupExist(ctx context.Context, groupID int) (bool, error) {
	query := `SELECT FROM public.group WHERE id = $1`
	err := r.store.db.QueryRowContext(ctx, query, groupID).Err()

	return store.HandleIsFieldFounded(err)
}

// AddGroupMember returns an error if it occurred.
// Returns the store.ErrUserNotFound if the user doesn't exist
// Returns the store.ErrGroupNotFound if the group doesn't exist
func (r *GroupRepository) AddGroupMember(ctx context.Context, userID, groupID int, inviterID int) error {

	isMember, err := r.IsUserGroupMember(ctx, userID, groupID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("the user is already a member of the group")
	}

	var query string
	if inviterID != 0 {
		query = `INSERT INTO groupmember (user_id, group_id, invited_by_id) VALUES ($1, $2, $3)`
		_, err = r.store.db.ExecContext(ctx, query, userID, groupID, inviterID)
	} else {
		query = `INSERT INTO groupmember (user_id, group_id) VALUES ($1, $2)`
		_, err = r.store.db.ExecContext(ctx, query, userID, groupID)
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *GroupRepository) IsUserGroupMember(ctx context.Context, userID, groupID int) (bool, error) {
	query := `SELECT FROM groupmember WHERE user_id = $1 AND group_id = $2`
	err := r.store.db.QueryRowContext(ctx, query, userID, groupID).Scan()

	return store.HandleIsFieldFounded(err)
}

func (r *GroupRepository) GetGroupsUserMemberOf(ctx context.Context, userID int) ([]models.Group, error) {
	var groups []models.Group

	query := `SELECT id, custom_name, university_id, specialization_name, start_year, course_number, group_number, created_at FROM "group" WHERE id IN 
                          (SELECT group_id FROM groupmember WHERE user_id = $1) ORDER BY id`
	err := r.store.db.SelectContext(ctx, &groups, query, userID)
	for i := range groups {
		groups[i].CompileFullGroupNameAndCompareCustom()
	}
	return groups, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetGroupMembers(ctx context.Context, groupID int) ([]models.User, error) {
	var users []models.User
	query := `SELECT * FROM public.user WHERE id IN 
                                (SELECT user_id FROM groupmember WHERE group_id = $1) ORDER BY id`
	err := r.store.db.SelectContext(ctx, &users, query, groupID)

	return users, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetMembersCount(ctx context.Context, groupID int) (int, error) {
	var countMembers int

	query := `SELECT count(*) FROM groupmember WHERE group_id = $1`
	err := r.store.db.QueryRowContext(ctx, query, groupID).Scan(&countMembers)
	if err != nil {
		return 0, store.HandleErrorNoRows(err)
	}
	return countMembers, nil
}

func (r *GroupRepository) GetMemberRoles(ctx context.Context, userID, groupID int) ([]models.Role, error) {
	var roles []models.Role

	query := `SELECT id, name, description FROM public.role 
				WHERE id in (SELECT role_id FROM groupmemberroles 
				      WHERE group_member_id = (SELECT id FROM groupmember 
				          WHERE user_id = $1 AND group_id = $2))`
	rows, err := r.store.db.QueryContext(ctx, query, userID, groupID)
	if err != nil {
		return roles, store.HandleErrorNoRows(err)
	}

	for rows.Next() {
		role := models.Role{}
		//Parsing row with role
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}

		roles = append(roles, role)
	}

	return roles, nil
}

func (r *GroupRepository) GetRolePermissions(ctx context.Context, roleID int) ([]models.Permission, error) {
	var permissions []models.Permission
	query := `SELECT id, name FROM permission 
				WHERE id in ( SELECT permission_id FROM rolepermissions WHERE role_id = $1) ORDER BY id`
	err := r.store.db.SelectContext(ctx, permissions, query, roleID)

	return permissions, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetRole(ctx context.Context, roleID int) (*models.Role, error) {
	role := &models.Role{}

	query := `SELECT id, name, description FROM public.role WHERE id = $1`
	err := r.store.db.SelectContext(ctx, role, query, roleID)

	return role, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetRoleByName(ctx context.Context, roleName string) (*models.Role, error) {
	role := &models.Role{}

	query := `SELECT id, name, description FROM public.role WHERE name = $1`
	err := r.store.db.SelectContext(ctx, role, query, roleName)

	return role, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetGroupInvite(ctx context.Context, groupID int) (*models.GroupInvite, error) {
	invite := &models.GroupInvite{}

	query := `SELECT group_id, inviter_id, hash, expires_at FROM groupinvitehashes
				WHERE group_id = $1`
	err := r.store.db.QueryRowContext(ctx, query, groupID).Scan(
		&invite.GroupID,
		&invite.InviterID,
		&invite.InviteHash,
		&invite.ExpiresAt)
	return invite, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) GetGroupInviteByHash(ctx context.Context, inviteHash string) (*models.GroupInvite, error) {
	invite := &models.GroupInvite{}

	query := `SELECT group_id, inviter_id, hash, expires_at FROM groupinvitehashes
				WHERE hash = $1`
	err := r.store.db.QueryRowContext(ctx, query, inviteHash).Scan(
		&invite.GroupID,
		&invite.InviterID,
		&invite.InviteHash,
		&invite.ExpiresAt)
	return invite, store.HandleErrorNoRows(err)
}

func (r *GroupRepository) AddGroupInviteHash(ctx context.Context, invite *models.GroupInvite) error {
	query := `INSERT INTO groupinvitehashes (group_id, inviter_id, hash, expires_at) VALUES
				($1, $2, $3, $4)`

	err := r.store.db.QueryRowContext(ctx, query,
		invite.GroupID,
		invite.InviterID,
		invite.InviteHash,
		invite.ExpiresAt).Err()

	return err
}

func (r *GroupRepository) DeleteGroupInvites(ctx context.Context, groupID int) error {
	query := `DELETE FROM groupinvitehashes WHERE group_id = $1`
	_, err := r.store.db.ExecContext(ctx, query, groupID)
	return err
}

func (r *GroupRepository) DeleteGroupInviteByHash(ctx context.Context, hash string) error {
	query := `DELETE FROM groupinvitehashes WHERE hash = $1`
	_, err := r.store.db.ExecContext(ctx, query, hash)
	return err
}
