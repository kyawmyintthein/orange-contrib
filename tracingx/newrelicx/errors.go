package newrelicx

import "github.com/kyawmyintthein/orange-contrib/errorx"

type NotAvailable struct {
	*errorx.ErrorX
}

func NewNotAvailabeError() *NotAvailable {
	return errorx.NewErrorX("[%s] new-relic tracer is not avaliable", Package)
}

func (err *NotAvailable) Wrap(cause error) { err.Wrap(cause) }
