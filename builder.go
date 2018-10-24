package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Builder struct {
	dir       string
	binary    string
	errors    string
	useGodep  bool
	wd        string
	buildArgs []string
}

// NewBuilder creates new builder
func NewBuilder(dir string, bin string, useGodep bool, wd string, buildArgs []string) *Builder {
	if len(bin) == 0 {
		bin = "bin"
	}

	// We need to append .exe in Windows
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(bin, ".exe") { // check if it already has the .exe extension
			bin += ".exe"
		}
	}

	return &Builder{dir: dir, binary: bin, useGodep: useGodep, wd: wd, buildArgs: buildArgs}
}

func (b *Builder) Binary() string {
	return b.binary
}

func (b *Builder) Errors() string {
	return b.errors
}

func (b *Builder) Build() error {
	args := append([]string{"go", "build", "-o", filepath.Join(b.wd, b.binary)}, b.buildArgs...)

	var command *exec.Cmd
	if b.useGodep {
		args = append([]string{"godep"}, args...)
	}
	command = exec.Command(args[0], args[1:]...)

	output, err := command.CombinedOutput()

	if err != nil {
		b.errors = err.Error() + ": " + string(output)
	} else if command.ProcessState.Success() {
		b.errors = ""
	} else {
		b.errors = string(output)
	}

	if len(b.errors) > 0 {
		return fmt.Errorf(b.errors)
	}

	return err
}
