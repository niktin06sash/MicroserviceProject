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
	ErrorMissingMetadataConst          = "Metadata is required"
	ErrorRequiredRequestIDConst        = "RequestId is required"
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
	ErrorMissingMetadata          = errors.New(ErrorMissingMetadataConst)
	ErrorRequiredUserId           = errors.New(ErrorRequiredUserIdConst)
	ErrorRequiredRequestID        = errors.New(ErrorRequiredRequestIDConst)
	ErrorRequiredSessionId        = errors.New(ErrorRequiredSessionIdConst)
)
