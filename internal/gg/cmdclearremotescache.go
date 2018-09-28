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

const clearRemotesCacheUsage UsageError = `Usage: gg clear-remotes-cache/crc
Example: gg read clear-remotes-cache upgrade console

When you run "gg read", "r", "read-only", or "ro", gg reads glide.lock and
populates a memo of mappings from package name to package remote location,
based on existing dependencies.  When entries are absent in this cache, gg will
send an HTTP request to the package's domain to look up remote aliases for
vanity package domains like "gopkg.in".  These mappings are subject to change,
especially if you are in the process of setting up your own.  Using
"clear-remotes-cache" will void this cache.

However, when using "gg" offline to reconstruct a glide.lock, it can be handy
to use "gg read" to populate the cache before running "gg new".

See: gg help show-remotes-cache
`

func clearRemotesCacheCommand() Command {
	return Command{
		Names: []string{
			"clear-remotes-cache",
			"crc",
		},
		Usage: clearRemotesCacheUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			driver.memo.Remotes = make(map[string]string)
			return nil
		},
	}
}
