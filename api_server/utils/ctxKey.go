package utils

import "context"

type key string

const (
	ctxKey key = "UserID"
)

func NewContext(ctx context.Context, v int) context.Context {
	return context.WithValue(ctx, ctxKey, v)
}

func ContextValue(ctx context.Context) (int64, bool) {
	v, ok := ctx.Value(ctxKey).(int)
	return int64(v), ok
}
