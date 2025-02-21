//go:build !wasip1

package v2

func NewRunnerV2() *RunnerV2 {
	panic("error: NewRunnerV2() is only intended for use with `GOOS=wasip1 GOARCH=wasm`")
}
