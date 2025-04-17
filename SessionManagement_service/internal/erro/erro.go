package erro

import "errors"

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
	ErrorMissingRequestIDConst         = "Error missing request ID"
	ErrorRequiredUserIdConst           = "UserId is required"
	ErrorRequiredSessionIdConst        = "SessionId is required"
)

var (
	ErrorGetSession               = errors.New(ErrorGetSessionConst)
	ErrorSetSession               = errors.New(ErrorSetSessionConst)
	ErrorGetUserIdSession         = errors.New(ErrorGetUserIdSessionConst)
	ErrorGetExpirationTimeSession = errors.New(ErrorGetExpirationTimeSessionConst)
	ErrorSessionParse             = errors.New(ErrorSessionParseConst)
	ErrorAuthorized               = errors.New(ErrorAuthorizedConst)
	ErrorInvalidSessionID         = errors.New(ErrorInvalidSessionConst)
	ErrorInternalServer           = errors.New(ErrorInternalServerConst)
	ErrorContextTimeOut           = errors.New(ErrorContextTimeOutConst)
	ErrorMissingRequestID         = errors.New(ErrorMissingRequestIDConst)
	ErrorRequiredUserId           = errors.New(ErrorRequiredUserIdConst)
	ErrorRequiredSessionId        = errors.New(ErrorRequiredSessionIdConst)
)
