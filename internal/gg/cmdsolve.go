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

const solveUsage UsageError = `Usage: gg solve/s
Example: gg read-only solve show-solution

Collects the transitive minimum dependencies of the modules in the stage and
finds a solution that takes the highest of all specified modules.  This may
involve discovering and fetching previously unknown dependencies, digesting all
of their references and building a model of package imports.

If executed as the only argument at the command line, this will read, solve,
then write back.
`

func solveCommand() Command {
	return Command{
		Names: []string{
			"solve",
			"s",
		},
		Usage: solveUsage,
		Write: true,
		Niladic: func(ctx context.Context, driver *Driver) error {
			memo := driver.memo
			state := driver.next

			msg := "Solving dependency graph"
			driver.err.Start(msg)
			next, err := state.Solve(ctx, memo, driver.err)
			driver.err.Stop(msg)
			if err != nil {
				return err
			}
			state = next

			driver.push(state)
			return nil
		},
	}
}
