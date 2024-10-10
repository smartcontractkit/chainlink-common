//go:build !wasip1

package wasm

func NewRunner() *Runner {
	panic("error: NewRunner() is only intended for use with `GOOS=wasip1 GOARCH=wasm`. For testing, use testutils.NewRunner() instead.")
}
