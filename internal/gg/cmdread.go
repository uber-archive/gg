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

const readUsage UsageError = `Usage: gg read/r
Example: gg read show-solution

Reads a glide.lock onto the stage and solves its dependency graph, collecting
unmentioned transitive dependencies and choosing the highest version mentioned.
`

func readCommand() Command {
	return Command{
		Names: []string{
			"read",
			"r",
		},
		Usage: readUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			return driver.ExecuteArguments(ctx, "read-glide-lock", "solve")
		},
	}
}
