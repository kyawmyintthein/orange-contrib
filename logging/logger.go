package logging

import "context"

type KV map[string]interface{}

type Logger interface {
	InfoKV(context.Context, KV, string)
}

func DefaultLogger() Logger {
	return nil
}
