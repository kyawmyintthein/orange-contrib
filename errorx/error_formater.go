package errorx

type ErrorFormatter interface {
	GetArgs() []interface{}
	GetMessage() string
	FormattedMessage() string
}
