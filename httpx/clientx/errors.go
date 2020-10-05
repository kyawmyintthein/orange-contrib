package clientx

import "github.com/kyawmyintthein/orange-contrib/errorx"

type ServerError struct {
	*errorx.ErrorX
}

func NewServerError(url string, statusCode int) *ServerError {
	return &ServerError{
		errorx.NewErrorX("server return 5xx status code : %d from URL: %s", statusCode, url),
	}
}
