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
)

const showSolutionUsage UsageError = `Usage: gg show-solution/ss
Example: gg read-only show-solution

Shows all of the modules on the solution stage that would be written to
glide.lock by "gg write-only" or checked out to vendor by "gg checkout".
The solution shows the hashes, versions, and references that correspond to each
module.
`

const solutionLegend = `
  T: needed for tests only.
  G: has a glide.lock.
  D: has a Gopkg.lock (dep)
  C: has a CHANGELOG.md.
`

func showSolutionCommand() Command {
	return Command{
		Names: []string{
			"show-solution",
			"ss",
		},
		Usage: showSolutionUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			modules := driver.next.Modules()
			if err := driver.memo.FinishModules(ctx, driver.err, modules); err != nil {
				return err
			}
			showSolution(driver.out, driver.next, modules)
			return nil
		},
	}
}

// showSolution writes a report describing all the modules in a solution.
func showSolution(out io.Writer, state *State, modules Modules) {
	var recommend string

	if len(state.Solution) > 0 {
		fmt.Fprintf(out, "Locked modules: (%d)\n", len(state.Solution))
		for _, name := range state.Solution.Names() {
			module := state.Solution[name].Module
			conflict := ""
			if modules.Conflicts(module) {
				conflict = " " + yellow + "(conflict)" + clear
				recommend = fmt.Sprintf("📎 "+yellow+"Looks like there is a conflict."+clear+" gg show-module %s for details.\n", module.Name)
			}
			fmt.Fprintf(out, "* %s%s\n", module, conflict)
		}
	}

	if len(state.Frontier) > 0 {
		fmt.Fprintf(out, "Unlocked modules: (%d)\n", len(state.Frontier))
		for _, module := range state.Frontier {
			fmt.Fprintf(out, "* %s\n", module)
		}
	}

	fmt.Fprintf(out, "%s", solutionLegend[1:])
	fmt.Fprintf(out, "%s", recommend)
}
