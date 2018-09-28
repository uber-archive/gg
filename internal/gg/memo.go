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
	"io"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/multierr"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Memo captures all the data necessary to manage dependencies, ensuring that
// expensive operations occur only once during a session.
type Memo struct {
	GitDir            string
	WorkDir           string
	GoPath            []string
	Repository        *git.Repository
	Fetched           StringSet                  // Package -> has been fetched
	Patterns          Patterns                   // Remote patterns from config
	Mirrors           map[int]struct{}           // Remote patterns that correspond to gitolite mirrors.
	Remotes           map[string]string          // Package -> Remote
	Refs              StringGraph                // Hash -> Refs for commit hashes only
	Versions          map[string][]plumbing.Hash // Root -> []Hash for commit hashes only
	FinishedVersions  map[string]Modules         // Root -> Modules
	Packages          map[string]Packages        // Hash:Package -> Packages
	Name              string                     // Name of own package
	OwnPackages       Packages                   // Imports and exports of working copy
	Excludes          StringSet                  // directory names to exclude from the working copy
	Recommended       map[string]Version         // config recommended versions for add missing workflow
	Finished          map[plumbing.Hash]ModuleResult
	Commits           map[plumbing.Hash]*object.Commit
	VendorCache       string
	PulledVendorCache bool
	Offline           bool

	// Metrics
	GitFetchCalls            int
	GitFetchDuration         time.Duration
	RemoteForPackageCalls    int
	RemoteForPackageDuration time.Duration
	GitCommitMemoHits        int
	GitResolveCommitCalls    int
	GitResolveCommitDuration time.Duration
	GitDigestRefsCalls       int
	GitDigestRefsDuration    time.Duration
	PrimedRemotes            int
	ReadOwnModulesCalls      int
	ReadOwnModulesDuration   time.Duration
	ReadGitPackagesCalls     int
	ReadGitPackagesDuration  time.Duration
}

// ModuleResult represents the module and error returned by the Finish method,
// to replay a prior result.
type ModuleResult struct {
	Module Module
	Error  error
}

// NewMemo returns a new memo for a dependency management session.
func NewMemo(gitDir, workDir string, goPath []string) (*Memo, error) {
	repo, err := Repository(gitDir)
	if err != nil {
		return nil, err
	}
	return &Memo{
		GitDir:           gitDir,
		WorkDir:          workDir,
		GoPath:           goPath,
		Repository:       repo,
		Fetched:          make(StringSet),
		Remotes:          make(map[string]string),
		Refs:             make(map[string]StringSet),
		Versions:         make(map[string][]plumbing.Hash),
		FinishedVersions: make(map[string]Modules),
		Packages:         make(map[string]Packages),
		OwnPackages:      NewPackages(),
		Recommended:      make(map[string]Version),
		Commits:          make(map[plumbing.Hash]*object.Commit),
		Finished:         make(map[plumbing.Hash]ModuleResult),
	}, nil
}

// ReadConfig reads the gg.toml in the working directory or any parent
// thereof.
func (memo *Memo) ReadConfig() error {
	// Read configuration in working copy.
	config, err := ReadOwnConfig(memo.WorkDir)
	if err != nil {
		return err
	}
	memo.VendorCache = config.Cache
	memo.Patterns = config.ReadPatterns()
	memo.Mirrors = config.ReadGitoliteMirrors()
	memo.Excludes = config.ReadExcludes()
	memo.Recommended = config.ReadRecommended()
	return err
}

// Repository gets or creates a bare git repository at the given path.
func Repository(path string) (*git.Repository, error) {
	var err error
	var repo *git.Repository
	repo, err = git.PlainInit(path, true)
	if err == git.ErrRepositoryAlreadyExists {
		repo, err = git.PlainOpen(path)
	}
	if err != nil {
		return nil, err
	}
	_ = ioutil.WriteFile(filepath.Join(path, ".gitignore"), []byte("*\n"), 0644)
	return repo, nil
}

// ReadOwnModules reads the glide.lock in the working copy and renders a slice
// of normalized, fetched, and cached Modules.
func (memo *Memo) ReadOwnModules(ctx context.Context, out ProgressWriter) (Modules, error) {
	start := time.Now()
	modules, err := ReadOwnModules()
	if err != nil {
		return nil, err
	}
	end := time.Now()
	memo.ReadOwnModulesDuration += end.Sub(start)
	memo.ReadOwnModulesCalls++

	// Prime the memo of package to remote mappings
	// from our own lockfile, as the authority for this cache.
	for _, module := range modules {
		select {
		case <-ctx.Done():
			return modules, ctx.Err()
		default:
		}
		if module.Remote != "" {
			memo.Remotes[module.Name] = module.Remote
			memo.PrimedRemotes++
		}
	}

	if err := memo.FinishModules(ctx, out, modules); err != nil {
		return nil, err
	}

	return modules, nil
}

// FinishModules normalizes, fetches, and caches all the given modules,
// providing a progress indicator.
func (memo *Memo) FinishModules(ctx context.Context, out ProgressWriter, modules Modules) error {
	start := time.Now()
	out.Start("Reading modules")
	for i := range modules {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := memo.FinishModule(ctx, out, &modules[i]); err != nil {
			fmt.Fprintf(out, "Failed to finish reading module %s: %s\n", modules[i].Summary(), err)
		}
		out.Progress("Reading modules", i+1, len(modules), start, time.Now())
	}
	out.Stop("Reading modules")
	return nil
}

// FinishModule fills in the blanks in a Module.
// Glide does not necessarily populate all of the fields we need, so this
// function attempts to fill in and validate.
// The VCS must be "git" or "", and we normalize to empty and therefore implied
// in the glide.lock.
// The Remote location for a Package is requires an HTTP request to the
// corresponding domain, which we obtain and write back to glide.lock.
// The vendor Root is a cache key implied by the Remote.
// Even if multiple packages use the same repository, we will fetch the
// repository once.
// We also make sure that we have fetched the corresponding version from the
// remote at least once in this process or cached in a prior.
// We ensure that we have read all of the remote's references into the memo
// for reverse lookups later.
// We then lookup the commit timestamp for the hash.
func (memo *Memo) FinishModule(ctx context.Context, out ProgressWriter, module *Module) error {
	if module.Hash == NoHash {
		return nil
	}
	if result, ok := memo.Finished[module.Hash]; ok {
		*module = result.Module
		return result.Error
	}
	err := memo.memoFinishModule(ctx, out, module)
	memo.Finished[module.Hash] = ModuleResult{Module: *module, Error: err}
	return err
}

func (memo *Memo) memoFinishModule(ctx context.Context, out ProgressWriter, module *Module) error {
	module.Finished = true

	// Stage 0: Package

	if err := memo.FinishRemote(ctx, out, module); err != nil {
		return err
	}

	// Stage 1: Remote, Root, ExactRemote

	var commit *object.Commit
	var err error
	if module.Hash == NoHash {
		// Skip this
	} else if commit, err = memo.Commit(ctx, out, module.Hash); err != nil {
		if err := memo.Fetch(ctx, out, module, FetchMaxAttempts); err != nil {
			module.FetchError = err
			fmt.Fprintf(out, "Unable to fetch module %s: %s\n", module.Summary(), err)
		} else if commit, err = memo.Commit(ctx, out, module.Hash); err != nil {
			module.Warnings = append(module.Warnings, fmt.Sprintf("Dependency %s no longer exists locally nor at %s: %s\n", module.Summary(), module.Remote, err))
		} else {
			module.Fetched = true
		}
	} else {
		module.Fetched = true
	}

	if module.Fetched {
		// Normalize the package hash to the commit hash.
		module.Hash = commit.Hash
	}

	// Stage 2: Fetched, Hash, Time, Glidelock, Deplock, Modules

	if module.Ref == "" {
		if err := memo.DigestRefs(ctx, out, *module); err != nil {
			module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot digest references: %s", err))
		}
		memo.finishModuleRef(module)
		if strings.HasPrefix(module.Ref, "tags/") {
			ref := strings.TrimPrefix(module.Ref, "tags/")
			module.Version = ParseVersion(ref)
		} else if strings.HasPrefix(module.Ref, "heads/") {
			ref := strings.TrimPrefix(module.Ref, "heads/")
			module.Version = ParseVersion(ref)
		}
	}

	if module.Time == (time.Time{}) {
		ts, err := memo.CommitTime(ctx, out, module.Hash)
		if err != nil {
			module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot infer commit timestamp: %s", err))
		}
		module.Time = ts
	}

	// Stage 3: Ref, Refs, Version, Time
	// Can now order versions, can decide whether to upgrade.

	// Read requirements from lockfile
	if commit != nil && !module.NoLock {
		if file, err := commit.File("glide.lock"); err == nil {
			module.Glidelock = file.Hash
			lock, err := memo.readGlideLock(module)
			if err != nil {
				module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot read glide.lock: %s", err))
			} else {
				lock.TestImports = nil
				modules, err := ModulesFromGlideLock(lock)
				if err != nil {
					module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot interpret glide.lock: %s", err))
				} else {
					module.Modules = modules
				}
			}
		} else if file, err := commit.File("Gopkg.lock"); err == nil {
			module.Deplock = file.Hash
			lock, err := memo.readDepLock(module)
			if err != nil {
				module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot read Gopkg.lock: %s", err))
			} else {
				modules, err := ModulesFromDepLock(lock)
				if err != nil {
					module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot interpret Gopkg.lock: %s", err))
				} else {
					module.Modules = modules
				}
			}
		} else {
			module.NoLock = true
		}
	}

	// Stage 4: Glidelock, Deplock, Modules
	// Can now solve dependency graph.
	return nil
}

// FinishPackages populates the Packages property of each module by analyzing
// the imports of the files in the git repository.
func (memo *Memo) FinishPackages(ctx context.Context, out ProgressWriter, modules Modules) error {
	start := time.Now()
	out.Start("Reading packages")
	for i := range modules {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := memo.finishPackages(ctx, out, &modules[i]); err != nil {
			return err
		}
		out.Progress("Reading packages", i+1, len(modules), start, time.Now())
	}
	out.Stop("Reading packages")
	return nil
}

func (memo *Memo) finishPackages(ctx context.Context, out ProgressWriter, module *Module) error {
	if !module.Packages.Defined() {
		if err := memo.digestGitPackages(ctx, out, module); err != nil {
			module.Warnings = append(module.Warnings, fmt.Sprintf("Cannot read packages: %s", err))
		}
	}
	// Stage 5: Packages
	return nil
}

// digestGitPackages reads and memoizes the imports of every package in the given
// module from the Git repository.
// The result varies by the root package path (inferred from the dependency
// Module), which determines what packages are exported.
// A single repository might be seen from multiple import package paths,
// especially over different versions.
// For example, some versions of github.com/uber-go/thriftrw predate the
// creation of the go.uber.org/thriftrw alias.
func (memo *Memo) digestGitPackages(ctx context.Context, out ProgressWriter, module *Module) error {
	repo := memo.Repository
	key := module.Hash.String() + ":" + module.Name

	if module.Packages.Defined() {
		// Prime cache from lockfile.
		memo.Packages[key] = module.Packages
		return nil
	}

	commit, err := memo.Commit(ctx, out, module.Hash)
	if err != nil {
		return err
	}

	tree, err := repo.TreeObject(commit.TreeHash)
	if err != nil {
		return fmt.Errorf("error attempting to get a Git tree to analyze Go packages in %s from commit %s: %s", module.Summary(), module.Hash, err)
	}

	start := time.Now()
	if err := ReadGitPackages(out, repo, tree, module); err != nil {
		return err
	}
	end := time.Now()
	memo.ReadGitPackagesDuration += end.Sub(start)
	memo.ReadGitPackagesCalls++

	memo.Packages[key] = module.Packages
	return nil
}

// FinishRemote populates the Remote and Root fields, based on the Package
// field, based on the cached mapping or an HTTP request.
func (memo *Memo) FinishRemote(ctx context.Context, out ProgressWriter, module *Module) error {
	if err := memo.finishRemote(ctx, out, module); err != nil {
		return err
	}
	module.Root = RootForRemote(module.Remote)
	return nil
}

// finishRemote returns the remote URL for the given go package name,
// preferring the cached remote URL even over the remote URL supplied
// by a lockfile, then falling back to looking up the remote URL by
// sending an HTTP request to the package's domain.
// This method may reveal that the module's package name is a prefix of the
// requested package name, and will return the truncated package name.
func (memo *Memo) finishRemote(ctx context.Context, out ProgressWriter, module *Module) error {
	// Prefer remote from cache (overrides).
	if remote, ok := memo.Remotes[module.Name]; ok {
		module.Remote = remote
		return nil
	}

	// Truncate package names that contain a .git to indicate the repository
	// location.
	if index := strings.Index(module.Name, ".git/"); index >= 0 {
		module.Name = module.Name[:index+4]
		module.ExactRemote = true
	}

	// Consult the configuration for pattern matches.
	// The configuration has more authority than a lockfile.
	if name, remote, rule := memo.Patterns.Replace(module.Name); rule >= 0 {
		memo.Remotes[name] = remote
		if _, ok := memo.Mirrors[rule]; ok {
			module.GitoliteMirror = true
		}
		module.Name = name
		module.Remote = remote
		module.ExactRemote = true
		return nil
	}

	// Use the value specified in the lockfile otherwise.
	if module.Remote != "" {
		return nil
	}

	// Best guess in offline mode.
	if memo.Offline {
		module.Remote = "https://" + module.Name
		memo.Remotes[module.Name] = module.Remote
		module.Warnings = append(module.Warnings, fmt.Sprintf("The remote location %s may be corrupt.  This module was obtained in offline mode so no HTTP request was sent to validate the assumed location of its remote repository.", module.Name))
		return nil
	}

	// Consult the web, the source of truth, as a last resort.
	// ShowModule(out, NewState(), *module)
	status := fmt.Sprintf("Looking up remote for package %s", module.Name)
	out.Start(status)
	start := time.Now()
	rem, name := RemoteForPackage(module.Name)
	end := time.Now()
	module.Remote = rem
	module.Name = name
	out.Stop(status)

	memo.Remotes[module.Name] = module.Remote
	// fmt.Fprintf(out, "Remote for package %s is %s.\n", module.Name, module.Remote)
	memo.RemoteForPackageDuration += end.Sub(start)
	memo.RemoteForPackageCalls++
	return nil
}

// finishModuleRef is a utility of memoFinishModule that looks up all
// references that refer to a module's commit hash.
// The references for the module must be cached first, with DigestRefs.
func (memo *Memo) finishModuleRef(module *Module) {
	var bestRef string
	var bestVersion Version

	prefix := "refs/vendor/" + module.Root + "/"

	for _, ref := range memo.Refs[module.Hash.String()].Keys() {
		if strings.HasPrefix(ref, prefix) {
			ref = strings.TrimPrefix(ref, prefix)
			if strings.HasPrefix(ref, "tags/") {
				verRef := strings.TrimPrefix(ref, "tags/")
				version := ParseVersion(verRef)
				if bestVersion.Before(version) {
					bestVersion = version
					bestRef = ref
				}
			} else if bestVersion == NoVersion && (ref == "heads/master" || ref > bestRef) {
				// If none of the references are versions, choose the last
				// reference in lexical order.  This is arbitrary but consistent.
				// We chose last because empty string would always be first.
				bestRef = strings.TrimPrefix(ref, prefix)
			}
		}
	}

	// TODO trace history of known branches (or just heads/master) in search of
	// this commit hash.
	// if bestRef == "" && bestVersion == NoVersion {
	// 	commit, err := memo.Repository.Commit(out, plumbing.NewHash(module.Hash))
	// }

	module.Ref = bestRef
	module.Version = bestVersion
}

// Commit looks up the corresponding commit object for the hash of a commit or
// tag.
// Follows tags to their corresponding commit.
// Memoizes the commit for every hash leading to that commit so every
// subsequent lookup is fast.
// We use this utility to normalize all lockfile hashes to the target commit
// hash.
// We normalize hashes to ensure that a reverse lookup from hash to reference
// works regardless of whether a legacy lockfile has a level of indirection
// through tags.
func (memo *Memo) Commit(ctx context.Context, out ProgressWriter, hash plumbing.Hash) (*object.Commit, error) {
	if memo.VendorCache != "" && !memo.PulledVendorCache {
		err := GitPullVendorCache(out, memo.GitDir, memo.VendorCache)
		if err != nil {
			fmt.Fprintf(out, "Unable to fetch vendor references cache: %s\n", err)
		}
		memo.PulledVendorCache = true
	}

	if commit, ok := memo.Commits[hash]; ok {
		memo.GitCommitMemoHits++
		return commit, nil
	}

	repo := memo.Repository
	start := time.Now()

	var tag *object.Tag
	var hashes []plumbing.Hash
	var commit *object.Commit
	var err error

	// Follow tags to a commit
FollowTags:
	for {
		hashes = append(hashes, hash)
		tag, err = repo.TagObject(hash)
		if err != nil {
			break
		}
		switch tag.TargetType {
		case plumbing.TagObject:
			hash = tag.Target
			continue FollowTags
		case plumbing.CommitObject:
			hash = tag.Target
			break FollowTags
		default:
			// TODO return a distinguished error type
			return nil, fmt.Errorf("error digesting refs: hash does not refer to a commit %v: %s", tag, err)
		}
	}

	commit, err = repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("error getting commit for %s: %v", hash, err)
	}
	for _, hash := range hashes {
		memo.Commits[hash] = commit
	}

	if err == nil {
		end := time.Now()
		memo.GitResolveCommitCalls++
		memo.GitResolveCommitDuration += end.Sub(start)
	}

	return commit, err
}

// CommitTime returns the commit time for a given commit or tag hash.
func (memo *Memo) CommitTime(ctx context.Context, out ProgressWriter, hash plumbing.Hash) (time.Time, error) {
	commit, err := memo.Commit(ctx, out, hash)
	if err != nil {
		return time.Time{}, err
	}
	return commit.Committer.When, nil
}

// Fetch ensures that the given module has been synced with its remote URL once
// in the scope of this memo.
// Fetch does nothing in offline mode.
// Fetch does not implicitly digest the new references.
// Fetch does not guarantee any references will exist for the module.
func (memo *Memo) Fetch(ctx context.Context, out ProgressWriter, module *Module, maxAttempts int) error {
	if memo.Offline {
		return nil
	}
	if module.Name == "" {
		fmt.Fprintf(out, "Cannot fetch %s because the package name is blank.\n", module.Summary())
		return nil
	}
	if module.Remote == "" {
		fmt.Fprintf(out, "Cannot fetch %s because the remote is empty.\n", module.Summary())
		return nil
	}

	if err := memo.FinishRemote(ctx, out, module); err != nil {
		return err
	}

	if _, ok := memo.Fetched[module.Remote]; ok {
		return nil
	}
	memo.Fetched[module.Remote] = struct{}{}

	fetching := fmt.Sprintf("Fetching %s", module.Remote)
	out.Start(fetching)
	defer out.Stop(fetching)

	// Attempt to create a gitolite mirror if neccessary.
	if module.GitoliteMirror && !module.GitoliteMirrorCreated {
		if err := GitoliteCreateMirror(out, module.Remote); err != nil {
			fmt.Fprintf(out, "Error attempting to create mirror at %s: %s\n", module.Remote, err)
		}
		module.GitoliteMirrorCreated = true
	}

	// Fetch retry loop with exponential back-off, full jitter.
	var attempts uint
	for {
		start := time.Now()
		err := GitFetchRootRemote(out, memo.GitDir, module.Root, module.Remote)
		if err != nil {
			attempts++
			if attempts > uint(maxAttempts) {
				return err
			}

			waitns := (1 << attempts) * FetchFirstAttemptWait.Nanoseconds()
			if waitns > FetchMaxAttemptWait.Nanoseconds() {
				waitns = FetchMaxAttemptWait.Nanoseconds()
			}
			wait := time.Duration(rand.Int63n(waitns + 1))
			fmt.Fprintf(out, "Error fetching %s. Attempt %d. Retrying in %v.\n", module.Remote, attempts, wait)
			time.Sleep(wait)

			// Lose faith in the cached remote for this package, as once
			// happened for Apache Thrift when they moved their repository.
			if attempts%2 == 0 {
				// Invalidate cache
				delete(memo.Remotes, module.Name)
				module.Remote = ""
				// Retrieve
				if err := memo.FinishRemote(ctx, out, module); err != nil {
					return err
				}
			}

			continue
		}
		end := time.Now()
		memo.GitFetchDuration += end.Sub(start)
		memo.GitFetchCalls++
		break
	}

	// Reload the repository because shelling out to git behind it invalidates its cache.
	repo, err := Repository(memo.GitDir)
	if err != nil {
		return err
	}
	memo.Repository = repo

	return nil
}

// DigestRefs updates the memo of references in a package root by reading the
// vendored git cache.
func (memo *Memo) DigestRefs(ctx context.Context, out ProgressWriter, module Module) error {
	start := time.Now()

	prefix := "refs/vendor/" + module.Root + "/"
	var versions []plumbing.Hash

	refsIter, err := memo.Repository.References()
	if err != nil {
		return fmt.Errorf("error attempting to digest references from the git repository for package root %s: %s", module.Root, err)
	}

	for {
		ref, err := refsIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error digesting references from the git repository for package root %s: %s", module.Root, err)
		}
		if ref.Type() == plumbing.InvalidReference {
			fmt.Fprintf(out, "warning: Unable to digest reference: invalid reference %s\n", ref)
		} else if ref.Type() == plumbing.HashReference {
			name := ref.Name().String()
			if !strings.HasPrefix(name, prefix) || strings.Contains(name, "/phabricator/") {
				continue
			}
			// Also, follow references all the way to a commit hash for
			// normalization.
			commit, err := memo.Commit(ctx, out, ref.Hash())
			if err != nil {
				return fmt.Errorf("error digesting references from git repository for package root %s, following git commit for hash %s: %s", module.Root, ref.Hash(), err)
			}
			memo.Refs.Add(commit.Hash.String(), name)
			versions = append(versions, commit.Hash)
		}
		// Ignores symbolic references
	}

	memo.Versions[module.Root] = versions

	end := time.Now()
	memo.GitDigestRefsDuration += end.Sub(start)
	memo.GitDigestRefsCalls++
	return nil
}

// ReadVersions returns a list of versions of the given module with the same
// package, root, and test flag.
// The given module must first be fetched and its references digested.
// Each module is normalized, read, and cached, so you can depend on
// all of each version's fields to be populated.
// The returned modules are in total order based on their version, timestamp,
// and hash.
func (memo *Memo) ReadVersions(ctx context.Context, out ProgressWriter, module Module) (Modules, error) {
	status := fmt.Sprintf("Finding versions of %s", module.Name)
	out.Start(status)
	defer out.Stop(status)

	if modules, ok := memo.FinishedVersions[module.Root]; ok && modules != nil {
		return modules, nil
	}

	modules := make(Modules, 0, 1)
	if err := memo.Fetch(ctx, out, &module, FetchMaxAttempts); err != nil {
		return nil, err
	}
	if err := memo.DigestRefs(ctx, out, module); err != nil {
		return nil, err
	}
	for _, hash := range memo.Versions[module.Root] {
		module := Module{
			Hash: hash,
			Name: module.Name,
			Root: module.Root,
			Test: module.Test,
		}
		modules = append(modules, module)
	}
	if err := memo.FinishModules(ctx, out, modules); err != nil {
		return nil, err
	}

	memo.FinishedVersions[module.Root] = modules

	sort.Sort(modules)
	return modules, nil
}

func (memo *Memo) readGlideLock(module *Module) (*GlideLock, error) {
	repo := memo.Repository

	blob, err := repo.BlobObject(module.Glidelock)
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}
	defer func() {
		err = multierr.Append(err, reader.Close())
	}()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}

	glidelock, err := ReadGlideLock(bytes)

	return glidelock, err
}

func (memo *Memo) readDepLock(module *Module) (*DepLock, error) {
	repo := memo.Repository

	blob, err := repo.BlobObject(module.Deplock)
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}

	reader, err := blob.Reader()
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}
	defer func() {
		err = multierr.Append(err, reader.Close())
	}()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error readling lockfile for commit %s: %s", module.Hash, err)
	}

	deplock, err := ReadDepLock(bytes)

	return deplock, err
}

// ReadOwnPackages returns the working copy's memoized package name and
// packages.  The returned packages are not mutable.
func (memo *Memo) ReadOwnPackages(ctx context.Context, out ProgressWriter) (string, Packages, error) {
	if memo.OwnPackages.Defined() {
		return memo.Name, memo.OwnPackages.Clone(), nil
	}
	out.Start("Reading packages in working copy")
	name, packages, err := ReadOwnPackages(out, memo.WorkDir, memo.GoPath, memo.Excludes)
	out.Stop("Reading packages in working copy")
	memo.Name = name
	memo.OwnPackages = packages
	return name, packages.Clone(), err
}

// FindModule finds a module that satisfies the given spec and test requirement.
// The spec is at least a package name, optionally followed by @version, @ref, or @hash.
// Without a specific version, FindModule will take recommended packages from gg.toml.
func (memo *Memo) FindModule(ctx context.Context, out ProgressWriter, spec string, test bool) (Module, error) {
	parts := strings.SplitN(spec, "@", 2)
	name := parts[0]
	var ref string
	if len(parts) == 2 {
		ref = parts[1]
	}

	min, max := ParseHashPrefix(ref)

	module := Module{
		Name:    name,
		Test:    test,
		Version: ParseVersion(ref),
	}

	if err := memo.FinishRemote(ctx, out, &module); err != nil {
		return module, err
	}

	if err := memo.Fetch(ctx, out, &module, FetchMaxAttempts); err != nil {
		fmt.Fprintf(out, "warning: Failed to fetch for %s: %s\n", module.Summary(), err)
	}

	modules, err := memo.ReadVersions(ctx, out, module)
	if err != nil {
		return module, err
	}

	if len(modules) == 0 {
		return module, fmt.Errorf("no versions of %s found online or in cache", module.Name)
	}

	var ok bool
	if module.Version != NoVersion {
		found, ok := modules.FindVersion(module.Version)
		if !ok {
			return module, fmt.Errorf("cannot find a version of %s that satisfies version %s", module.Name, module.Version)
		}
		module = found
	} else if min != NoHash {
		found, ok := modules.FindHash(min, max)
		if !ok {
			return module, fmt.Errorf("cannot find a version of %s between hashes [%s, %s]", module.Name, min, max)
		}
		module = found
	} else if ref == "" {
		if recommended, ok := memo.Recommended[module.Name]; ok {
			if module, ok := modules.FindVersion(recommended); ok {
				return module, nil
			}
		}

		versionedModules := modules.FilterNumberedVersions()
		if len(versionedModules) == 0 {
			module, ok = modules.FindReference("heads/master")
			if !ok {
				return module, fmt.Errorf("unable to find a version tag or master branch")
			}
		} else {
			module = versionedModules[len(versionedModules)-1]
		}
	} else {
		module, ok = modules.FindReference(ref)
		if !ok {
			return module, fmt.Errorf("cannot find specified reference")
		}
	}

	return module, nil
}

// RootForRemote computes a cache key for a remote repository, for the vendor
// reference in the git cache, by stripping the protocol and normalizing the
// extension.
func RootForRemote(remote string) string {
	remote = strings.TrimSuffix(remote, ".git")

	if strings.Contains(remote, "@") {
		parts := strings.SplitN(remote, "@", 2)
		remote = parts[1]
	}

	switch {
	case strings.HasPrefix(remote, "https://"):
		remote = strings.TrimPrefix(remote, "https://")
	case strings.HasPrefix(remote, "git://"):
		remote = strings.TrimPrefix(remote, "git://")
	}

	return strings.Replace(remote, ":", "/", -1)
}
