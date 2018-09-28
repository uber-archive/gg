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

const deplockUsage UsageError = `Usage: gg deplock/gopkglock/dl <module>
Example: deplock go.uber.org/fx@1.0

Shows the Gopkg.lock for the module in the solution, or the Gopkg.lock for the
specified version, tag, or branch.

If the most recently read Gopkg.lock has a different version of the same
package, shows the differences between the proposed and prior versions'
Gopkg.locks.
`

func deplockCommand() Command {
	return Command{
		Names: []string{
			"deplock",
			"gopkglock",
			"dl",
		},
		Usage:         deplockUsage,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
			module, err := driver.FindSolutionOrExpressModule(ctx, name, false)
			if err != nil {
				return err
			}
			if module.Deplock == NoHash {
				return fmt.Errorf("no Gopkg.lock found in module %s", module.Summary())
			}

			var cmd *exec.Cmd
			if prior, ok := driver.prev.Solution[module.Name]; ok && prior.Module.Deplock != NoHash && prior.Module.Deplock != module.Deplock {
				cmd = exec.Command("git", "diff", prior.Module.Deplock.String(), module.Deplock.String())
			} else {
				cmd = exec.Command("git", "show", module.Deplock.String())
			}
			cmd.Env = GitEnv(driver.memo.GitDir)
			cmd.Stdin = driver.in
			cmd.Stdout = driver.out
			cmd.Stderr = driver.err
			return cmd.Run()
		},
	}
}
