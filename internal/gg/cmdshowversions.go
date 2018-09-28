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
)

const showVersionsUsage UsageError = `Usage: gg show-versions/sv <module>
Example: gg show-versions go.uber.org/fx

Show versions will fetch a module from its remote location (if possible)
then produce a list of all known versions of that module in its current
git repository.
`

func showVersionsCommand() Command {
	return Command{
		Names: []string{
			"show-versions",
			"sv",
		},
		Usage:         showVersionsUsage,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
			memo := driver.memo
			module := Module{
				Name: name,
			}
			if err := memo.FinishRemote(ctx, driver.err, &module); err != nil {
				fmt.Fprintf(driver.err, "warning: Unable to determine remote location for %s: %q\n", name, err)
			}
			if err := memo.Fetch(ctx, driver.err, &module, FetchMaxAttempts); err != nil {
				fmt.Fprintf(driver.err, "warning: Unable to fetch current versions of %s: %q\n", name, err)
			}
			if err := memo.DigestRefs(ctx, driver.err, module); err != nil {
				fmt.Fprintf(driver.err, "warning: Unable to digest current references to %s: %q\n", name, err)
			}

			modules, err := memo.ReadVersions(ctx, driver.err, module)
			if err != nil {
				return fmt.Errorf("unable to read latest versions of %s: %s", name, err)
			}
			fmt.Fprintf(driver.out, "Versions:\n")
			if len(modules) == 0 {
				fmt.Fprintf(driver.out, "* No versions.\n")
			}
			for _, module := range modules {
				fmt.Fprintf(driver.out, "* %s\n", module)
			}

			return nil
		},
	}
}
