package contexts

import "context"

// Value gets a value from the context and casts it to T.
func Value[T any](ctx context.Context, key any) T {
	v := ctx.Value(key)
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

type key string

const (
	keyOrg      key = "org"
	keyUser     key = "user"
	keyWorkflow key = "workflow"
)

func WithOrg(ctx context.Context, org string) context.Context {
	return context.WithValue(ctx, keyOrg, org)
}

func OrgValue(ctx context.Context) (org string) {
	return Value[string](ctx, keyOrg)
}

func WithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, keyUser, user)
}

func UserValue(ctx context.Context) (org string) {
	return Value[string](ctx, keyUser)
}

func WithWorkflow(ctx context.Context, workflow string) context.Context {
	return context.WithValue(ctx, keyWorkflow, workflow)
}

func WorkflowValue(ctx context.Context) (org string) {
	return Value[string](ctx, keyWorkflow)
}
