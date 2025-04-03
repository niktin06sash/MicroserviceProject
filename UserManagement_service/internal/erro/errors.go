package erro

import "errors"

const (
	ErrorGetEnvDBConst                 = "DB get environment error"
	ErrorNotPostConst                  = "Method is not POST"
	ErrorNotGetConst                   = "Method is not GET"
	ErrorNotDeleteConst                = "Method is not DELETE"
	ErrorReadAllConst                  = "ReadAll error"
	ErrorUnmarshalConst                = "Unmarshal error"
	ErrorMarshalConst                  = "Marshal error"
	ErrorNotEmailConst                 = "This email format is not supported"
	ErrorUniqueEmailConst              = "This email has already been registered"
	ErrorHashPassConst                 = "Hash-Password error"
	ErrorInternalServerConst           = "Internal Server Error"
	ErrorEmailNotRegisterConst         = "This email is not registered"
	ErrorFoundUserConst                = "Person not found"
	ErrorInvalidPasswordConst          = "Invalid Password"
	ErrorInvalidSessionIDConst         = "Session not found"
	ErrorGetSessionConst               = "Error get session"
	ErrorSetSessionConst               = "Error set session"
	ErrorGetUserIdSessionConst         = "UserID not found in session"
	ErrorGetExpirationTimeSessionConst = "ExpirationTime not found in session"
	ErrorSessionParseConst             = "Error parse session data"
	ErrorUnexpectedDataConst           = "Unexpected data type"
	ErrorStartTransactionConst         = "Transaction creation error"
	ErrorCommitTransactionConst        = "Transaction commit error"
	ErrorAuthorizedConst               = "The current session is active"
	ErrorGetUserIdConst                = "Error getting the UserId from the request context"
	ErrorContextTimeoutConst           = "The timeout context has expired"
	ErrorSendKafkaMessageConst         = "Error Kafka Message"
)

var (
	ErrorGetEnvDB                 = errors.New(ErrorGetEnvDBConst)
	ErrorNotPost                  = errors.New(ErrorNotPostConst)
	ErrorNotGet                   = errors.New(ErrorNotGetConst)
	ErrorNotDelete                = errors.New(ErrorNotDeleteConst)
	ErrorReadAll                  = errors.New(ErrorReadAllConst)
	ErrorUnmarshal                = errors.New(ErrorUnmarshalConst)
	ErrorMarshal                  = errors.New(ErrorMarshalConst)
	ErrorNotEmail                 = errors.New(ErrorNotEmailConst)
	ErrorUniqueEmail              = errors.New(ErrorUniqueEmailConst)
	ErrorHashPass                 = errors.New(ErrorHashPassConst)
	ErrorInternalServer           = errors.New(ErrorInternalServerConst)
	ErrorEmailNotRegister         = errors.New(ErrorEmailNotRegisterConst)
	ErrorFoundUser                = errors.New(ErrorFoundUserConst)
	ErrorInvalidPassword          = errors.New(ErrorInvalidPasswordConst)
	ErrorInvalidSessionID         = errors.New(ErrorInvalidSessionIDConst)
	ErrorGetSession               = errors.New(ErrorGetSessionConst)
	ErrorSetSession               = errors.New(ErrorSetSessionConst)
	ErrorGetUserIdSession         = errors.New(ErrorGetUserIdSessionConst)
	ErrorGetExpirationTimeSession = errors.New(ErrorGetExpirationTimeSessionConst)
	ErrorSessionParse             = errors.New(ErrorSessionParseConst)
	ErrorUnexpectedData           = errors.New(ErrorUnexpectedDataConst)
	ErrorStartTransaction         = errors.New(ErrorStartTransactionConst)
	ErrorCommitTransaction        = errors.New(ErrorCommitTransactionConst)
	ErrorAuthorized               = errors.New(ErrorAuthorizedConst)
	ErrorGetUserId                = errors.New(ErrorGetUserIdConst)
	ErrorContextTimeout           = errors.New(ErrorContextTimeoutConst)
	ErrorSendKafkaMessage         = errors.New(ErrorSendKafkaMessageConst)
)
