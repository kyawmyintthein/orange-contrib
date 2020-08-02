package clientx

import "github.com/kyawmyintthein/orange-contrib/errorx"

type ServerError struct {
	*errorx.ErrorX
}

func NewServerError(statusCode int) *ServerError {
	return errorx.NewErrorX("server return 5xx status code : %d", statusCode)
}
