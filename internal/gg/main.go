// Copyright (c) 2018 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gg

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

// Main runs gg in the OS environment.
func Main() error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	return Environment{
		WorkDir:    workDir,
		Getenv:     os.Getenv,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		Filesystem: osfs.New("/"),
	}.Run(context.Background())
}

// Environment represents an environment for running gg, either normally or in
// a test.
type Environment struct {
	WorkDir    string
	Getenv     func(string) string
	Stdin      io.ReadCloser
	Stdout     io.Writer
	Stderr     io.Writer
	Filesystem billy.Filesystem
}

// Run runs the gg command in an environment and context.
func (env Environment) Run(ctx context.Context) error {
	// Build a memo.
	goPath := strings.Split(env.Getenv("GOPATH"), ":")
	gitDir := path.Join(env.WorkDir, GGCachePath)
	memo, err := NewMemo(gitDir, env.WorkDir, goPath)
	if err != nil {
		return err
	}

	if err := memo.ReadConfig(); err != nil {
		return err
	}

	// Bypass command line if there is a gg server running and injected into
	// the environment.
	addr := env.Getenv("GG_ADDR")
	if addr != "" {
		return RunClient(memo, addr, os.Args[1:])
	}

	driver, err := NewDriver(memo, env.Stdin, env.Stdout, env.Stderr)
	if err != nil {
		return err
	}

	if len(os.Args) <= 1 {
		fmt.Fprintf(env.Stdout, "%s", ggUsage)
		return nil
	}

	return driver.ExecuteArguments(ctx, os.Args[1:]...)
}
