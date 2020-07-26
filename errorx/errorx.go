package errorx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"strings"
	"sync"
)

const (
	NoErrorMessage string = "NoErrorMessage"
)

type ErrorX struct {
	id            string
	code          int
	messageFormat string
	cause         error
	args          []interface{}

	stack       []uintptr
	framesOnce  sync.Once
	stackFrames []StackFrame
}

func New(code int, id string, messageFormat string, args ...interface{}) *ErrorX {
	stack := make([]uintptr, 2)
	stackLength := runtime.Callers(3, stack)
	err := &ErrorX{
		id:            id,
		code:          code,
		cause:         nil,
		messageFormat: messageFormat,
		args:          args,
		stack:         stack[:stackLength],
	}
	return err
}

func (e *ErrorX) ID() string {
	return e.id
}

func (e *ErrorX) Code() int {
	return e.code
}

func (e *ErrorX) Message() string {
	return e.messageFormat
}

func (e *ErrorX) GetArgs() []interface{} {
	return e.args
}

// Return nested error
func (e *ErrorX) GetMessage() string {
	return e.messageFormat
}

func (e *ErrorX) Wrap(err error) error {
	e.cause = err
	return e
}

func (e *ErrorX) Error() string {
	return e.FormattedMessage()
}

func (e *ErrorX) FormattedMessage() string {
	if e.messageFormat != "" {
		argsMap := make(map[string]string)
		msg := e.messageFormat
		if len(e.args) != 0 {
			previousKey := ""
			for _, v := range e.args {
				if previousKey != "" {
					argsMap[previousKey] = v.(string)
				}
				previousKey = v.(string)
			}
		}
		for k, v := range argsMap {
			msg = strings.Replace(msg, fmt.Sprintf("{{var_%s}}", k), v, -1)
		}
		return msg
	} else if e.id != "" {
		argsMap := make(map[string]string)
		if len(e.args) != 0 {
			previousKey := ""
			for _, v := range e.args {
				if previousKey != "" {
					argsMap[previousKey] = v.(string)
				}
				previousKey = v.(string)
			}
		}
		var buf bytes.Buffer
		for k, v := range argsMap {
			buf.WriteString(fmt.Sprintf("%s:%v, ", k, v))
		}

		if len(argsMap) != 0 {
			buf.Truncate(buf.Len() - 2)
			return fmt.Sprintf("%s : [%s]", e.id, buf.String())
		}
		return fmt.Sprintf("%s", e.id)
	}
	return NoErrorMessage
}

func (w *ErrorX) Cause() error { return w.cause }

func (e *ErrorX) StackAddrs() string {
	buf := bytes.NewBuffer(make([]byte, 0, len(e.stack)*8))
	for _, pc := range e.stack {
		fmt.Fprintf(buf, "0x%x ", pc)
	}
	bufBytes := buf.Bytes()
	return string(bufBytes[:len(bufBytes)-1])
}

func (e *ErrorX) StackFrames() []StackFrame {
	e.framesOnce.Do(func() {
		e.stackFrames = make([]StackFrame, len(e.stack))
		for i, pc := range e.stack {
			frame := &e.stackFrames[i]
			frame.PC = pc
			frame.Func = runtime.FuncForPC(pc)
			if frame.Func != nil {
				frame.FuncName = frame.Func.Name()
				frame.File, frame.LineNumber = frame.Func.FileLine(frame.PC - 1)
			}
		}
	})
	return e.stackFrames
}

func (e *ErrorX) GetStack() string {
	stackFrames := e.StackFrames()
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	for _, frame := range stackFrames {
		_, _ = buf.WriteString(frame.FuncName)
		_, _ = buf.WriteString("\n")
		fmt.Fprintf(buf, "\t%s:%d +0x%x\n",
			frame.File, frame.LineNumber, frame.PC)
	}
	return buf.String()
}

func (e *ErrorX) GetStackAsJSON() interface{} {
	stackFrames := e.StackFrames()
	buf := bytes.NewBuffer(make([]byte, 0, 256))
	var (
		data []byte
		i    interface{}
	)
	data = append(data, '[')
	for i, frame := range stackFrames {
		if i != 0 {
			data = append(data, ',')
		}
		name := path.Base(frame.FuncName)
		frameBytes := []byte(fmt.Sprintf(`{"filepath": "%s", "name": "%s", "line": %d}`, frame.File, name, frame.LineNumber))
		data = append(data, frameBytes...)
	}
	data = append(data, ']')
	buf.Write(data)
	_ = json.Unmarshal(data, &i)
	return i
}

func GetErrorMessages(e error) string {
	return extractFullErrorMessage(e, false)
}

func GetErrorMessagesWithStack(e error) string {
	return extractFullErrorMessage(e, true)
}

func extractFullErrorMessage(e error, includeStack bool) string {
	type causer interface {
		Cause() error
	}

	var ok bool
	var lastClErr error
	errMsg := bytes.NewBuffer(make([]byte, 0, 1024))
	dbxErr := e
	for {
		_, ok := dbxErr.(StackTracer)
		if ok {
			lastClErr = dbxErr
		}

		errorWithFormat, ok := dbxErr.(ErrorFormatter)
		if ok {
			errMsg.WriteString(errorWithFormat.FormattedMessage())
		}

		errorCauser, ok := dbxErr.(causer)
		if ok {
			innerErr := errorCauser.Cause()
			if innerErr == nil {
				break
			}
			dbxErr = innerErr
		} else {
			// We have reached the end and traveresed all inner errors.
			// Add last message and exit loop.
			errMsg.WriteString(dbxErr.Error())
			break
		}
		errMsg.WriteString(", ")
	}

	stackError, ok := lastClErr.(StackTracer)
	if includeStack && ok {
		errMsg.WriteString("\nSTACK TRACE:\n")
		errMsg.WriteString(stackError.GetStack())
	}
	return errMsg.String()
}

func Cause(err error) error {
	type causer interface {
		Cause() error
	}

	for err != nil {
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}
