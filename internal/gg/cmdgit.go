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
	"os/exec"

	"github.com/google/shlex"
)

const gitUsage UsageError = `Usage: gg git <command>
Example: gg git show-ref

Runs a git command on the .gg git repository, which caches all versions of
every dependency in a refs/vendor section.
`

func gitCommand() Command {
	return Command{
		Names: []string{
			"git",
		},
		Usage: gitUsage,
		Monadic: func(ctx context.Context, driver *Driver, command string) error {
			args, err := shlex.Split(command)
			if err != nil {
				return err
			}
			cmd := exec.Command("git", args...)
			cmd.Env = GitEnv(driver.memo.GitDir)
			cmd.Stdin = driver.in
			cmd.Stdout = driver.out
			cmd.Stderr = driver.err
			return cmd.Run()
		},
	}
}
