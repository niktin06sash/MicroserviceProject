package repository

type RepositoryResponse struct {
	Success        bool
	SuccessMessage string
	Place          string
	Data           map[string]any
	Errors         *ErrorResponse
}
type ErrorResponse struct {
	Message string
	Type    string
}
