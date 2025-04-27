package erro

type ErrorType string

const (
	ClientErrorType ErrorType = "ClientError"
	ServerErrorType ErrorType = "ServerError"
)

type CustomError struct {
	ErrorName string
	ErrorType ErrorType
}
type ErrorInterface interface {
	Error() string
	GetTypeError() ErrorType
}

func (ce CustomError) Error() string {
	return ce.ErrorName
}
func (ce CustomError) GetTypeError() ErrorType {
	return ce.ErrorType
}
