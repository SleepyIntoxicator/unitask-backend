package models

import (
	"errors"
	"strconv"
)

var (
	ErrAccessTokenIsInvalid      = errors.New("invalid access token")
	ErrAccessTokenIsExpired      = errors.New("the access token has expired")
	ErrAccessTokenIsNotValidYet  = errors.New("the access token is not valid yet")
	ErrTaskCannotPointToItself   = errors.New("the task cannot point to itself")
	ErrLimitLessThanZero         = errors.New("the limit of page can't be less than zero")
	ErrOffsetLessThanZero        = errors.New("the page of page can't be less than zero")
	ErrLimitOrOffsetLessThanZero = errors.New("the limit of items or offset can't be less than zero")
	ErrLimitOrOffsetTooLarge     = errors.New("exceeded the maximum value of limit or offset")

	ErrUserIsNotGroupMember = errors.New("the user is not a member of the group")
)

type ServerError struct {
	Text string `json:"error_text"`
	Code int    `json:"code"`
	Name string `json:"name"`
}

func New(err error, errCode int, errName string) ServerError {
	return ServerError{
		Text: err.Error(),
		Code: errCode,
		Name: errName,
	}
}

func (e *ServerError) Error() map[string]string {
	return map[string]string{
		"code":  strconv.Itoa(e.Code),
		"error": e.Text,
		"name":  e.Name,
	}
}

var (
	ErrorAccessTokenIsExpired = New(ErrAccessTokenIsExpired, 400, "access_token_expired")
)
