package models

import (
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"time"
)

type Group struct {
	ID           int    `json:"id" db:"id"`
	UniversityID int    `json:"university_id" db:"university_id"`
	//If the CustomName matches the generated one (FullName), then
	//the response does not specify
	CustomName   string `json:"custom_name,omitempty" db:"custom_name"`
	FullName     string `json:"full_name" db:"-"`

	SpecializationName string    `json:"name" db:"specialization_name"`
	StartYear          string    `json:"start_year" db:"start_year"`
	CourseNumber       int       `json:"course_number" db:"course_number"`
	GroupNumber        string    `json:"-" db:"group_number"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

func (g *Group) Validate() error {
	return validation.ValidateStruct(
		g,
		validation.Field(&g.SpecializationName, validation.Required, validation.RuneLength(2, 16)),
	)

}

func (g *Group) CompileFullGroupNameAndCompareCustom() {
	g.FullName = fmt.Sprintf("%s-%d%s (%s)", g.SpecializationName, g.CourseNumber, g.GroupNumber, g.StartYear)
	if g.FullName == g.CustomName {
		g.CustomName = ""
	}
}

type UpdateGroup struct {
	UniversityID *int
	Name         *string
}

func (up *UpdateGroup) Validate() error {
	if up.UniversityID == nil && up.Name == nil {
		return errors.New("update structure has no values")
	}
	return nil
}

type University struct {
	ID       int       `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	Location string    `json:"location" db:"location"`
	Site     string    `json:"site" db:"site"`
	AddedAt  time.Time `json:"added_at" db:"added_at"`
}

type GroupMember struct {
	ID      int
	UserID  int
	GroupID int
}

type GroupInvite struct {
	GroupID    int       `json:"group_id" db:"group_id"`
	InviterID  int       `json:"inviter_id" db:"inviter_id"`
	InviteHash string    `json:"invite_hash" db:"hash"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
}

func (gi *GroupInvite) Validate() error {
	return errors.New("not implemented")
}
