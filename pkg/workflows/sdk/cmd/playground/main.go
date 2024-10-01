package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var port = flag.String("port", ":8080", "port to run on")

//go:embed www
var www embed.FS

func main() {
	flag.Parse()

	files, err := fs.Sub(www, "www")
	if err != nil {
		log.Fatalln("Failed to strip filesystem prefix www/:", err)
	}
	http.Handle("/", http.FileServerFS(files))
	http.Handle("/format-chart", http.HandlerFunc(formatChart))
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatalln(err)
	}
}

func formatChart(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	dir, err := writeFiles(req.Context(), req.Body)
	if err != nil {
		log.Println("Failed to write files:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	formatted, err := os.ReadFile(filepath.Join(dir, "workflow.go"))
	if err != nil {
		log.Println("Failed to read workflow file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		Chart    string `json:"chart"`
		Workflow string `json:"workflow"`
		Error    string `json:"error"`
	}{Workflow: string(formatted)}

	cmd := exec.CommandContext(req.Context(), "go", "run", "main.go", "workflow.go")
	cmd.Dir = dir
	chart, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			err = fmt.Errorf("%s: %s", exitErr.ProcessState, exitErr.Stderr)
		}
		log.Println("Failed to format chart:", err)
		resp.Error = err.Error()
	} else {
		resp.Chart = string(chart)
		log.Println("Formatted chart:", resp.Chart)
	}

	e := json.NewEncoder(w)
	err = e.Encode(resp)
	if err != nil {
		log.Println("Failed to write JSON response:", err)
	}
}

func writeFiles(ctx context.Context, req io.Reader) (dir string, err error) {
	dir, err = os.MkdirTemp("", "format-chart")
	if err != nil {
		err = fmt.Errorf("failed to create temp dir: %s", err)
		return
	}

	b, err := io.ReadAll(req)
	if err != nil {
		err = fmt.Errorf("failed to read request: %s", err)
		return
	}

	err = os.WriteFile(filepath.Join(dir, "workflow.go"), b, 0666)
	if err != nil {
		err = fmt.Errorf("failed to write workflow.go: %s", err)
		return
	}
	err = os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0666)
	if err != nil {
		err = fmt.Errorf("failed to write main.go: %s", err)
		return
	}
	err = os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0666)
	if err != nil {
		err = fmt.Errorf("failed to write go.mod: %s", err)
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to get working directory: %s", err)
		return
	}
	var goWork bytes.Buffer
	fmt.Fprintf(&goWork, `go 1.23

use %s`, wd)
	err = os.WriteFile(filepath.Join(dir, "go.work"), goWork.Bytes(), 0666)
	if err != nil {
		err = fmt.Errorf("failed to write go.work: %s", err)
		return
	}

	cmd := exec.CommandContext(ctx, "goimports", "-w", "workflow.go")
	cmd.Dir = dir
	_, err = cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			err = fmt.Errorf("%s: %s", exitErr.ProcessState, exitErr.Stderr)
		}
		log.Println("Failed to run goimports:", err)
	}

	return
}

const mainGo = `package main

import (
	"fmt"
	"log"
)

func main() {
	spec, err := buildWorkflow().Spec()
	if err != nil {
		log.Fatalln("Failed to get spec:", err)
	}
	chart, err := spec.FormatChart()
	if err != nil {
		log.Fatalln("Failed to format chart:", err)
	}
	fmt.Println(chart)
}
`

const goMod = `module example.workflow/builder

go 1.23

require github.com/smartcontractkit/chainlink-common v0.0.0-00010101000000-000000000000
`
