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
	"os/exec"
)

const changelogUsage UsageError = `Usage: gg changelog/cl <module>
Example: changelog go.uber.org/fx@1.0

Shows the changelog for the module in the solution, or the changelog for the
specified version, tag, or branch.

If the most recently read glide.lock has a different version of the same
package, shows the differences between the proposed and prior versions' change
logs.
`

func changelogCommand() Command {
	return Command{
		Names: []string{
			"changelog",
			"cl",
		},
		Usage:         changelogUsage,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
			module, err := driver.FindSolutionOrExpressModule(ctx, name, false)
			if err != nil {
				return err
			}
			if module.Changelog == NoHash {
				return fmt.Errorf("no changelog found in module %s", module.Summary())
			}

			var cmd *exec.Cmd
			if prior, ok := driver.prev.Solution[module.Name]; ok && prior.Module.Changelog != NoHash && prior.Module.Changelog != module.Changelog {
				cmd = exec.Command("git", "diff", prior.Module.Changelog.String(), module.Changelog.String())
			} else {
				cmd = exec.Command("git", "show", module.Changelog.String())
			}
			cmd.Env = GitEnv(driver.memo.GitDir)
			cmd.Stdin = driver.in
			cmd.Stdout = driver.out
			cmd.Stderr = driver.err
			return cmd.Run()
		},
	}
}
