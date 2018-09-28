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

const showMissingPackagesUsage UsageError = `Usage: gg show-missing-packages/smp
Example: gg read show-missing-packages

Reads all of the packages in the working copy and traces all of their imports
to packages in the currently proposed solution, then shows all of the packages
that the solution does not satisfy.  A package is missing if it is needed
either to build a command or run a test.  So, if a package is a transitive
dependency of a "main" package in your working copy, or a transitive dependency
of any "_test" file in your working copy, and is absent in both the working
copy and any of the modules in the staged solution, it will be reported
missing.
`

func showMissingPackagesCommand() Command {
	return Command{
		Names: []string{
			"show-missing-packages",
			"smp",
		},
		Usage: showMissingPackagesUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}
			ShowMissingPackages(driver.out, driver.next, packages)
			return nil
		},
	}
}

// ShowMissingPackages writes a report of which packages are missing from a
// solution to satisfy the transitive imports of the commands and tests in the
// working copy.
func ShowMissingPackages(out io.Writer, state *State, packages Packages) {
	imports, testImports := MissingPackages(packages, state.Modules().Packages())
	missingImports := imports.Keys()
	missingTestImports := testImports.Keys()
	if len(missingImports) == 0 {
		fmt.Fprintf(out, "No missing packages.\n")
	} else {
		fmt.Fprintf(out, "Missing packages:\n")
		for _, pkg := range missingImports {
			fmt.Fprintf(out, "- %s\n", pkg)
		}
	}
	if len(missingTestImports) == 0 {
		fmt.Fprintf(out, "No missing packages for tests.\n")
	} else {
		fmt.Fprintf(out, "Missing packages for tests:\n")
		for _, pkg := range missingTestImports {
			fmt.Fprintf(out, "- %s\n", pkg)
		}
	}
}
