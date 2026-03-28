package constant

var (
	ErrServer          = "unexpected error occured"
	ErrParseReqBody    = "error parsing request body"
	ErrParseQueryParam = "error parsing query param"
	ErrValidation      = "validation error"
	ErrNotAuthorized   = "you're not authorized"
	ErrNotFound        = "data not found"
	ErrInvalidJson     = "invalid JSON format"
	ErrInvalidToken    = "Invalid or expired token"

	ErrUserNotFound = "User not found"
	ErrEmailExists  = "Email already exists"
	ErrInvalidCreds = "Invalid credentials"
)
