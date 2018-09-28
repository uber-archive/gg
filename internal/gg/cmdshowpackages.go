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
	"strings"
)

const showPackagesLegend = `
Flags:
c--- indicates a "main" package, for building a command.
-e-- indicates a package that exists in the working copy.
--i- indicates a package that another package imports.
---t indicates a package that another package imports for tests.
`

const showPackagesUsage UsageError = `Usage: gg show-packages/sp
Example: gg read show-packages

Reads the working copy and every package in the dependency solution and
produces a list of all the packages, with flags.

At the console, show-packages accepts a prefix argument and filters
for packages with that prefix.

gg> sp gopkg.in/
` + showPackagesLegend

const showOwnPackagesUsage UsageError = `Usage: gg show-own-packages/sop
Example: gg show-own-packages

Reads the working copy and shows all of its packages, with flags.
` + showPackagesLegend

func showPackagesCommand() Command {
	return Command{
		Names: []string{
			"show-packages",
			"sp",
		},
		Usage:             showPackagesUsage,
		Read:              true,
		SuggestPackage:    true,
		OptionallyMonadic: true,
		Monadic: func(ctx context.Context, driver *Driver, prefix string) error {
			return driverShowPackages(ctx, driver, prefix)
		},
	}
}

func showOwnPackagesCommand() Command {
	return Command{
		Names: []string{
			"show-own-packages",
			"sop",
		},
		Usage: showOwnPackagesUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			name, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}
			if name != "" {
				fmt.Fprintf(driver.out, "Name: %s\n", name)
			}
			showPackages(driver.out, packages, "")
			return nil
		},
	}
}

func driverShowPackages(ctx context.Context, driver *Driver, prefix string) error {
	_, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
	if err != nil {
		return err
	}
	modules := driver.next.Modules()
	if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
		return err
	}
	packages = packages.Clone()
	packages.Include(modules.Packages())
	showPackages(driver.out, packages, prefix)
	return nil
}

// showPackages writes a report of what packages are in an import graph, with a
// given prefix.
func showPackages(out io.Writer, packages Packages, prefix string) {
	for _, pkg := range packages.All.Keys() {
		if strings.HasPrefix(pkg, prefix) {
			fmt.Fprintf(out, "%s\n", pkg)
		}
	}
}
