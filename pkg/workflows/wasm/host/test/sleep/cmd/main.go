//go:build wasip1

package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(1 * time.Hour)
	fmt.Printf("hello")
}
