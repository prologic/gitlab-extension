package contracts

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{err.Error()}
}
