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

const showImportsUsage UsageError = `Usage: gg show-imports/si <package>
Example: gg read add go.uber.org/fx show-imports go.uber.org/fx

Shows lists of packages related to the given package, from anywhere
in the current dependency solution or the working copy.
1. A list of the packages that package imports.
2. A list of the packages that package imports for tests.
3. A list of the packages that import that package.
4. A list of the packages that import that package for tests.
For a list of subpackages, use "gg show-packages <package>" or
"gg sp <package>".
`

func showImportsCommand() Command {
	return Command{
		Names: []string{
			"show-imports",
			"si",
		},
		Usage:          showImportsUsage,
		SuggestPackage: true,
		Read:           true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
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
			ShowImports(driver.out, packages, name)
			return nil
		},
	}
}

// ShowImports writes a report covering the cross-section of the import graph
// that concerns a particular package: which packages import it or import it
// just for tests, and which packages it imports or imports just for tests.
func ShowImports(out io.Writer, packages Packages, name string) {
	imports := packages.Imports[name]
	if len(imports) > 0 {
		fmt.Fprintf(out, "Packages that %s imports:\n", name)
	}
	for _, imp := range imports.Keys() {
		fmt.Fprintf(out, "* %s\n", imp)
	}

	testImports := packages.TestImports[name]
	if len(testImports) > 0 {
		fmt.Fprintf(out, "Packages that %s imports for tests:\n", name)
	}
	for _, imp := range testImports.Keys() {
		fmt.Fprintf(out, "* %s\n", imp)
	}

	coImports := packages.CoImports[name]
	if len(coImports) > 0 {
		fmt.Fprintf(out, "Packages that import %s:\n", name)
	}
	for _, imp := range coImports.Keys() {
		fmt.Fprintf(out, "* %s\n", imp)
	}

	coTestImports := packages.CoTestImports[name]
	if len(coTestImports) > 0 {
		fmt.Fprintf(out, "Test packages that import %s:\n", name)
	}
	for _, imp := range coTestImports.Keys() {
		fmt.Fprintf(out, "* %s\n", imp)
	}
}
