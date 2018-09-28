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

const initUsage UsageError = `Usage: gg init
Example: gg init

Reads all of the packages in the working copy and creates a new glide.lock and
vendor solution by choosing the newest version or master branch of any module
that provides a package imported by the working copy.

This does not use a perfect heuristic.  You may need to run tests to verify the
solution, look for conflicts in the solution, and possibly add the correct
versions manually.
`

func initCommand() Command {
	return Command{
		Names: []string{
			"init",
		},
		Usage: initUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			return driver.ExecuteArguments(ctx, "add-missing", "write")
		},
	}
}
