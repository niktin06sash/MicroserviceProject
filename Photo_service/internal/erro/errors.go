package erro

const PhotoServiceUnavalaible = "Photo-Service is unavailable"
const RequestTimedOut = "Request timed out"
const ClientErrorType = "Client"
const ServerErrorType = "Server"
const ErrorType = "type"
const ErrorMessage = "message"
const DeleteSomeonePhoto = "Attempt to delete someone else's photo"
const InvalidUserIDFormat = "Invalid userID format in request"
const UnregisteredUserID = "Unregistered userid has been entered"
const NonExistentData = "A non-existent data has been entered"
const ContextCanceled = "Context canceled or timeout"
const LargeFile = "File too large - max 10 MB"
const InvalidFileFormat = "Invalid file format"
const ErrorAfterReqPhotos = "Error after request into photos: %v"
const ErrorAfterReqUsersID = "Error after request into usersid: %v"

type CustomError struct {
	Message string
	Type    string
}
