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

const showDependenciesUsage UsageError = `Usage: gg show-dependencies/sd
Example: gg read-only show-dependencies

Shows a list of all the dependencies involved in the solution.  For each
dependency, it shows which modules depended on it, what version they requested,
and what version the resolver chose.
`

func showDependenciesCommand() Command {
	return Command{
		Names: []string{
			"show-dependencies",
			"sd",
		},
		Usage: showDependenciesUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			ShowDependencies(driver.out, driver.next.Modules())
			return nil
		},
	}
}

// ShowDependencies writes a report about all of the dependencies expressed in
// a solution, including all the constraints expressed for each module in the
// solution.
func ShowDependencies(out io.Writer, modules Modules) {
	fmt.Fprintf(out, "Dependencies:\n")
	for _, module := range modules {
		fmt.Fprintf(out, "* %s\n", module.Name)
		fmt.Fprintf(out, "  ON %s\n", module)
		for _, dependee := range modules {
			for _, dependency := range dependee.Modules {
				if dependency.Name == module.Name {
					fmt.Fprintf(out, "  AS %s BY %s\n", dependency, dependee.Name)
				}
			}
		}
	}
}
