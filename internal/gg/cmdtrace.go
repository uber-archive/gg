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

const traceUsage UsageError = `Usage: gg trace/tr/t <package>
Example: gg read show-packages trace go.uber.org/fx

Traces the shortest sequence of imports from a command or test package in the
working copy to the target package.
`

func traceCommand() Command {
	return Command{
		Names: []string{
			"trace",
			"tr",
			"t",
		},
		Usage:          traceUsage,
		Read:           true,
		SuggestPackage: true,
		Monadic: func(ctx context.Context, driver *Driver, from string) error {
			name, ownPackages, err := driver.memo.ReadOwnPackages(ctx, driver.err)
			if err != nil {
				return err
			}

			state := driver.next
			modules := state.Modules()
			if err := driver.memo.FinishPackages(ctx, driver.err, modules); err != nil {
				return err
			}
			return Trace(driver.out, name, ownPackages, modules.Packages(), from)
		},
	}
}
