package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID                int       `json:"id"`
	Login             string    `json:"login" binding:"required" db:"login"`
	FullName          string    `json:"full_name" binding:"required" db:"full_name"`
	Email             string    `json:"email" binding:"required" db:"email"`
	Password          string    `json:"password,omitempty"`
	EncryptedPassword string    `json:"-" db:"encrypted_password"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required, validation.RuneLength(2, 32)),
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.By(requiredIf(u.EncryptedPassword == "")), validation.RuneLength(6, 100)),
	)
}

func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := EncryptPassword(u.Password)
		if err != nil {
			return err
		}

		u.EncryptedPassword = enc
	}

	return nil
}

func (u *User) Sanitize() {
	u.Password = ""
}

func (u *User) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password)) == nil
}

func EncryptPassword(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

/*// ?
type UserRegistry struct {
	PasswordHash [256]string
}
*/
type UserSignUp struct {
}

type UserSignIn struct {
	Login             string `json:"login"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"encrypted_password"`
	UserAgent         string `json:"user_agent"`
	IP                string `json:"ip"`
}

func (s *UserSignIn) Validate() error {
	return validation.ValidateStruct(
		s,
		validation.Field(&s.Login, validation.By(requiredIf(s.Email == "")), validation.RuneLength(2, 32)),
		validation.Field(&s.Email, validation.By(requiredIf(s.Login == "")), is.Email),
		validation.Field(&s.Password, validation.Required, validation.RuneLength(6, 100)),
		validation.Field(&s.IP, validation.Required, is.IP),
	)
}

/*func (u *UserSignIn) BeforeVerify() error {
	if u.EncryptedPassword == "" {
		encPwd, err := EncryptPassword(u.Password)
		if err != nil {
			return err
		}
		u.EncryptedPassword = string(encPwd)
	}
	return nil
}*/

type UserRole struct {
	UserID int
	RoleID int
}
