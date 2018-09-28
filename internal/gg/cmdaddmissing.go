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

import "context"

const addMissingUsage UsageError = `Usage: gg add-missing/am
Example: gg add-missing

Adds missing modules to the solution by finding modules that export packages
that are needed to build commands and tests in the working copy.

This command does not guarantee that it will choose the correct version, so you
should run this, check out the solution, and run your tests to verify the
result.  You can fall back on manually adding the correct versions.
`

func addMissingCommand() Command {
	return Command{
		Names: []string{
			"add-missing",
			"am",
		},
		Usage: addMissingUsage,
		Write: true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			name, packages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			memo := driver.memo
			state := driver.next

			recommended := memo.Recommended
			state, err = AddMissing(ctx, memo, driver.err, state, name, packages, recommended)
			if err != nil {
				return err
			}
			driver.push(state)

			ShowDiff(driver.out, driver.prev.Modules(), driver.next.Modules())
			return nil
		},
	}
}
