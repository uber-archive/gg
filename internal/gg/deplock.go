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
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// DepLock is a model for the Gopkg.lock in TOML format.
type DepLock struct {
	Projects []DepProject `toml:"projects"`
	// SolveMeta DepSolveMeta `toml:"solve-meta"`
}

// DepProject is a model for each "project" entry in a Gopkg.lock file.
type DepProject struct {
	Name string `toml:"name"`
	// Packages []string `toml:"packages"`
	Revision string `toml:"revision,omitempty"` // hash
	Version  string `toml:"version,omitempty"`  // like "v1.0.0", presumably inferred from ref, without "tags/" prefix.
	Branch   string `toml:"branch,omitempty"`   // like "feature", not "heads/feature"
	Source   string `toml:"source,omitempty"`   // remote
}

// type DepSolveMeta struct {
// 	AnalyzerName    string `toml:"analyzer-name"`
// 	AnalyzerVersion int    `toml:"analyzer-version"`
// 	InputsDigest    string `toml:"inputs-digest"`
// 	SolverName      string `toml:"solver-name"`
// 	SolverVersion   int    `toml:"solver-version"`
// }

// ReadDepLock reads a Gopkg.lock into a DepLock model.
func ReadDepLock(bytes []byte) (*DepLock, error) {
	var lock DepLock
	err := toml.Unmarshal(bytes, &lock)
	return &lock, err
}

// WriteDepLock writes a DepLock model to a Gopkg.lock writer.
func WriteDepLock(lock *DepLock, writer io.Writer) error {
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(lock)
}

// ReadOwnDepLock reads the Gopkg.lock in the working copy.
func ReadOwnDepLock() (*DepLock, error) {
	file, err := os.Open("Gopkg.lock")
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return ReadDepLock(bytes)
}

// WriteOwnDepLock writes a Gopkg.lock in the working copy.
func WriteOwnDepLock(lock *DepLock) error {
	file, err := os.Create("Gopkg.lock")
	if err != nil {
		return err
	}
	return WriteDepLock(lock, file)
}
