package models

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

type Subject struct {
	ID int	`json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (s *Subject) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(2, 64)))

}