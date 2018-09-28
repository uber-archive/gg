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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVersion(t *testing.T) {
	table := []struct {
		give string
		want Version
	}{
		{
			"",
			NoVersion,
		},
		{
			"v",
			NoVersion,
		},
		{
			"v0",
			NoVersion,
		},
		{
			"v1",
			Version{1, 0, 0},
		},
		{
			"v0.1",
			Version{0, 1, 0},
		},
		{
			"v0.0.1",
			Version{0, 0, 1},
		},
		{
			"v0.0.0.1",
			NoVersion,
		},
		{
			"v0.0.1-beta1",
			NoVersion,
		},
		{
			"1.2.3",
			Version{1, 2, 3},
		},
		{
			"v10.20.30",
			Version{10, 20, 30},
		},
		{
			"v12.345.678",
			Version{12, 345, 678},
		},
		{
			"01",
			NoVersion,
		},
		{
			"0.0.0",
			NoVersion,
		},
		{
			"bogus",
			NoVersion,
		},
		{
			"1.01",
			NoVersion,
		},
		{
			"1.2.01",
			NoVersion,
		},
		{
			"101",
			Version{101, 0, 0},
		},
	}

	for _, tt := range table {
		t.Run(tt.give, func(t *testing.T) {
			got := ParseVersion(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCanUpgradeTo(t *testing.T) {
	tests := []struct {
		src, dst Version
		want     bool
	}{
		{
			src:  Version{0, 0, 0},
			dst:  Version{0, 0, 0},
			want: false,
		},
		{
			src:  Version{0, 1, 0},
			dst:  Version{0, 1, 0},
			want: false,
		},
		{
			src:  Version{0, 0, 1},
			dst:  Version{0, 0, 1},
			want: false,
		},
		{
			src:  Version{1, 0, 0},
			dst:  Version{1, 0, 0},
			want: false,
		},
		{
			src:  Version{1, 0, 0},
			dst:  Version{2, 0, 0},
			want: false,
		},
		{
			src:  Version{1, 1, 0},
			dst:  Version{1, 2, 0},
			want: true,
		},
		{
			src:  Version{1, 1, 0},
			dst:  Version{2, 0, 0},
			want: false,
		},
		{
			src:  Version{1, 1, 1},
			dst:  Version{1, 1, 2},
			want: true,
		},
		{
			src:  Version{1, 0, 0},
			dst:  Version{1, 1, 0},
			want: true,
		},
		{
			src:  Version{0, 1, 0},
			dst:  Version{0, 2, 0},
			want: false,
		},
		{
			src:  Version{0, 1, 0},
			dst:  Version{0, 1, 1},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.src.String()+"_to_"+tt.dst.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.src.CanUpgradeTo(tt.dst))
		})
	}
}
