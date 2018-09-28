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

const fetchUsage UsageError = `Usage: gg fetch <module>

Forces a fetch from the git remote for a particular module.
`

func fetchCommand() Command {
	return Command{
		Names: []string{
			"fetch",
		},
		Usage:         fetchUsage,
		SuggestModule: true,
		Monadic: func(ctx context.Context, driver *Driver, name string) error {
			// Invalidate the fetch cache for this module.
			module := &Module{Name: name}
			if err := driver.memo.FinishRemote(ctx, driver.err, module); err != nil {
				return err
			}
			delete(driver.memo.Fetched, module.Remote)

			// Fetch again.
			return driver.memo.Fetch(ctx, driver.err, module, FetchMaxAttempts)
		},
	}
}
