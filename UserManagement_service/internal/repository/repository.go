package repository

//go:generate mockgen -source=repository.go -destination=mocks/mock.go
const (
	KeyUserID = "id"
	KeyUser   = "user"
)

func BadResponse(erro error, place string) *RepositoryResponse {
	return &RepositoryResponse{Errors: erro, Success: false, Place: place}
}
func SuccessResponse(data map[string]any, place string, succmessage string) *RepositoryResponse {
	return &RepositoryResponse{Data: data, Place: place, Success: true, SuccessMessage: succmessage}
}

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         error
}
