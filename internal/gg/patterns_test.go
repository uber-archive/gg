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

func TestPatternSplit(t *testing.T) {
	tests := []struct {
		give string
		want []string
	}{
		{
			give: "github.com/*/*",
			want: []string{"github.com", "/", "*", "/", "*"},
		},
		{
			give: "go.uber.org/*",
			want: []string{"go.uber.org", "/", "*"},
		},
		{
			give: "code.uber.internal/...",
			want: []string{"code.uber.internal", "/", "..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			assert.Equal(t, PatternSplit(tt.give), tt.want)
		})
	}
}

func TestPatternMatch(t *testing.T) {
	patterns := NewPatterns([][2]string{
		{
			"github.com/*/*",
			"ssh://gitolite@code.uber.internal:github/*/*",
		},
		{
			"go.uber.org/*",
			"ssh://gitolite@code.uber.internal:github/uber-go/*",
		},
		{
			"code.uber.internal/...",
			"ssh://gitolite@code.uber.internal:...",
		},
		{
			"example.com",
			"ssh://example.com/*",
		},
	})

	tests := []struct {
		give string
		part string
		want string
		rule int
	}{
		{
			give: "github.com/akshayjshah",
			want: "",
			part: "",
			rule: -1,
		},
		{
			give: "github.com/akshayjshah/negotiate",
			part: "github.com/akshayjshah/negotiate",
			want: "ssh://gitolite@code.uber.internal:github/akshayjshah/negotiate",
			rule: 0,
		},
		{
			give: "github.com/akshayjshah/negotiate/internal/util",
			part: "github.com/akshayjshah/negotiate",
			want: "ssh://gitolite@code.uber.internal:github/akshayjshah/negotiate",
			rule: 0,
		},
		{
			give: "code.uber.internal/infra/lottery",
			part: "code.uber.internal/infra/lottery",
			want: "ssh://gitolite@code.uber.internal:infra/lottery",
			rule: 2,
		},
		{
			give: "go.uber.org/fx",
			part: "go.uber.org/fx",
			want: "ssh://gitolite@code.uber.internal:github/uber-go/fx",
			rule: 1,
		},
		{
			give: "example.com/a/b/c",
			part: "",
			want: "",
			rule: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			part, want, index := patterns.Replace(tt.give)
			assert.Equal(t, part, tt.part)
			assert.Equal(t, want, tt.want)
			assert.Equal(t, index, tt.rule)
		})
	}
}
