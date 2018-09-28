# gg

GG is an interactive dependency manager and debugger for Go.

* GG uses Git to cache your project's transitive dependencies and infers their
  version numbers and publication dates from tags and timestamps.
  So, all dependencies must be in Git repositories.
* GG uses existing Glide and Dep manifests and lockfiles to collect dependencies.
* GG then uses the Go 1.11 deterministic solver to quickly settle the dependency graph.
* Because the Go 1.11 dependency solver does not guarantee that the solution
  consists of mutally compatible versions, GG provides an interactive debugger.
* GG can reveal incompatible constraints between modules in a lockfile.
* GG can reveal missing and extra modules and packages by reading the import
  graphs of your dependencies where they rest in Git.
* GG can list and deeply inspect versions of modules.
* GG can help you both automatically upgrade all modules within the same
  semantic version range or help you manually add or upgrade specific versions.
* GG can emulate all of Glide's workflows, except without race conditions.
* GG can emulate all of Dep's workflows, except without downgrades.

## Install

    go get github.com/uber/gg

## Usage

```
A go git dependency management suite.
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
```

## Example

This is a dependency report generated for gg itself:

```console
$ gg show-solution
Locked modules: (26)
* c7af1294 2018-04-10    1.  4.  2     github.com/Masterminds/semver
* 3fe1cac1 2018-09-29       master     github.com/RyanCarrier/dijkstra
* 62c6fe61 2016-07-26    1.  4.  0     github.com/chzyer/readline
* 346938d6 2016-10-29    1.  1.  0 T   github.com/davecgh/go-spew
* 1615341f 2018-09-23    1. 12.  0 T   github.com/emirpasic/gods
* 6f453133 2015-01-27       master     github.com/google/shlex
* d14ea06f 2015-07-10       master T   github.com/jbenet/go-context
* 81db2a75 2018-08-30    0.  5.  0 T   github.com/kevinburke/ssh_config
* ae18d6b8 2018-08-23    1.  0.  0 T   github.com/mitchellh/go-homedir
* c37440a7 2017-02-27    0.  2.  0 T   github.com/pelletier/go-buffruneio
* 792786c7 2016-01-10    1.  0.  0 T   github.com/pmezard/go-difflib
* 1744e297 2017-11-10    1.  0.  0 T   github.com/sergi/go-diff
* f1873551 2016-10-26    1.  3.  0 T   github.com/src-d/gcfg
* facf9a85 2018-01-06    0.  1.  0 TD  github.com/stretchr/objx
* f35b8ab0 2018-06-09    1.  2.  2 TD  github.com/stretchr/testify
* 640f0ab5 2018-07-03    0.  2.  0 T   github.com/xanzy/ssh-agent
* 3b8db5e9 2016-12-15    1.  1.  0  G  go.uber.org/atomic
* 3c493748 2017-06-30    1.  1.  0  G  go.uber.org/multierr
* 5295e836 2018-09-27       master T   golang.org/x/crypto
* 4dfa2610 2018-09-26       master T   golang.org/x/net
* e4b3c5e9 2018-09-28       master T   golang.org/x/sys
* f21a4dfb 2017-12-14    0.  3.  0 T   golang.org/x/text
* 98262648 2018-09-19    4.  3.  0 T   gopkg.in/src-d/go-billy.v4
* d3cec13a 2018-09-06    4.  7.  0 T   gopkg.in/src-d/go-git.v4
* ec4a0fea 2017-11-15    0.  1.  2 T   gopkg.in/warnings.v0
* 5420a8b6 2018-03-28    2.  2.  1 T   gopkg.in/yaml.v2
  T: needed for tests only.
  G: has a glide.lock.
  D: has a Gopkg.lock (dep)
  C: has a CHANGELOG.md.
```

## Rationale

Glide and dep both have correctness, performance, and usability problems.
The vgo proposal and Go 1.11 modules address the correctness and performance
problems, account for the coëxistence of major versions, provide a better
caching solution, isolate projects, and integrate dependency management in the
build toolchain.
To use Go modules, all of a project's transitive dependencies must on-board to
Go modules.
We can fast-forward to Go module resolution semantics by combining lock files
and Git metadata to obtain all of the information that would be written in a Go
module file.

Because in practice most projects' transitive dependencies are in Git, we can
make some simplifying assumptions about the scope of a dependency manager's
responsibilities and delegate a lot of heavy lifting to Git.
This assumptoin makes it possible for a small and effective tool that addresses
some problems of the past, realizes some solutions of the future, and then goes
on to project what we can build on Go modules to solve some unresolved
problems like debugging a dependency graph.

- [x] Implement back-tracking "max of mins" constraint solver based on a total
  order of modules, inferring semantic versions where possible by reverse
  lookup from hash to tag name, and then fall back to comparing by timestamp.
- [x] Automatically settle existing lockfiles on the corresponding deterministic
  solution.
- [x] Convert any lock or manifest file into Glide or Dep format.
- [ ] Generate Go module files.
- [x] Reliable caching by automatically pushing and pulling transitive
  dependencies in a Git `refs/vendor` namespace.
- [x] Harness git for fast checkouts of the vendor directory.
- [ ] Faster checkouts my mapping lockfile hashes to vendor tree hashes.
- [x] Fully automated *and* manual workflows for informed dependency upgrades,
  to empower the consumer to test an upgrade before comitting and explore
  alternatives.
- [x] Address Glide's inability to resolve conflicts between imports and test
  imports.
  Deterministically promote test dependencies to binary dependencies as if they
  were merely another version.
- [x] Reconcile the trade-off between Dep's constraint solver and Go module's.
  Dep implicitly downgrades dependencies, searching a multidimensional space
  for one of the possible working solutions, but doesn't reliably converge
  on any particular solution.
  Go modules quickly settle on the newest version of the versions locked by
  your transitive dependencies, regardless of whether there are conflicts
  among their constraints.
  This tool can use the latter strategy, but provide diagnostics so you can
  identify and resolve the problem by fixing either your own constraints
  or the constraints in your dependencies.
- [x] Enable organizational constraints in a shared configuration file like
  "never upgrade Apache Thrift past version 0.9" that assist automatic upgrades
  in a way that keeps your organization consistent across projects.
- [x] Support organizational remote git repository URL rewrite rules for
  mirrors, using the same shared configuration file.
- [ ] Isolate every tool's dependencies in a separate section of the lockfile
  so their transitive dependencies do not interfere with the solution, but
  share the same git cache.

The back-of-the-napkin proof-of-value (“BotN PoV”) for this project is to
assume the orders of magnitude for various factors that impact a company's
bottom line.

- 1,000 engineers
- 10% wasted on Glide or Dep problems
- $100,000/year per engineer
- $10,000,000/yr wasted on glide or dep problems
- 1 year until go modules solve the problem definitively

These parameters suggested that allocating 1, 10, or even 100 engineers to the
problem for a year would save the company money.
However, there are alternative solutions.
Uber has elected to pursue a monorepo and accelerate to Go 1.11 modules
to relieve these issues with Glide and Dep.
Uber then decided to share this solution with the Go community, since it might
be useful elsewhere.
