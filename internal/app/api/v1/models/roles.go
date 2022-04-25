package models

type Permission struct {
	ID   int
	Name string
}

type RolePermissions struct {
	ID         int
	Permission Permission
	State      bool //TODO: Optimize struct
}

type Role struct {
	ID          int               `db:"id"`
	Name        string            `db:"name"`
	Description string            `db:"description"`
	Permissions []RolePermissions `db:"-"`
}
