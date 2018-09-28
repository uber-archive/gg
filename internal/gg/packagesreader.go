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

// file build.go provides methods for reading the package import graph for the
// working copy and for modules in a Git repository by commit hash.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const packageImportCommentPrefix = "// import "

var gitExcludes = StringSet{
	".git":     {},
	".gg":      {},
	"vendor":   {},
	"testdata": {},
}

// ReadOwnPackages infers the package name and reads the import graph of
// all the packages provided in the working copy.
func ReadOwnPackages(out io.Writer, workDir string, goPath []string, excludes StringSet) (string, Packages, error) {
	// Find the name of the module (or "" for the root of GOPATH).
	var name string
	var srcDir string
	var found bool
	for _, goPath := range goPath {
		goPrefix := goPath + "/src"
		if workDir != goPrefix {
			goPrefix += "/"
			if !strings.HasPrefix(workDir, goPrefix) {
				continue
			}
		}
		srcDir = goPrefix
		name = strings.TrimPrefix(workDir, goPrefix)
		found = true
	}

	if !found {
		return name, Packages{}, fmt.Errorf("The working copy is not in the GOPATH")
	}

	module := &Module{Name: name}
	excludes = excludes.Clone()
	excludes.Include(gitExcludes)
	entry := FSEntry{
		path:  filepath.Join(srcDir, name),
		isDir: true,
	}
	err := readPackages(entry, module, excludes)
	return name, module.Packages, err
}

// ReadGitPackages reads the package import graph for all the Go files in a git
// tree.
// ReadGitPackages also fills in missing details on the given module object,
// like the consistent hash of its changelog, glide.lock, or Gopkg.toml, if
// present as well as a warning if a package's import comment does not match
// the module's expected name.
func ReadGitPackages(out io.Writer, repo *git.Repository, tree *object.Tree, module *Module) error {
	entry := GitEntry{
		name: filepath.Base(module.Name),
		mode: filemode.Dir,
		hash: tree.Hash,
		repo: repo,
	}
	return readPackages(entry, module, gitExcludes)
}

func readPackages(entry TreeEntry, module *Module, excludes StringSet) error {
	module.Packages = NewPackages()
	walker := Walk(filepath.Dir(module.Name), entry)
	for {
		path, entry, err := walker.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if path == module.Name+"/CHANGELOG.md" {
			module.Changelog = entry.Hash()
		} else if path == module.Name+"/glide.lock" {
			module.Glidelock = entry.Hash()
		} else if path == module.Name+"/Gopkg.lock" {
			module.Deplock = entry.Hash()
		} else if entry.IsDir() && excludes.Has(entry.Name()) {
			walker.Skip()
			continue
		} else if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			reader, err := entry.Reader()
			if err != nil {
				return err
			}
			digestGoFile(path, reader, module)
		}
	}
	return nil
}

func digestGoFile(path string, reader io.Reader, module *Module) {
	exp := filepath.Dir(path)

	// Parse file
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, reader, parser.ImportsOnly|parser.ParseComments)
	if err != nil {
		module.Warnings = append(module.Warnings, fmt.Sprintf("Unable to parse Go file %s: %s", path, err))
		return
	}

	// Extract package import comment
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	for _, comment := range cmap[f.Name] {
		for _, comment := range comment.List {
			if strings.HasPrefix(comment.Text, packageImportCommentPrefix) {
				quoted := strings.TrimPrefix(comment.Text, packageImportCommentPrefix)
				if unquoted, err := strconv.Unquote(quoted); err == nil {
					if unquoted != exp {
						module.Warnings = append(module.Warnings, fmt.Sprintf("The package %q must be imported as %q according to its package import comment.", exp, unquoted))
					}
				}
			}
		}
	}

	// Extract imports and exports
	imports := module.Packages.Imports
	coImports := module.Packages.CoImports
	if strings.HasSuffix(path, "_test.go") {
		imports = module.Packages.TestImports
		coImports = module.Packages.CoTestImports
	} else if f.Name.Name == "main" {
		module.Packages.Command(exp)
	} else {
		module.Packages.Export(exp)
	}
	for _, imp := range f.Imports {
		if imp, err := strconv.Unquote(imp.Path.Value); err == nil {
			if !isBuiltin(imp) {
				imports.Add(exp, imp)
				coImports.Add(imp, exp)
			}
		}
	}
}
