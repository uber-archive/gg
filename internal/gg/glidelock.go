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
	"time"

	yaml "gopkg.in/yaml.v2"
)

// GlideLock is a model of the root object of a glide.lock in YAML format.
// GG extends the glide.lock format to cache as much as possible that can be
// deterministically inferred from a commit hash and the corresponding git
// objects, and elides the properties of glide.lock, like an update timestamp,
// that would make the hash of the glide.lock unsuitable as a cache key.
type GlideLock struct {
	// Updated timestamp deliberately omitted since it would invalidate the
	// deterministic mapping from glide.lock content hash to vendor tree hash.
	Generator   string            `yaml:"generator"`
	Imports     []GlideLockImport `yaml:"imports,omitempty"`
	TestImports []GlideLockImport `yaml:"testImports,omitempty"`
}

// GlideLockImport is a model of an imported module or test module from a
// glide.lock.
type GlideLockImport struct {
	GlideLockRequirement `yaml:",inline"`

	Warnings []string `yaml:"warnings,omitempty"`
	// Changelog is the hash of a CHANGELOG.md file in the git repository, if
	// present.
	Changelog string `yaml:"changelog,omitempty"`
	// Glidelock is the hash of a glide.lock file in the git repository, if
	// present.
	Glidelock string `yaml:"glidelock,omitempty"`
	// Commands is specific to gg and lists all of the "main" packages in the
	// module.

	// GitoliteMirror indicates that the remote is a gitolite mirror.
	GitoliteMirror bool `yaml:"gitoliteMirror,omitempty"`
	// GitoliteMirrorCreated indicates that the gitolite mirror was already created.
	GitoliteMirrorCreated bool `yaml:"gitoliteMirrorCreated,omitempty"`

	// Requirements are a cache of the lock file provided by the dependency.
	Requirements []GlideLockRequirement `yaml:"requirements,omitempty"`
	// NoRequirements indicates that glide.lock nor Gopkg.in exist in the module.
	NoRequirements bool `yaml:"noRequirements,omitempty"`

	Commands []string `yaml:"commands,omitempty"`
	// Exports is specific to gg, may or may not include the eponymous package
	// name, and differs from Subpackages because prefix is included to
	// distinguish whether an eponymous package is actually exported.
	Exports []string `yaml:"exports,omitempty"`
	// Imports is specific to gg and captures the imports of every package
	// exported by this module.
	Imports map[string][]string `yaml:"imports,omitempty"`
	// TestImports is specific to gg and captures the test imports of every
	// package exported by this module.
	TestImports map[string][]string `yaml:"testImports,omitempty"`
}

// GlideLockRequirement models a requirement in a glide.lock.
type GlideLockRequirement struct {
	// Name is standard to Glide and corresponds to the module.Name name
	// that is or is the prefix of all packages exported by this module.
	// We infer the repository location from the package name, possibly using
	// an HTTP request to translate a vanity name to a remote repository
	// location.
	Name string `yaml:"name"`
	// Root is specific to gg and caches the normalized cache key for the
	// remote URL.
	Root string `yaml:"root,omitempty"`
	// Repo is standard to Glide and corresponds to the module.Remote URL.
	Repo string `yaml:"repo,omitempty"`
	// Version is the commit hash of the locked dependency.
	Version string `yaml:"version,omitempty"`
	// Revision is specific to gg and captures the normalized version number
	// corresponding to the commit, if gg infers the version from a tag name.
	Revision string `yaml:"revision,omitempty"`
	// Ref is specific to gg and captures the ref corresponding to the tag that
	// represents the highest version number, or the lexicographically highest
	// branch name that refers back to the corresponding git commit.
	Ref string `yaml:"ref,omitempty"`
	// Time is specific to gg and caches the timestamp of the corresponding git
	// commit.
	Time time.Time `yaml:"time,omitempty"`
	// VCS is de-facto standard, but gg requires it to be "git" if present,
	// and leaves it absent when updating to normalize.
	VCS string `yaml:"vcs,omitempty"`
}

// ReadOwnGlideLock reads the glide.lock in the working directory.
func ReadOwnGlideLock() (*GlideLock, error) {
	bytes, err := ioutil.ReadFile("glide.lock")
	if err != nil {
		return &GlideLock{}, err
	}
	return ReadGlideLock(bytes)
}

// WriteOwnGlideLock writes the glide.lock in the working directory.
func WriteOwnGlideLock(lock *GlideLock) error {
	bytes, err := WriteGlideLock(lock)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("glide.lock", bytes, 0644)
}

// ReadGlideLock parses the glide.lock format from the given data.
func ReadGlideLock(bytes []byte) (*GlideLock, error) {
	var l GlideLock
	err := yaml.Unmarshal(bytes, &l)
	if err != nil {
		return &l, err
	}
	return &l, nil
}

// WriteGlideLock formats the glide.lock into the returned byte slice.
func WriteGlideLock(lock *GlideLock) ([]byte, error) {
	return yaml.Marshal(lock)
}
