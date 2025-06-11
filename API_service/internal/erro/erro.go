package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const APIServiceUnavalaible = "API-Service is unavailable"
const TooManyRequests = "Too many requests"
const RequiredSession = "Required session in cookie"
const RequestTimedOut = "Request timed out"
const PageNotFound = "Page not found"
const (
	ErrorType       = "type"
	ErrorMessage    = "message"
	ClientErrorType = "ClientError"
	ServerErrorType = "ServerError"
)
