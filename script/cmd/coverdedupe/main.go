// coverdedupe merges duplicate coverage blocks in a Go cover profile.
//
// go test ./... -coverpkg=./... emits one block per tested package for every
// instrumented file. Blocks that were not executed in a package are recorded
// with count 0. SonarQube treats those duplicates as uncovered even when another
// package's test run hit the same lines. This tool keeps the maximum hit count
// per block, matching go tool cover -func.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	inPath := flag.String("in", "", "input coverage profile path")
	outPath := flag.String("out", "", "output coverage profile path")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: coverdedupe -in coverage.raw.txt -out coverage.txt")
		os.Exit(2)
	}

	in, err := os.Open(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open input: %v\n", err)
		os.Exit(1)
	}
	defer in.Close()

	out, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	if err := dedupe(in, out); err != nil {
		fmt.Fprintf(os.Stderr, "dedupe: %v\n", err)
		os.Exit(1)
	}
}

func dedupe(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var mode string
	order := make([]string, 0, 4096)
	counts := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			mode = line
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return fmt.Errorf("invalid coverage line: %q", line)
		}

		key := fields[0] + " " + fields[1]
		count, err := strconv.Atoi(fields[2])
		if err != nil {
			return fmt.Errorf("invalid hit count in %q: %w", line, err)
		}

		if prev, ok := counts[key]; !ok {
			order = append(order, key)
			counts[key] = count
		} else if count > prev {
			counts[key] = count
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if mode == "" {
		return fmt.Errorf("coverage profile missing mode line")
	}

	if _, err := fmt.Fprintln(w, mode); err != nil {
		return err
	}
	for _, key := range order {
		if _, err := fmt.Fprintf(w, "%s %d\n", key, counts[key]); err != nil {
			return err
		}
	}
	return nil
}
