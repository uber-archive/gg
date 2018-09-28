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

const pushUsage UsageError = `Usage: gg push [<remote>]

gg collects all the versions of your project's transitive dependencies in a git
repository called .gg and namespaces their versions under refs/vendor.  The
push command sends the vendored references to your configured or specified
remote cache.

See: gg help config
See: gg help pull
`

func pushCommand() Command {
	return Command{
		Names: []string{
			"push",
		},
		Usage:             pushUsage,
		OptionallyMonadic: true,
		Monadic: func(ctx context.Context, driver *Driver, remote string) error {
			if remote == "" {
				remote = driver.memo.VendorCache
			}
			if remote == "" {
				return fmt.Errorf("cache location must be specified on command or in gg.toml")
			}
			return GitPushVendorCache(driver.err, driver.memo.GitDir, remote)
		},
	}
}
