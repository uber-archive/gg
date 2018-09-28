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

// DepManifest models the TOML schema of a Gopkg.toml file.
type DepManifest struct {
	Required    []string                `toml:"required,omitempty"`
	Ignored     []string                `toml:"ignored,omitempty"`
	Constraints []DepManifestConstraint `toml:"constraint,omitempty"`
	Overrides   []DepManifestConstraint `toml:"override,omitempty"`
	Prune       DepManifestPrune        `toml:"prune,omitempty"`
	Metadata    map[string]string       `toml:"metadata,omitempty"`
	NoVerify    []string                `toml:"noverify,omitempty"`
}

// DepManifestConstraint models a Gopkg.toml constraint.
type DepManifestConstraint struct {
	Name   string `toml:"name"`
	Source string `toml:"source,omitempty"` // remote

	// one of
	Branch   string `toml:"branch,omitempty"`
	Revision string `toml:"revision,omitempty"` // hash
	Version  string `toml:"version,omitempty"`  // semver predicate

	Metadata map[string]string `toml:"metadata"`
}

// DepManifestPrune models the prune block in a Gopkg.toml.
type DepManifestPrune struct {
	GoTests        bool `toml:"go-tests,omitempty"`
	UnusedPackages bool `toml:"unused-packages,omitempty"`
}

// ReadDepManifest reads a Gopkg.toml into a DepManifest model.
func ReadDepManifest(bytes []byte) (*DepManifest, error) {
	var manifest DepManifest
	err := toml.Unmarshal(bytes, &manifest)
	return &manifest, err
}

// WriteDepManifest writes a DepManifest model to a Gopkg.toml writer.
func WriteDepManifest(manifest *DepManifest, writer io.Writer) error {
	encoder := toml.NewEncoder(writer)
	return encoder.Encode(manifest)
}

// ReadOwnDepManifest reads the Gopkg.toml in the working copy.
func ReadOwnDepManifest() (*DepManifest, error) {
	file, err := os.Open("Gopkg.toml")
	if err != nil {
		return &DepManifest{}, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return &DepManifest{}, err
	}
	return ReadDepManifest(bytes)
}

// WriteOwnDepManifest writes a Gopkg.toml in the working copy.
func WriteOwnDepManifest(manifest *DepManifest) error {
	file, err := os.Create("Gopkg.toml")
	if err != nil {
		return err
	}
	return WriteDepManifest(manifest, file)
}
