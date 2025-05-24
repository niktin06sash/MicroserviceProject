package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const APIServiceUnavalaible = "API-Service is unavailable"
const TooManyRequests = "Too many requests"
const RequiredSession = "Required session in cookie"
const RequestTimedOut = "Request timed out"

const (
	ClientErrorType = "ClientError"
	ServerErrorType = "ServerError"
)

type CustomError struct {
	Message string
	Type    string
}
