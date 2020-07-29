package cb

import "context"

type CircuitBreaker interface {
	IsEnabled() bool
	Do(context.Context, string, func() error, func(error) error) error
}
