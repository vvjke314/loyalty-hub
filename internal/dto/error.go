package dto

type ErrorResponse struct {
	Message string `json:"error"`
}

func NewErrorResponse(msg string) ErrorResponse {
	return ErrorResponse{
		Message: msg,
	}
}
