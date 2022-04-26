package apiserver

import (
	"back-end/internal/app/api/v1/models"
	"errors"
)

var (
	errIncorrectEmailOrPassword = errors.New("incorrect email of password")
	errNotAuthenticated         = errors.New("not authenticated")
	errInvalidRequest        = errors.New("invalid request")
	errUserNotFoundInContext = errors.New("user not found in context")

	errAppTokenNotFound = errors.New("app token wasn't found in header")

	errJSONEOF      = errors.New("EOF")
	errJSONParseEOF = errors.New("unexpected EOF when parsing JSON")
	errMissingRequestBody = errors.New("missing request body")

	errUnhandledError = errors.New("unhandled error")
	errInvalidAuthorizationKey = errors.New("invalid authorization key")
)

const (
	errUnauthorizedHTML = "<h1>401 Unauthorized</h1>"
	errForbiddenHTML    = "<h1>403 Forbidden</h1>"
)

var (
	ErrorInvalidAuthorizationKey = models.New(errInvalidAuthorizationKey, 401, "invalid_authorization_key")
	)
