package constant

var (
	ErrServer          = "unexpected error occured"
	ErrParseReqBody    = "error parsing request body"
	ErrParseQueryParam = "error parsing query param"
	ErrValidation      = "validation error"
	ErrNotAuthorized   = "you're not authorized"
	ErrNotFound        = "data not found"
	ErrInvalidJson     = "invalid JSON format"
	ErrInvalidToken    = "invalid or expired token"
	ErrInvalidParam    = "invalid param"

	ErrUserNotFound   = "User not found"
	ErrUsernameExists = "Username already exists"
	ErrInvalidCreds   = "Invalid credentials"
)
