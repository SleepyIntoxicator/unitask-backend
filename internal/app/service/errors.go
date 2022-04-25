package service

import (
	"errors"
)

var (
	//	Auth/app/authorization
	ErrAppNameIsAlreadyOccupied = errors.New("this app name is already occupied")
	ErrAppAuthorization         = errors.New("invalid authorization data")
	ErrInvalidAppToken          = errors.New("invalid app token")
	ErrInvalidAppID             = errors.New("invalid app id")

	//	Auth/registry
	ErrInvalidUserEmail       = errors.New("invalid user email")
	ErrInvalidUserLogin       = errors.New("invalid user login")
	ErrMailLoginAlreadyUsing  = errors.New("this login or email address is already in use")
	ErrEmailIsAlreadyOccupied = errors.New("this email is already occupied")
	ErrLoginIsAlreadyOccupied = errors.New("this login is already occupied")
	//ErrPasswordIsTooLight	  = errors.New("this password is too light")

	//	Auth/user/login
	ErrIncorrectLoginOrPassword = errors.New("incorrect login or password")
	//ErrIncorrectEmailOrPassword = errors.New("incorrect email or password")

	//	Auth/user/authorization
	ErrInvalidUserToken               = errors.New("invalid user token")
	ErrInvalidTokenPair               = errors.New("invalid access-refresh token pair")
	ErrInvalidRefreshToken            = errors.New("invalid refresh token")
	ErrAccessTokenExpired             = errors.New("the access token has expired")
	ErrAccessTokenRefreshRateExceeded = errors.New("token refresh rate exceeded")
	ErrAccessTokenIsBlacklisted       = errors.New("the access token is blacklisted")
	//ErrInvalidAccessToken	 = errors.New("invalid access token")

	// Object not found

	ErrAppNotFound                = errors.New("app not found")
	ErrUserNotFound               = errors.New("user not found")
	ErrUserNotFoundInContext      = errors.New("user not found in context")
	ErrAppIDNotFoundInContext     = errors.New("app id not found in context")
	ErrRequestIDNotFoundInContext = errors.New("request id not found in context")
	ErrTaskNotFound               = errors.New("task not found")
	ErrSubjectNotFound            = errors.New("subject not found")
	ErrUniversityNotFound         = errors.New("university not found")
	ErrGroupHaveNoMembers         = errors.New("group have no members")

	ErrGroupNotFound            = errors.New("group not found")
	ErrGroupHaveNotInvitation   = errors.New("group doesn't have an invitation")
	ErrUserIsNotGroupMember     = errors.New("the user in not a member of the group")
	ErrUserNotMemberOfAnyGroups = errors.New("the user is not a member of any groups")

	ErrInvalidLimitOrPage = errors.New("the limit of page or page number can't be less than zero")

	ErrGroupNameIsAlreadyOccupied = errors.New("this group name is already occupied")
)

var (
//ErrorInvalidAccessToken = models.New(ErrInvalidAccess)
)
