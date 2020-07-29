package main

import (
	"fmt"

	"github.com/kyawmyintthein/orange-contrib/errorx"
)

type MyError struct {
	*errorx.ErrorX
	*errorx.ErrorWithCode
	*errorx.ErrorWithID
	*errorx.ErrorWithHttpStatus
	*errorx.ErrorStacktrace
}

func NewMyError() *MyError {
	statusCode := errorx.GenerateHttpStatusCodeFromErrorCode(40000)
	return &MyError{
		errorx.NewErrorX("my error"),
		errorx.NewErrorWithCode(40000),
		errorx.NewErrorWithID("file_not_found"),
		errorx.NewErrorWithHttpStatus(statusCode),
		errorx.NewErrorWithStackTrace(2, 2),
	}
}

func main() {
	err := getMyError()
	fmt.Println("Error")
	fmt.Println(err)
	fmt.Println("---------------------------------------------------------------------------------------------")
	myerror, ok := err.(*MyError)
	if ok {
		fmt.Println("Error type casting")
		fmt.Printf("err type is *MyError, Error : %s \n", myerror)
		fmt.Println("---------------------------------------------------------------------------------------------")
	}

	rootCause, ok := err.(errorx.Causer)
	if ok {
		fmt.Println("Cause")
		fmt.Println(rootCause.Cause())
		fmt.Println("---------------------------------------------------------------------------------------------")
	}

	errWithCode, ok := err.(errorx.ErrorCode)
	if ok {
		fmt.Println("Error Code")
		fmt.Println(errWithCode.Code())
		fmt.Println("---------------------------------------------------------------------------------------------")
	}

	errWithID, ok := err.(errorx.ErrorID)
	if ok {
		fmt.Println("ID")
		fmt.Println(errWithID.ID())
		fmt.Println("---------------------------------------------------------------------------------------------")
	}

	errWithHttpStatusCode, ok := err.(errorx.HttpError)
	if ok {
		fmt.Println("Http Status")
		fmt.Println(errWithHttpStatusCode.StatusCode())
		fmt.Println("---------------------------------------------------------------------------------------------")
	}

	errStacktrace, ok := err.(errorx.StackTracer)
	if ok {
		fmt.Println("Stacktrace")
		fmt.Println(errStacktrace.GetStack())
		fmt.Println("---------------------------------------------------------------------------------------------")
	}
}

func getMyError() error {
	err := NewMyError()
	err.Wrap(fmt.Errorf("root cause"))
	return err
}
