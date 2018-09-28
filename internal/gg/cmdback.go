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

const backUsage UsageError = `Usage: gg back

Reverts the stage to the prior solution, undoing the last command to change the
state.
`

func backCommand() Command {
	return Command{
		Names: []string{
			"back",
		},
		Usage: backUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			if len(driver.history) == 0 {
				driver.next = NewState()
				return nil
			}
			driver.future = append(driver.future, driver.next)
			driver.next = driver.history[len(driver.history)-1]
			driver.history = driver.history[0 : len(driver.history)-1]
			return nil
		},
	}
}
