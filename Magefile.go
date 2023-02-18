//go:build mage

package main

import (
	"context"
	"os"
	"os/exec"
)

func Build(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "../../bin/demo")
	cmd.Env = []string{
		"GOOS=linux",
		"GOARCH=amd64",
	}
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "cmd/demo"
	return cmd.Run()
}
