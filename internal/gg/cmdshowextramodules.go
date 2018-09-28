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

const showExtraModulesUsage UsageError = `Usage: gg show-extra-modules/sxm
Example: gg read show-extra-modules

Reads all the packages in the working copy and traces all their imports
to packages in the currently proposed solution, then shows all the modules
that provide no packages that the solution uses.
`

func showExtraModulesCommand() Command {
	return Command{
		Names: []string{
			"show-extra-modules",
			"sxm",
		},
		Usage: showExtraModulesUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}
			ShowExtraModules(driver.out, driver.next, packages)
			return nil
		},
	}
}

// ShowExtraModules writes a report of what modules in the solution are not
// needed to build the commands and tests in the working copy.
func ShowExtraModules(out io.Writer, state *State, ownPackages Packages) {
	packages := state.Modules().Packages()
	packages.Include(ownPackages)
	modules := ExtraModules(ownPackages, packages, state.Modules())
	fmt.Fprint(out, "Extra modules:\n")
	if len(modules) == 0 {
		fmt.Fprintf(out, "* No extra modules.\n")
	}
	for _, module := range modules {
		fmt.Fprintf(out, "* %s\n", module)
	}
}
