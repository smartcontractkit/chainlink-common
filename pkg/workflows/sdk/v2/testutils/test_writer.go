package testutils

import "io"

type testWriter struct {
	logs [][]byte
}

func (tw *testWriter) Write(p []byte) (n int, err error) {
	tmp := make([]byte, len(p))
	copy(tmp, p)
	tw.logs = append(tw.logs, tmp)
	return len(p), nil
}

var _ io.Writer = (*testWriter)(nil)
