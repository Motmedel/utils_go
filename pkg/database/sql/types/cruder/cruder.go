package cruder

import "context"

type Cruder[C any, R any, U any, P any, D any] interface {
	Create(context.Context, C) (string, error)
	Read(context.Context, string) (R, error)
	ReadBulk(context.Context) ([]R, error)
	Update(context.Context, U) (string, error)
	Replace(context.Context, P) (string, error)
	Delete(context.Context, string) (string, error)
}
