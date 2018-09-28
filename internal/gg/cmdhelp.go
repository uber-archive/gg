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

const ggUsage UsageError = `A go git dependency management suite.
Runs a sequence of dependency management commands until one fails.
Usage: gg <command>..., gg console
Example:
  gg read upgrade show-solution checkout exec 'make test'
  gg r    u       ss            co       x    'make test'
Workflows:
  init (adds missing dependencies, solves, writes vendor and lockfile)
  i/install (reads lockfile, checks out vendor)
  solve (collects and merges transitive lockfiles)
  up/upgrade (upgrades modules in lockfile based their versions)
  am/add-missing (adds arbitrary modules to satisfy missing imports)
  prune (removes dependencies unless imported)
Files:
  r/read                     w/write
  rgy/read-glide-yaml        wgy/write-glide-yaml
  rgl/read-glide-lock        wgl/write-glide-lock
  rdl/read-dep-lock          wdl/write-dep-lock
  rdt/read-dep-toml          wdt/write-dep-toml
  gl/glidelock <module>      dl/deplock <module>
  cl/changelog <module>      co/checkout
Observe:
  diff
  ss/show-solution           sc/show-conflicts
  sm/show-module <module>    si/show-imports <package>
  sv/show-versions <module>  trace <package>
  smp/show-missing-packages  sxm/show-extra-modules
  sop/show-own-packages      sss/show-shallow-solution
Orient:
  new  mark  reset  back  fore  off/offline  on/online  quiet
Cache:
  push  pull  fetch  src/show-remotes-cache  crc/clear-remotes-cache
Decide: (you are here)
Act:
  a/add <module>    at/add-test <module>  rm/remove <module>
  x/exec <command>  sh/shell              git <command>
Debugging: metrics, cpuprofile...
Help: help, help <topic>, help config
`

func helpCommand() Command {
	return Command{
		Names: []string{
			"help",
			"h",
		},
		Usage: ggUsage,
		Monadic: func(ctx context.Context, driver *Driver, cmd string) error {
			usage, ok := driver.help[cmd]
			if !ok {
				fmt.Fprintf(driver.err, "%s", ggUsage)
				return fmt.Errorf("unrecognized command: %q", cmd)
			}
			fmt.Fprintf(driver.out, "%s", usage)
			return nil
		},
	}
}
