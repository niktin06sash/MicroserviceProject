package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const APIServiceUnavalaible = "API-Service is unavailable"
const PhotoServiceUnavalaible = "Photo-Service is unavailable"
const UserServiceUnavalaible = "User-Service is unavailable"
const TooManyRequests = "Too many requests"
const RequiredSession = "Required session in cookie"
const RequestTimedOut = "Request timed out"
const PageNotFound = "Page not found"
const RequiredFormPhoto = "Photo file is required"
const (
	ErrorType       = "type"
	ErrorMessage    = "message"
	ClientErrorType = "Client"
	ServerErrorType = "Server"
)

type CustomError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func ServerError(reason string) *CustomError {
	return &CustomError{Message: reason, Type: ServerErrorType}
}
func ClientError(reason string) *CustomError {
	return &CustomError{Message: reason, Type: ClientErrorType}
}
