package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const UserServiceUnavalaible = "User-Service is unavailable"
const RequestTimedOut = "Request timed out"
const ClientErrorType = "Client"
const ServerErrorType = "Server"
const ErrorType = "type"
const ErrorMessage = "message"

const (
	ErrorSearchOwnProfile             = "You cannot search for your own profile by ID"
	ErrorInvalidCountDinamicParameter = "Invalid count dinamic parameter in request"
	ErrorInvalidDinamicParameter      = "Invalid dinamic parameter"
	ErrorInvalidQueryParameter        = "Invalid query parameter"
	ErrorInvalidCountQueryParameters  = "More than one parameter in a query-request"
	ErrorInvalidDataReq               = "Invalid data in request's body"
	ErrorInvalidReqMethod             = "Invalid request method"
	ErrorPasswordEmpty                = "Password is missing or empty"
	ErrorGetEnvDBConst                = "DB get environment error"
	ErrorReadAllConst                 = "ReadAll error"
	ErrorUnmarshal                    = "Data unmarshal error: %v"
	ErrorRequiredUserIDConst          = "UserID in request is required!"
	ErrorNotRequiredUserIDConst       = "UserID in request is not-required!"
	ErrorRequiredSessionIDConst       = "SessionID in request is required!"
	ErrorNotRequiredSessionIDConst    = "SessionID in request is not-required!"
	ErrorNotEmailConst                = "This email format is not supported"
	ErrorInternalServerConst          = "Internal Server Error"
	ErrorEmailNotRegisterConst        = "Not registered has been entered"
	ErrorFoundUserConst               = "Person not found"
	ErrorInvalidPasswordConst         = "Invalid Password"
	ErrorDbRepositoryErrorConst       = "DB-Repository Error"
	ErrorUnexpectedDataConst          = "Unexpected data type"
	ErrorStartTransactionConst        = "Transaction creation error"
	ErrorCommitTransactionConst       = "Transaction commit error"
	ErrorDBOpenConst                  = "DB-Open error"
	ErrorDBPingConst                  = "DB-Ping error"
	ErrorGrpcResponseConst            = "Grpc's response error"
	ErrorGetUserIdConst               = "Error getting the UserId from the request context"
	ErrorContextTimeoutConst          = "The timeout context has expired"
	ErrorSendKafkaMessageConst        = "Error Kafka Message"
	ErrorRolbackTransactionConst      = "Rolback Transaction error"
	ErrorPanicConst                   = "Panic Error"
	ErrorGrpcRollbackConst            = "Rollback's error"
	ErrorMissingUserIDConst           = "Error missing userID"
	ErrorMissingSessionIDConst        = "Error missing sessionID"
	ErrorMissingRequestIDConst        = "Error missing requestID"
	ErrorInvalidUserIDFormat          = "Invalid userID format in request"
	ErrorAllRetryFailedConst          = "All retry attempts failed"
	ErrorUniqueEmailConst             = "Already registered email has been entered"
	ErrorIncorrectPassword            = "Incorrect password has been entered"
	ErrorIDNotRegisterConst           = "Unregistered id has been entered"
	ErrorAfterReqUsers                = "Error after request into users: %v"
	ErrorGenerateHashPassword         = "Generate HashPassword : %v"
	ErrorSetProfiles                  = "Set profiles-cache error: %v"
	ErrorHgetAllProfiles              = "HGetAll profiles-cache error: %v"
	ErrorExpireProfiles               = "Expire profiles-cache error: %v"
	ErrorDelProfiles                  = "Del profiles-cache error: %v"
	ErrorMarshal                      = "Data marshal error: %v"
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
