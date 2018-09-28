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

const showShallowSolutionUsage UsageError = `Usage: gg show-shallow-solution/sss
Example: gg read sss

Reads all packages in the working copy and lists all of the modules that
exports one of the modules the working copy directly depends upon.
This is the basis for writing a manifest file like glide.yaml or Gopkg.toml,
which implies transitive dependencies.
`

func showShallowSolutionCommand() Command {
	return Command{
		Names: []string{
			"show-shallow-solution",
			"sss",
		},
		Usage: showShallowSolutionUsage,
		Read:  true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			state := driver.next
			modules := state.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}

			shallow := ShallowSolution(packages, modules)

			showShallowSolution(driver.out, shallow, packages)
			return nil
		},
	}
}

// showShallowSolution writes a report of all the modules that provide
// packages that are directly imported by commands and tests in the working copy.
// These are the modules you would expect in a manifest file, that does not
// capture the transitive dependencies of the working copy.
func showShallowSolution(out io.Writer, modules Modules, packages Packages) {
	fmt.Fprint(out, "Manifest modules (direct dependencies):\n")
	if len(modules) == 0 {
		fmt.Fprintf(out, "* No dependencies.\n")
	}
	for _, module := range modules {
		fmt.Fprintf(out, "* %s\n", module)
	}
}
