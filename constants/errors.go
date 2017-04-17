package constants

const (
	// ErrorAuth is to be sent when a user is not logged in but must be
	// to perform some action
	ErrorAuth = "not_logged_in"

	// ErrorCSRF is to be sent when a user's csrf token is invalid for
	// a request
	ErrorCSRF = "csrf_invalid"

	// ErrorArguments is to be sent when a user's arguments are invalid
	ErrorArguments = "arguments_invalid"

	// Error403 is to be sent when a user is authenticated but Forbidden
	// from performing an action
	Error403 = "action_forbidden"

	// Error404 is to be sent when a requested resource was not found
	Error404 = "not_found"

	// Error500 is a catch-all that can be sent on error
	Error500 = "internal_error"
)
