//go:build wasip1

package main

import "math"

func main() {
	// allocate more bytes than the binary should be able to access, 64 megs
	_ = make([]byte, int64(128*math.Pow(10, 6)))
}
