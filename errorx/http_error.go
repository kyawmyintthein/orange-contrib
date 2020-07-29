package errorx

type HttpError interface {
	StatusCode() int
}
