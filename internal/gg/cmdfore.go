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
)

const foreUsage UsageError = `Usage: gg fore

Reverts the stage to a reverted state, undoing the previous back command.
`

func foreCommand() Command {
	return Command{
		Names: []string{
			"fore",
		},
		Usage: foreUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			if len(driver.future) == 0 {
				fmt.Fprintf(driver.err, "No future states to return to.\n")
				return nil
			}
			driver.history = append(driver.history, driver.next)
			driver.next = driver.future[len(driver.future)-1]
			driver.future = driver.future[0 : len(driver.future)-1]
			return nil
		},
	}
}
