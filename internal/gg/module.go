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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// NoModule indicates that a module could not be found.
var NoModule Module

// Module is the model of a Go dependency in a Git repository.
type Module struct {
	// Name is the Go package name that addresses the root of the repository.
	Name string

	// Hash is the commit hash in the module's git repository that refers to
	// this revision.
	// In a glide.lock written by the glide tool, this might be the hash
	// of a tag, but GG will normalize this to the hash of the commit.
	Hash plumbing.Hash

	// Version is the major, minor, and patch version number triple that
	// denotes this module exactly.
	// The version indicates an order, but does not capture the semantics or
	// ordering of semantic versions with suffixes like 1.0.0-rc1.
	// This tool treats any version number that is more specific than three
	// bare numbers as versionless and compares them just using timestamps, so
	// release candidates are ordered by time and inelligible for upgrade using
	// semver.
	// The version corresponds to a git tag that resolves to this module's
	// commit hash.
	Version Version

	// Remote is the URL of the repository that contains this module.
	// Go uses an HTTP request to the domain name in the first path component
	// in the package name to look up the remote URL, or failing that,
	// infer the remote URL from the package name directly.
	// GG caches the remote location to be more resilient in the face of
	// network weather and consistent in the face of network inconsistency.
	Remote string

	// ExactRemote indicates that the remote string came from a config pattern
	// match and that the AddMissing workflow should not search upward.
	ExactRemote bool

	// Root is a cache key for the remote URL, suitable for use in git
	// references in the .gg bare repository.
	Root string

	// Time is the commit timestamp for this module, which we can use to
	// construct a total order from multiple revisions of the same module
	// regardless of whether we can infer a version number.
	Time time.Time

	// Ref is the best reference that resolves to this module's commit hash.
	// If any of the references to this commit parses as an exact three-digit
	// version number, the highest version reference is best.
	// Otherwise, if the reference is "heads/master", it is the best.
	// Otherwise, the last reference in lexicographic order is best.
	Ref string

	// Refs are all of the references that refer to this commit as of the last
	// time this module's repository was successfully fetched.
	Refs []string

	// Test indicates that this module is only needed for tests.
	Test bool

	// Modules are the shallow constraints of this module, as expressed in its
	// lockfile.
	Modules Modules

	// NoLock indicates that neither glide.lock nor Gopkg.in were found in the
	// module.  If Modules are empty and NoLock is true, we can skip looking
	// for requirements.
	NoLock bool

	// Packages is this module's package import graph.
	Packages Packages

	// Warnings are any warnings encountered while attempting to read this
	// module's metadata, particularly whether the package import comments
	// match up with the expected values based on this module's package name.
	Warnings []string

	// Glidelock is the git hash of the module's glide.lock, or NoHash if
	// absent.
	Glidelock plumbing.Hash

	// Deplock is the git hash of this module's Gopkg.lock, or NoHash if
	// absent.
	Deplock plumbing.Hash

	// Changelog is the git hash of this module's CHANGELOG.md, or NoHash if
	// absent.
	Changelog plumbing.Hash

	// GitoliteMirror indicates that the remote is a Gitolite mirror.
	// The mirror may need to be created before the module can be fetched.
	GitoliteMirror bool

	// GitoliteMirrorCreated indicates that the mirror was at some point in the
	// past created so we can bypass this step before fetching.
	GitoliteMirrorCreated bool

	// Fetched indicates that this module has been fetched during this session.
	Fetched bool

	// FetchError indicates the error produced when this module was fetched in
	// this session.
	FetchError error

	// Finished indicates that this module has already been fetched in this
	// session.
	Finished bool
}

// Summary produces a unique description of the module, suitable for printing
// inline.
func (module Module) Summary() string {
	test := ""
	if module.Test {
		test = "#test"
	}
	if module.Version != NoVersion {
		return module.Name + "@" + module.Version.String() + test
	}
	if module.Ref != "" {
		return module.Name + "@" + module.Ref + test
	}
	return module.Name + "@" + module.Hash.String() + test
}

// String produces a string representation of the module, suitable for use in a
// report with aligned columns.
func (module Module) String() string {
	hash := "        "
	if module.Hash == NoHash {
		hash = "########"
	} else {
		hash = fmt.Sprintf("%-8s\n", module.Hash)[:8]
	}

	date := "          "
	if module.Time != (time.Time{}) {
		y, m, d := module.Time.Date()
		date = fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	}

	var version string
	if module.Version != NoVersion {
		version = fmt.Sprintf("%4d.%3d.%3d", module.Version[0], module.Version[1], module.Version[2])
	} else if strings.HasPrefix(module.Ref, "tags/") {
		version = fmt.Sprintf("%12s", strings.TrimPrefix(module.Ref, "tags/"))
	} else if strings.HasPrefix(module.Ref, "heads/") {
		version = fmt.Sprintf("%12s", strings.TrimPrefix(module.Ref, "heads/"))
	} else {
		version = "           ?"
	}
	if len(version) > 12 {
		version = version[:12]
	}

	test := " "
	if module.Test {
		test = "T"
	}

	cl := " "
	if module.Changelog != NoHash {
		cl = "C"
	}

	lock := " "
	if module.Glidelock != NoHash {
		lock = "G"
	} else if module.Deplock != NoHash {
		lock = "D"
	}

	refs := strings.Join(module.Refs, " ")
	if len(refs) > 0 {
		refs = " (" + refs + ")"
	}

	name := module.Name
	if name == "" {
		name = "-"
	}

	warnings := ""
	if len(module.Warnings) == 1 {
		warnings = " " + yellow + "(warning)" + clear
	} else if len(module.Warnings) > 1 {
		warnings = fmt.Sprintf(" "+yellow+"(%d warnings)"+clear, len(module.Warnings))
	}

	return fmt.Sprintf("%s %s %s %s%s%s %s%s%s", hash, date, version, test, lock, cl, name, refs, warnings)
}

// Equal returns whether the given modules have the same hash and test fields.
func (module Module) Equal(other Module) bool {
	return module.Name == other.Name && module.Hash == other.Hash && module.Test == other.Test
}

// Before returns whether the module is before another module, based on name,
// time, version, and commit hash, in order of priority.
// This order is suitable and sufficient for deterministically arriving at the
// most recent version of every package in a closure of modules over their
// transitive dependencies on exact hashes (the solver).
// The order is also suitable for sorting modules in a solution for display to
// the user.
// This order is not sufficient for inferring whether a module can be upgraded
// to another.
// In practice, all versions will be orderable by time without taking the
// version or hash into account, so the version is taken into account only for
// tests that do not specify time stamps.
// Ordering by hash if timestamps are exactly the same is a measure of paranoia
// to ensure a deterministic albeit arbitrary order for even the oddest cases.
// This order is also suitable for displaying modules (full solutions with many
// names or versions of the same name) in ascending chronological order.
func (module Module) Before(other Module) bool {
	if module.Name != other.Name {
		return module.Name < other.Name
	}
	if module.Version != other.Version {
		return module.Version.Before(other.Version)
	}
	if !module.Time.Equal(other.Time) {
		return module.Time.Before(other.Time)
	}
	return HashBefore(module.Hash, other.Hash)
}

// Better returns whether the module is better suited than the other if
// we needed to add one version of the module and knew nothing else about the
// code that depends on the module.
// We will favor a module with a version, the highest version, over all others.
// Otherwise, we will favor the "master" branch over all others.
// This is for the "add missing modules" workflow, and specifically avoiding
// development branches.
func (module Module) Better(other Module) bool {
	if module.Ref == "heads/master" && other.Version == NoVersion {
		return true
	}
	if module.Version == NoVersion {
		return false
	}
	if module.Version == other.Version {
		return module.Time.After(other.Time)
	}
	return other.Version.Before(module.Version)
}

// CanUpgradeTo returns a heuristic for whether this module can be upgraded to
// another module.
// If either package has a semantic version, they are upgradable based on the
// semver rules.
// If a package has no reference, it can always upgrade to "master", to heal
// glide.locks written by glide.
// Otherwise, a module can only upgrade to a newer revision with the same
// reference, based on their commit timestamps.
func (module Module) CanUpgradeTo(other Module) bool {
	// Never travel backward in time.
	// Staying in the same time is useful only because many tests use the zero
	// time for all versions.
	if module.Time.After(other.Time) {
		return false
	}
	// If either has a semantic version, abide by semantic versions.
	if module.Version != other.Version {
		return module.Version.CanUpgradeTo(other.Version)
	}
	// Otherwise, heal missing references by upgrading to master.
	if module.Ref == "" {
		return other.Ref == "heads/master"
	}
	// Otherwise, ugprade only if they have the same branch.
	return module.Time.Before(other.Time) && module.Ref == other.Ref
}

// Modules is a slice of modules with an order based on name, timestamp,
// version, and hash.
// A slice of modules can represent the dependencies of a module, a whole
// module solution, or even just revisions with the same package name.
type Modules []Module

// Packages returns the union of all packages.
func (modules Modules) Packages() Packages {
	packages := NewPackages()
	for _, module := range modules {
		packages.Include(module.Packages)
	}
	return packages
}

// Equal returns whether the slices are equivalent.
func (modules Modules) Equal(others Modules) bool {
	if len(modules) != len(others) {
		return false
	}
	for i, module := range modules {
		if !module.Equal(others[i]) {
			return false
		}
	}
	return true
}

// Len returns the length of the slice of modules.
func (modules Modules) Len() int {
	return len(modules)
}

// Less returns whether a pair of modules are in descending order, so they can
// be swapped to sort.
func (modules Modules) Less(i, j int) bool {
	return modules[i].Before(modules[j])
}

// Swap swaps modules in the slice, so they can sort.
func (modules Modules) Swap(i, j int) {
	modules[i], modules[j] = modules[j], modules[i]
}

// String returns a string representation of the slice of modules, using their
// inline summaries.
func (modules Modules) String() string {
	var strs []string
	sort.Sort(modules)
	for _, module := range modules {
		strs = append(strs, module.Summary())
	}
	return "[" + strings.Join(strs, " ") + "]"
}

// Index returns a map of these modules, indexed by package name, and also a
// string set of their package names.
func (modules Modules) Index() map[string]Module {
	named := make(map[string]Module)
	for _, module := range modules {
		named[module.Name] = module
	}
	return named
}

// FindReference returns the module that has the shortest reference with the
// given suffix and whether one such was found, suitable for finding a module
// in a slice of module versions.
func (modules Modules) FindReference(ref string) (Module, bool) {
	var found Module
	var ok bool
	for _, module := range modules {
		if (module.Ref == ref || strings.HasSuffix(module.Ref, "/"+ref)) && (!ok || len(ref) < len(found.Ref)) {
			found = module
			ok = true
		}
	}
	return found, ok
}

// FindHash returns the first module in a range of hashes, and whether one such
// was found.
func (modules Modules) FindHash(min, max plumbing.Hash) (Module, bool) {
	for _, module := range modules {
		if HashBetween(min, module.Hash, max) {
			return module, true
		}
	}
	return Module{}, false
}

// FindVersion returns the module with the highest version that satisfies the
// given version's implied semantic version range, and whether one such was
// found.
func (modules Modules) FindVersion(version Version) (found Module, ok bool) {
	for _, module := range modules {
		if version == module.Version || version.CanUpgradeTo(module.Version) {
			version = module.Version
			found = module
			ok = true
		}
	}
	return
}

// FindBestVersion returns the module in a slice of module versions that would
// be the best guess to fill a missing module.
// The best version is the highest semantic version or the master version if
// there are no versioned modules.
func (modules Modules) FindBestVersion() (best Module, found bool) {
	for _, module := range modules {
		if module.Better(best) {
			best = module
			found = true
		}
	}
	return
}

// FindBestSemver finds the best module that qualifies a semver constraint as
// defined by Glide.
func (modules Modules) FindBestSemver(constraint *semver.Constraints) (Module, bool) {
	var best Module
	var found bool
	for _, module := range modules {
		if module.Version != NoVersion {
			version, err := semver.NewVersion(fmt.Sprintf("%d.%d.%d",
				module.Version[0],
				module.Version[1],
				module.Version[2],
			))
			if err == nil && constraint.Check(version) && module.Better(best) {
				best = module
				found = true
			}
		}
	}
	return best, found
}

// FilterNumberedVersions returns only modules that have an associated version
// number.
func (modules Modules) FilterNumberedVersions() Modules {
	var filtered Modules
	for _, module := range modules {
		if module.Version != NoVersion {
			filtered = append(filtered, module)
		}
	}
	return filtered
}

// Names returns the names of all the modules.
func (modules Modules) Names() StringSet {
	names := make(StringSet)
	for _, module := range modules {
		names.Add(module.Name)
	}
	return names
}

// FilterDependencies returns only the modules that are dependencies of the
// given module.
func (modules Modules) FilterDependencies(got Module) (dependencies []Dependency) {
	for _, module := range modules {
		for _, want := range module.Modules {
			if want.Name == got.Name {
				dependencies = append(dependencies, Dependency{
					module,
					want,
					got,
				})
			}
		}
	}
	return dependencies
}

// Conflicts returns whether the given module has conflicts with any other
// module in the solution.
func (modules Modules) Conflicts(got Module) bool {
	for _, dep := range modules.FilterDependencies(got) {
		if dep.Want.Conflicts(dep.Got) {
			return true
		}
	}
	return false
}

// Conflicts returns whether the module we want conflicts with the module we
// got instead.
func (module Module) Conflicts(other Module) bool {
	return module.Hash != other.Hash && module.Version != other.Version && !module.CanUpgradeTo(other)
}

// Dependency is a triple of a module, the version of a module it wants, the
// version of that module that it actually got.
type Dependency struct{ Module, Want, Got Module }
