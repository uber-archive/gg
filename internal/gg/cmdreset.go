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

const resetUsage UsageError = `Usage: gg reset

Resets the last read state to an empty staged solution.  Diff reports use this
as the former state.
`

func resetCommand() Command {
	return Command{
		Names: []string{
			"reset",
		},
		Usage: resetUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			driver.prev = NewState()
			return nil
		},
	}
}
