# TODO

- [ ] Add override method to solver that downgrades every module in the
      solution until it's possible for a particular module to be pinned to a
      downgraded version.  Integrate this in the read-dep-toml workflow to
      enforce overrides.

- [ ] Train read to use any manifest or lock file, and write back the same kind
      found, in order of precedence.

- [ ] Traverse submodules. Fetch would need to investigate .gitmodules in each
      repository root, parse it, fetch each of the contained module URLs,
      and then the tree walker would need to follow commit entries.
      (Use recursive fetch and follow commits in trees)

- [ ] Accumulate fatal errors in solver workflows, particularly failed fetches,
      failure to find any version, and produce a multi-error *with* the partially
      successful solution.

- Automatically read manifests and lock files for other package managers:
  - [x] Read glide.yaml
        add-glide-manifest/agm
        This is not specific enough alone, but could inform a workflow like add-missing.
  - [x] Read glide.lock
        read/r, read-only/ro
  - [x] Read Gopkg.toml
        add-dep-manifest/adm
        Like glide.yaml, this is not specific enough to constitute a lock file,
        but we can use it in a workflow like add-missing to fill in the blanks
        heuristically.
  - [x] Read Gopkg.lock
        read-dep-lock/rdl
  - [ ] Read go.mod
        read-mod-lock/rml
- Automatically generate manifests and lock files for other package managers:
  - [x] Write glide.yaml
        write-glide-yaml/wgy
        This would need to be a lossy operation, checking which members of the
        solution are shallow dependencies, finding the non-empty intersections between
        our own package imports and the exports of each package in the solution,
        then inferring a semver range, falling back to a branch, falling back
        to a hash.
  - [x] Write glide.lock
        write/w, write-only/wo
  - [x] Write Gopkg.toml
        write-dep-toml/wdt
        I don't even.
  - [x] Write Gopkg.lock
        write-dep-lock/wdl
  - [ ] Write go.mod
        write-mod-lock/wml

- [ ] Find modules by hash prefix in the place of a package name.

- [ ] Read config gg.toml files in all parent directories and merge results
      semantically.

- [ ] Sort sections of dependencies report.

- [ ] Add a field to gg.toml that suggests that the cache be in .git instead of
      .gg.

# DONE

- [x] Memoize ReadVersions

- [x] lint

- [x] Trap SIGINT in console mode. Thread cancellation context through commands.

- [x] build context ImportDir has a problem distinguishing multiple packages in
      the same directory.  We can replace it by parsing individual go files in
      the directory and building out the import graph accordingly.

- [x] On "show-module" report, display modules that depend on the displayed
      module, in addition to the modules it depends on.

- [x] Push and pull commands to sync vendor references with a remote repository.

- [x] Add fetch module command to invalidate the cache and refetch.

- [x] Add a field to gg.toml that instructs gg to prefetch vendored versions
      from a git repository's refs/vendor namespace.
      Consider a dual command that selectively pushes updates.

- [o] Whole gopath mode
      (gg works on a whole gopath without modification)

- [o] Detect TTY and provide alternate debouncing progress indicator implementation.
      (instead, provided an explicit quiet option)

- [x] read-dep-toml
- [x] write-dep-toml

- [x] Audit `var` slices, prefer `make()` idiom with initial capacity.

- [x] Use "module.name" instead of "package" throughout.  Package should
      exclusively refer to a package name.

- [o] Make upgrade smarter. If a dependency does not have a reference (as
      written by glide.lock), it might be pinned to a branch that is not "master".
      When finish modules, if there is no reference in glide.lock, attempt
      to upgrade the reference to "master" if the hash is in the history of "master".
      - Decided to that as an alternative, gg would just guess that the
        reference was "master" when a hash does not correspond exactly to any
        current branch or tag. gg will record the ref for any versions it adds
        and this is a good enough guess unless the application was previously
        pinned. Scanning git logs would probably be too slow to warrant.

- [x] Make gg lazier. Create "stages of completion" for a module description.
      - Cached references for the repository have been fetched.
      - Determined corresponding remote location.
      - Commit hash has been normalized and validated
      - Corresponding remote git repository has been fetched and references
        from the fetch have been digested.
      - Package reference has been determined
        - Now we can compare
        - Now we can upgrade
        - Now we can add to a solution
      - Package import graph has been read.
        - Now we can find missing packages.
        - Now we can write this package to the lock cache.
      
- [x] Add a hook to the remotes configuration for creating a git mirror if it
      is missing from gitolite.

- [x] Add a .gg/.gitignore * when creating a cache repository.

- [x] Add back or undo command to revert to previous state.

- [x] Document usage of gg.toml

- Support a configuration file in the context of a project, searching parent
  directories until we find a .gg.yaml or similar.

  - [x] Use RewriteRemote and rewrite rules in a shared configuration file
        found in a parent directory of the project.
  - [x] Organization-specific configured exclude directories in the working copy.
        Persist these exclude directories in the generated lock file to ensure
        consistency of the lock file hash as a cache key.
        Do not use these excludes on packages read from git, also to ensure that
        the module's commit hash is a consistent cache key for the resulting
        import graph.
        Instead, if there are exclude directories recorded in the committed
        glide.lock, we must use these to analyze imports in that package only.
  - [x] With shared constraints, automated workflows would avoid version ranges,
        e.g., newer versions of Apache Thrift, for an entire organization.

- [x] Verify that Deplocks are used if present.

- [x] Add a progress indicator to the add missing modules workflow.

- [x] Figure out why the add-missing workflow reports adding the same package
      repeatedly, as it does in RPS with thriftrw and yarpc.

- [x] Verify that UseAllFiles collects the imports for all architectures
      on the go build context.

- Validation
  - For own packages:
    - [x] Infer own package name from path under GOPATH
  - For git packages:
    - [x] Validate ImportComment and ImportPath at the point of dependency and
          track a warning if they differ.

- [x] Bug: Add missing dependencies needs to recognize itself.  Lottery
      installed itself because it failed to filter itself from the list of missing
      dependencies.
- [x] Debounce progress bar notifications to reduce churn
- [x] trace command to trace from a package to any package in the working copy.
- [x] rm-extra/rx command to remove extra modules.
- [x] write-glide-manifest/wgm command
- [x] show-extra-modules/sxm/sem extra dependencies report
- [x] show-shallow-solution/sss direct dependencies report
- [x] deplock/dl command to show Gopkg.lock
- [x] Read Gopkg.lock in git repositories if present.
- [x] Self-heal bad remotes after multiple failed attempts to fetch.
- [x] Progress indication for gg read command.
- [x] Suppress import comment warnings when finding versions of a remote.
      Warning should only be present when producing a solution.
- [x] Progress indication for gg upgrade command.
- [x] Retry with exponential backoff with full jitter on fetch.
- [x] Add gg git command for inspecting the gg .gg cache.
- [x] Cache shallow module constraints in glide.lock. This will avoid a trip to
      the git repository to collect and parse each package's glide.lock.
- [x] Rename "fix-missing/fm" to "add-missing/am"
- [x] Glide does not normalize tag hashes to the targetted commit hash.  In the
      FinishModule workflow, we should perform that normalization to avoid a
      problem with "show conflicts", which fails to recognize identical
      versions with different hashes. We should also ensure that we normalize
      all hashes before using them as memo cache keys, particularly for looking
      up references to a commit hash.
- [x] Refactor []Module to Modules throughout
- [x] Refactor hashes to plumbing.Hash throughout
- [x] Normalize hashes to commit hashes if they are tag hashes everywhere.
- [x] offline/online/off/on commands to enable or disable offline mode,
      to use the versions already in the cache without attempting to fetch
      updates or new dependencies.
- [x] glidelock/gl inspect glide.lock for module version.
- [x] changelog/cl inspect CHANGELOG.md for module version.
- [x] show-module/sm summary including import comment warning, neighboring constraints
  - [x] Diagnostic to show a module, what it depends on, what depends on it, for
        both tests and normal imports.
- [x] show import comment warning on module string
- [x] persist import comment warning/warnings generally in glide.lock
- [x] smp/show-missing-packages
- [x] Diagnostic to show a package, what it depends on, what depends on it, for
      both tests and normal imports.
- [x] Command line flags driver
- [x] Interactive driver
  - [x] Use a readline module
  - [x] Tab completion
- [x] Add exit command
- [x] Alias "x" without argument at the console to "exit" to avoid aggrevating
      newcomers.

- Workflows
  - [x] Fully automatic missing dependency selection workflow.
    - [x] Build a list of common prefixes for missing imports.
    - [x] Build a list of missing modules by...
      - [x] Using a list of well-known prefixes (ideally checked into a shared
        configuration file in an ancestor directory).
      - [x] Walking up the path until we find a fetchable remote.
    - For each missing module
      - If there is at least one version tagged with a version number
        - [x] Choose the highest version
      - If none of the versions have version number tags
        - [x] Choose master for both version and reference. We will recommend the
              new master for further upgrades.
  - [x] Fully automatic dependency ugprade workflow.
    - For each existing module
      - [x] If the module has an associated version, infer the semver range that
        contains that version, ~0.n or ^n.
      - [x] If the module has an associated reference, fetch the latest version
        of the reference and use that version only if the timestamp is newer.

- [x] fix-missing-modules/fm command
- [x] Address issue that adding a dependency with an extraneous path segment
      like golang.org/x/net/context checks out the golang.org/x/net package at
      vendor/golang.org/x/net/context.
- [x] Per-command usage help
  - [x] Interactive mode
  - [x] Command line mode
- [x] Workflow help for common operations like upgrade/ensure
- [x] Bypass the solver and validation steps in the gg install workflow.
      Just read glide.lock as literal and check out the solution, fetching only
      if the commit is missing locally.
- [x] Restructure "packages" field in glide.lock to capture each exported
      package and each of the packages it imports individually instead of
      collectively.
      This will allow a more useful missing packages algorithm, which
      only considers a package missing if it is depended upon from a
      root package.
      Detecting missing packages will be an input for adding missing modules
      heuristically, which can be overridden manually, so the hueristic doesn't
      have to be perfect.
      The problem is example packages that entrain dependencies
      that no end-consumer will need.
- [x] remove package command
- [x] show imports command (shows imports and test imports for a particular package)
- [x] show missing packages command
- [x] Cache remote-for-package lookups in .gg/.
      Decided to just cache the remote lookups in glide.lock
      and minimize them for well-known remotes.
- [x] Fix "Painter Schlamiel" performance bug in DigestRefs by skipping known
      hashes.
- [x] Cache imports in glide.lock to avoid scanning git for imports each load.
- [x] Report, for each package, which version did every other package depend on.
      show-dependencies.
- [x] Polish for differences between readline and commandline usage,
      like exec arguments.
- [x] A test import should be promoted to an import if it's used as an import,
      and upgrades resolved as usual.
      This is tricky because the non-test imports of a test-import should be
      treated as test-imports until the root promotes the module to non-test.
      Probably implies back-tracking the test-import and bringing it
      forward as a plain import.
      Also implies treating the imports of a test-import as test-imports.
- [x] Use package cache from lockfile to avoid dependency analysis.
- [x] Figure out why the solver looks up remotes for packages when the cached
      lockfile appears to have that information already.
      This can probably be fixed by caching the lockfile
      of each dependency, or by doing a pre-scan of every
      lockfile that collects all the remotes in it before
      finishing each module.
      > Turns out this was a symptom of the solver losing track of some
      dependencies, which were then missing from the lockfile, necessitating a
      lookup again on the next run.
- [x] Figure out why show-conflicts reveals missing dependencies
      that do not show up in show-solution.
- [x] sop/show-own-packages
- [x] Cache exports, imports, and test imports in glide.lock.
      We have to take care not to include the imports and exports
      in the working copy since that would break the possibility of
      using the hash of the glide.lock as a cache key for the corresponding
      vendor directory and packages cacheability.
      Solved this by only merging working copy packages before
      displaying a report.
- [x] State difference viewer. (Show what has been added or upgraded since a
      checkpoint)
- [x] Create a diagnostic that reveals mutual semver incompatibilities in a
      solution and suggests which packages need to be upgraded to resolve the
      conflict.
- [x] Relax dependency on semver. We just need to be able to split vM.m.p tags
      and generalize them to first-significant figure ranges.
- [x] Refactor Lock/Module into internal glidelock and glidelockmodulesmapper
      packags.
- [x] Minimize use of git binary.
- [x] Authenticate http or ssh sessions
      Found hint https://github.com/src-d/go-git/issues/377 for trying again
      later with go-git.  (Resolved by shelling out for fetch and invalidating
      the in memory go-git repo)
- [x] new command should read packages into the base solution.
- [x] Improve show-packages to show each package once on a line and whether
      it's in XIT (exports, imports, test imports)
- [x] Test imports should be tracked separately.
  - [x] Reveal Missing test imports separately.
- [x] Back-fill version number by parsing references as semver. Order imports
      accordingly, prefering version over timestamp.
- [x] Track missing dependencies in state. Do not attempt to automatically
      resolve missing dependencies.
- [x] gg shell command so gg commands can compose with the rest of the Unix
      command suite, with a persistent state server.
      - Run a TCP server on an ephemeral port.
      - Add GG_ADDR environment variable to subshell.
      - Intercept GG_ADDR environment variable in gg main.
      - Forward all commands as a single line to the GG_ADDR server
        and multiplex the response to standard output/errput.

# CRAY

- [ ] Plot dot graph of solution.

- [ ] Encourage an import comment if the remote does not match the import path.

- [ ] Provide JSON and TSV report formats

- [ ] Tab completion in CLI UX.

- [ ] Infer own package name from ImportComment if available.  To do this
      we need to verify that it's possible to compile all of the subpackages
      and record their paths under the ImportComment of the root package.

- [ ] Benchmark and use an immutable data structure with a parent reference for
      frequently copied or cloned structures.

- [ ] Reorder glide.lock in descending order of transitive dependency weight
      for faster dependency discovery with cold caches.

- [ ] gg as a go toolchain wrapper that allows us to fake a GOPATH and
  independent dependency resolution for every tool, never checking out vendor
  in the project root.
  - [ ] gg build
  - [ ] gg exec ... adds .gg/bin to PATH and ensures that the necessary
        tools are built and installed there.

  Consider using an alternate glide.lock and vendor in each command directory.
  Try to infer that the command package has a dependency on a parent package in
  the working copy by observing that it imports packages with a common prefix.

  ```yaml
  glide.lock
  imports:
  - ...
  testImports:
  - ...
  tools:
  - ...
  ```

- [o] Cache git refs in glide.lock
      Elected not to do this. Having the best ref should suffice.
