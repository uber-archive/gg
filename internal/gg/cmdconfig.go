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

const configUsage UsageError = `File: gg.toml

gg reads a gg.toml configuration file from the working directory or any parent
directory.  The configuration file can supply package name patterns to look up
remote locations, recommend versions for the add-missing modules workflow, and
exclude directories from the package import graph analysis of the working copy.

For example, the following block directs gg to use a mirror for projects on
Github.  Specifying the */* pattern helps gg infer that a longer package path
like github.com/x/y/z must come from module github.com/x/y, and avoid
attempting to fetch github.com/x/y/z as a whole repository.
With the gitoliteMirror flag, this directs gg to attempt to create the mirror
before fetching, in case it does not yet exist.

	[[remotes]]
	pattern = "github.com/*/*"
	remote = "mirror.com/github/*/*"
	gitoliteMirror = true

Using the "..." wildcard slurps the entire remaining path and spits it back
out.  This is effectively the same as any number of stars, but is more flexible
for domains that have repository paths at varying depths.

	[[remotes]]
	pattern = "example.com/..."
	remote = "mirror.com/example.com/..."

Since gg.toml can be in any parent directory, we can use it for
organization-wide hints about common version ranges.  The add-missing modules
workflow will favor the specified version over just using the most recent
version.  The version is a semantic version, so specifying "0.10" means the
most recent patch less than "0.11", whereas "1" would mean the highest version
less than "2.0.0".

	[[packages]]
	package = "git.apache.org/thrift"
	version = "0.10"

gg detects missing packages in the working copy by searching for packages
imported by go files.  It skips directories like .git, .gg, and vendor.
You can configure additional directories to exclude with the excludes section.

	[[excludes]]
	path = "go-build"

gg collects all of your project's dependencies in a local git repository called
.gg as a cache, populating its refs/vendor namespace.  These references can be
pushed to a remote repository.  gg will automatically fetch from this cache
before fetching from individual repositories if provided a cache repository URL
in gg.toml.

	cache = "https://example.com/my/cache"
`

func configCommand() Command {
	return Command{
		Names: []string{
			"config",
		},
		Usage: configUsage,
		Niladic: func(ctx context.Context, driver *Driver) error {
			return driver.ExecuteArguments(ctx, "help", "config")
		},
	}
}
