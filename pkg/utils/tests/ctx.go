//go:build !go1.24

package tests

func Context(tb TestingT) (ctx context.Context) {
	return getContext(tb)
}
