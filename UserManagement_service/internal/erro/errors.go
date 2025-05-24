package erro

const SessionServiceUnavalaible = "Session-Service is unavailable"
const UserServiceUnavalaible = "User-Service is unavailable"
const RequestTimedOut = "Request timed out"
const ClientErrorType = "ClientError"
const ServerErrorType = "ServerError"

type ErrorResponse struct {
	Message string
	Type    string
}

const (
	ErrorPasswordEmpty             = "Password is missing or empty"
	ErrorGetEnvDBConst             = "DB get environment error"
	ErrorNotPostConst              = "Method is not POST"
	ErrorNotGetConst               = "Method is not GET"
	ErrorNotDeleteConst            = "Method is not DELETE"
	ErrorReadAllConst              = "ReadAll error"
	ErrorUnmarshalConst            = "Unmarshal error"
	ErrorMarshalConst              = "Marshal error"
	ErrorRequiredUserIDConst       = "UserID in request is required!"
	ErrorNotRequiredUserIDConst    = "UserID in request is not-required!"
	ErrorRequiredSessionIDConst    = "SessionID in request is required!"
	ErrorNotRequiredSessionIDConst = "SessionID in request is not-required!"
	ErrorNotEmailConst             = "This email format is not supported"
	ErrorUniqueEmailConst          = "This email has already been registered"
	ErrorHashPassConst             = "Hash-Password error"
	ErrorInternalServerConst       = "Internal Server Error"
	ErrorEmailNotRegisterConst     = "This email is not registered"
	ErrorFoundUserConst            = "Person not found"
	ErrorInvalidPasswordConst      = "Invalid Password"
	ErrorDbRepositoryErrorConst    = "DB-Repository Error"
	ErrorUnexpectedDataConst       = "Unexpected data type"
	ErrorStartTransactionConst     = "Transaction creation error"
	ErrorCommitTransactionConst    = "Transaction commit error"
	ErrorDBOpenConst               = "DB-Open error"
	ErrorDBPingConst               = "DB-Ping error"
	ErrorGrpcResponseConst         = "Grpc's response error"
	ErrorGetUserIdConst            = "Error getting the UserId from the request context"
	ErrorContextTimeoutConst       = "The timeout context has expired"
	ErrorSendKafkaMessageConst     = "Error Kafka Message"
	ErrorRolbackTransactionConst   = "Rolback Transaction error"
	ErrorPanicConst                = "Panic Error"
	ErrorGrpcRollbackConst         = "Rollback's error"
	ErrorMissingUserIDConst        = "Error missing user ID"
	ErrorMissingSessionIDConst     = "Error missing session ID"
	ErrorMissingRequestIDConst     = "Error missing request ID"
	ErrorAllRetryFailedConst       = "All retry attempts failed"
)
