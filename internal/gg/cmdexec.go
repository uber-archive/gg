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
	"os"
)

const execUsage UsageError = `Usage: gg exec/x <command>
Example: gg read checkout exec 'go test .'

Executes a shell command.  When specified on the command line, the command must
be quoted as a single argument and the workflow will abort if the command exits
with a nonzero exit status code.  This is useful to avoid writing glide.lock if
tests fail, for example.  When specified in the console, all text that follows
the "exec" or "x" token is executed in a shell.  Exec uses the SHELL
environment variable or uses "sh".

The "x" command without an argument in the console is an alias for "exit" to
avoid aggrevating newcomers.
`

func execCommand() Command {
	return Command{
		Names: []string{
			"exec",
			"x",
		},
		Usage: execUsage,
		Monadic: func(ctx context.Context, driver *Driver, command string) error {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "sh"
			}
			return driver.ShellExec(shell, "-c", command)
		},
	}
}
