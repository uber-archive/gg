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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config is the schema for the gg.toml file, which configures gg when running
// in a child directory.
type Config struct {
	// Cache is the git URL of a git repository that serves as a refs/vendor
	// cache.
	Cache string `toml:"cache"`
	// Remotes override the default behavior for finding the remote repository
	// for modules that have matching name patterns.
	Remotes []ConfigRemote `toml:"remotes"`
	// Packages overrides the default version that gg will give in the
	// add-missing modules workflow.
	Packages []ConfigPackage `toml:"packages"`
	// Excludes adds paths to the list of directories to ignore in the working
	// copy directory tree, to discover the project's package import graph.
	Excludes []ConfigExclude `toml:"excludes"`
}

// ConfigRemote specifies the remote repository location pattern to use for
// matching module name patterns, instead of going to the web or inferring from
// the first component of the module name.
type ConfigRemote struct {
	// Pattern is a glob-like pattern that matches a module name, and may
	// include * for wild path components, or ... for any suffix.
	Pattern string `toml:"pattern"`
	// Remote is a glob-like pattern that indicates the corresponding remote
	// repository location, with corresponding path components from the
	// matching name.
	Remote string `toml:"remote"`
	// GitoliteMirror indicates that the remote is a gitolite mirror and may
	// need to be created with an ssh create command.
	GitoliteMirror bool `toml:"gitoliteMirror"`
}

// ConfigPackage specifies the version of a module to add to a solution in the
// add-missing workflow.
// The default behavior is to get the newest version, or failing that, the
// master branch.
type ConfigPackage struct {
	// Package is a module name.
	Package string `toml:"package"`
	// Version is a version number like "1" or "v1.2.3".
	Version string `toml:"version"`
}

// ConfigExclude specifies a directory name to exclude when searching for go
// files in the working copy tree.
type ConfigExclude struct {
	Path string `toml:"path"`
}

// ReadConfig reads gg.toml from bytes.
func ReadConfig(bytes []byte) (*Config, error) {
	var config Config
	err := toml.Unmarshal(bytes, &config)
	return &config, err
}

// ReadOwnConfig reads gg.toml from the working directory or any ancestor
// directory.
// The youngest ancestor directory shadows all others.
func ReadOwnConfig(workDir string) (*Config, error) {
	for {
		var file *os.File
		file, err := os.Open(filepath.Join(workDir, "gg.toml"))
		if err != nil {
			workDir, _ = filepath.Split(strings.TrimSuffix(workDir, string(os.PathSeparator)))
			if workDir == "" {
				return &Config{}, nil
			}
			continue
		}

		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return &Config{}, err
		}
		return ReadConfig(bytes)
	}
}

// ReadPatterns converts the gg.toml structure to pattern objects.
func (config *Config) ReadPatterns() Patterns {
	patterns := make(Patterns, 0, len(config.Remotes))
	for _, remote := range config.Remotes {
		patterns = append(patterns, Pattern{
			Match:   PatternSplit(remote.Pattern),
			Replace: PatternSplit(remote.Remote),
		})
	}
	return patterns
}

// ReadGitoliteMirrors collects a set of all module names with Gitolite
// mirrors.
func (config *Config) ReadGitoliteMirrors() map[int]struct{} {
	mirrors := make(map[int]struct{}, len(config.Remotes))
	for rule, remote := range config.Remotes {
		if remote.GitoliteMirror {
			mirrors[rule] = struct{}{}
		}
	}
	return mirrors
}

// ReadExcludes collects the directory names to exclude from the working copy
// tree.
func (config *Config) ReadExcludes() StringSet {
	excludes := make(StringSet)
	for _, exclude := range config.Excludes {
		excludes.Add(exclude.Path)
	}
	return excludes
}

// ReadRecommended collects the recommended versions for new versions of a
// known module.
func (config *Config) ReadRecommended() map[string]Version {
	recs := make(map[string]Version)
	for _, recommend := range config.Packages {
		recs[recommend.Package] = ParseVersion(recommend.Version)
	}
	return recs
}
