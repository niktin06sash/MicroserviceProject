package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const RequestTimedOut = "Request timed out"
const UserIdRequired = "UserID is required"
const UserIdInvalid = "Invalid userID format"
const SessionIdRequired = "SessionID is required"
const SessionIdInvalid = "Invalid sessionID format"
const SessionNotFound = "Session not found"
const AlreadyAuthorized = "Already Authorized"
const ClientErrorType = "Client"
const ServerErrorType = "Server"
const ErrorType = "type"
const ErrorMessage = "message"
const (
	ErrorGetSessionConst               = "Error get session"
	ErrorSetSessionConst               = "Error set session"
	ErrorGetUserIdSessionConst         = "UserID not found in session"
	ErrorGetExpirationTimeSessionConst = "ExpirationTime not found in session"
	ErrorSessionParseConst             = "Error parse session data"
	ErrorAuthorizedConst               = "The current session is active"
	ErrorInvalidSessionConst           = "Error Invalid Session ID"
	ErrorInternalServerConst           = "Internal Server Error"
	ErrorContextTimeOutConst           = "Context time is up"
	ErrorMissingMetadataConst          = "Metadata is required"
	ErrorRequiredTraceIDConst          = "TraceId is required"
	ErrorRequiredUserIdConst           = "UserId is required"
	ErrorRequiredSessionIdConst        = "SessionId is required"
	ErrorInvalidSessionIdFormat        = "Invalid SessionId format"
	ErrorHset                          = "Hset session error: %v"
	ErrorExpire                        = "Expire session error: %v"
	ErrorHgetAll                       = "HGetAll session error: %v"
	ErrorDelSession                    = "Del session error: %v"
	InvalidSession                     = "Session is invalid"
)

type CustomError struct {
	Message string
	Type    string
}
