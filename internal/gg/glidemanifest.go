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

	yaml "gopkg.in/yaml.v2"
)

// GlideManifest is a model for the glide.yaml YAML format.
type GlideManifest struct {
	Package     string                `yaml:"package,omitempty"`
	Homepage    string                `yaml:"homepage,omitempty"`
	License     string                `yaml:"license,omitempty"`
	Imports     []GlideManifestImport `yaml:"import,omitempty"`
	TestImports []GlideManifestImport `yaml:"testImport,omitempty"`
}

// GlideManifestImport models an import or test import in the glide.yaml YAML
// format.
type GlideManifestImport struct {
	Package string `yaml:"package,omitempty"`
	Version string `yaml:"version,omitempty"` // branch, hash, or semver predicate
	Repo    string `yaml:"repo,omitempty"`
}

// ReadOwnGlideManifest reads the glide.yaml in the working copy.
func ReadOwnGlideManifest() (GlideManifest, error) {
	bytes, err := ioutil.ReadFile("glide.yaml")
	if err != nil {
		return GlideManifest{}, err
	}
	return ReadGlideManifest(bytes)
}

// ReadGlideManifest reads a glide.yaml from the given bytes.
func ReadGlideManifest(bytes []byte) (GlideManifest, error) {
	var l GlideManifest
	err := yaml.Unmarshal(bytes, &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

// WriteOwnGlideManifest writes a manifest to the glide.yaml in the working
// copy.
func WriteOwnGlideManifest(manifest *GlideManifest) error {
	bytes, err := WriteGlideManifest(manifest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("glide.yaml", bytes, 0644)
}

// WriteGlideManifest translates a manifest model to bytes.
func WriteGlideManifest(manifest *GlideManifest) ([]byte, error) {
	return yaml.Marshal(manifest)
}
