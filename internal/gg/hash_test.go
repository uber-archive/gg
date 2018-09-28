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
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var (
	averyHashString = "a28ced3c08ed0b0a1cf11c341376b0e5f2bceeed"
	blakeHashString = "a28ced3c08ed0b0a1cf11c341376b0e5f2bceeee"
	careyHashString = "a28ced3c08ed0b0a1cf11c341376b0e5f2bceeef"
	drewHashString  = "a28ced3c08ed0b0a1cf11c341376b0e5f2bceef0"
	averyHash       = plumbing.NewHash(averyHashString)
	blakeHash       = plumbing.NewHash(blakeHashString)
	careyHash       = plumbing.NewHash(careyHashString)
	drewHash        = plumbing.NewHash(drewHashString)
	oneHash         = plumbing.NewHash("0000000000000000000000000000000000000001")
	twoHash         = plumbing.NewHash("0000000000000000000000000000000000000002")
	minusOneHash    = plumbing.NewHash("ffffffffffffffffffffffffffffffffffffffff")
	minusTwoHash    = plumbing.NewHash("fffffffffffffffffffffffffffffffffffffffe")
)

func TestHashString(t *testing.T) {
	assert.Equal(t, HashString(averyHash), averyHashString)
}

func TestNoHashString(t *testing.T) {
	assert.Equal(t, HashString(NoHash), "")
}

func TestParseHashPrefix(t *testing.T) {
	lh, _ := ParseHashPrefix(averyHashString)
	assert.Equal(t, lh, averyHash)
}

func TestParseTooLongHashPrefix(t *testing.T) {
	min, max := ParseHashPrefix(averyHashString + "f")
	assert.Equal(t, min, NoHash)
	assert.Equal(t, max, NoHash)
}

func TestParseBadRuneHashPrefixEven(t *testing.T) {
	hash := "a28ced3c08ed0b0a1cf11c341376b0e5f2bceeeE"
	min, max := ParseHashPrefix(hash)
	assert.Equal(t, min, NoHash)
	assert.Equal(t, max, NoHash)
}

func TestParseBadRuneHashPrefixOdd(t *testing.T) {
	hash := "a28ced3c08ed0b0a1cf11c341376b0e5f2bceeEe"
	min, max := ParseHashPrefix(hash)
	assert.Equal(t, min, NoHash)
	assert.Equal(t, max, NoHash)
}

func TestParseHashPrefixPrefixEven(t *testing.T) {
	minHash := "a28ced3c08ed0000000000000000000000000000"
	maxHash := "a28ced3c08edffffffffffffffffffffffffffff"
	prefix := "a28ced3c08ed"
	min, max := ParseHashPrefix(prefix)
	assert.Equal(t, min.String(), minHash)
	assert.Equal(t, max.String(), maxHash)
}

func TestParseHashPrefixPrefixOdd(t *testing.T) {
	minHash := "a28ced3c08e00000000000000000000000000000"
	maxHash := "a28ced3c08efffffffffffffffffffffffffffff"
	prefix := "a28ced3c08e"
	min, max := ParseHashPrefix(prefix)
	assert.Equal(t, min.String(), minHash)
	assert.Equal(t, max.String(), maxHash)
}

func TestHashBefore(t *testing.T) {
	tests := []struct {
		msg            string
		former, latter plumbing.Hash
		want           bool
	}{
		{"same", averyHash, averyHash, false},
		{"one less", averyHash, blakeHash, true},
		{"one more", blakeHash, averyHash, false},
		{"two less", averyHash, careyHash, true},
		{"two more", careyHash, averyHash, false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, HashBefore(tt.former, tt.latter))
		})
	}
}

func TestHashAfter(t *testing.T) {
	tests := []struct {
		msg            string
		former, latter plumbing.Hash
		want           bool
	}{
		{"same", averyHash, averyHash, false},
		{"one less", averyHash, blakeHash, false},
		{"one more", blakeHash, averyHash, true},
		{"two less", averyHash, careyHash, false},
		{"two more", careyHash, averyHash, true},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, HashAfter(tt.former, tt.latter))
		})
	}
}

func TestSubHash(t *testing.T) {
	tests := []struct {
		msg                  string
		former, latter, want plumbing.Hash
	}{
		{"same", averyHash, averyHash, NoHash},
		{"one less", averyHash, blakeHash, minusOneHash},
		{"one more", blakeHash, averyHash, oneHash},
		{"two less", averyHash, careyHash, minusTwoHash},
		{"two more", careyHash, averyHash, twoHash},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, HashSub(tt.former, tt.latter))
		})
	}
}

func TestDiffHash(t *testing.T) {
	tests := []struct {
		msg                  string
		former, latter, want plumbing.Hash
	}{
		{"same", averyHash, averyHash, NoHash},
		{"one less", averyHash, blakeHash, oneHash},
		{"one more", blakeHash, averyHash, oneHash},
		{"two less", averyHash, careyHash, twoHash},
		{"two more", careyHash, averyHash, twoHash},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, HashDiff(tt.former, tt.latter))
		})
	}
}

func TestHashBetween(t *testing.T) {
	tests := []struct {
		msg           string
		min, got, max plumbing.Hash
		want          bool
	}{
		{
			msg:  "in order",
			min:  averyHash,
			got:  blakeHash,
			max:  careyHash,
			want: true,
		},
		{
			msg:  "on lower bound",
			min:  averyHash,
			got:  averyHash,
			max:  careyHash,
			want: true,
		},
		{
			msg:  "on upper bound",
			min:  averyHash,
			got:  careyHash,
			max:  careyHash,
			want: true,
		},
		{
			msg:  "beneath lower",
			min:  blakeHash,
			got:  averyHash,
			max:  careyHash,
			want: false,
		},
		{
			msg:  "above upper",
			min:  averyHash,
			got:  careyHash,
			max:  blakeHash,
			want: false,
		},
		{
			msg:  "bad order",
			min:  careyHash,
			got:  blakeHash,
			max:  averyHash,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, HashBetween(tt.min, tt.got, tt.max), tt.want)
		})
	}
}
