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
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var errCannotDetectVCS = errors.New("Cannot detect VCS")

// TODO track metrics for RemoteForPackage HTTP request latency and timeouts.
// Raise suspicion if they are mostly timeouts.
// Avoid checking configured well-known package roots.

// RemoteForPackage returns the remote URL for a given Go package name.
//
// Pages like https://golang.org/x/net provide an html document with
// meta tags containing a location to work with. The go tool uses
// a meta tag with the name go-import which is what we use here.
// godoc.org also has one call go-source that we do not need to use.
// The value of go-import is in the form "prefix vcs repo". The prefix
// should match the vcsURL and the repo is a location that can be
// checked out. Note, to get the HTML document you you need to add
// ?go-get=1 to the url.
func RemoteForPackage(pkg string) (string, string) {
	rem := "https://" + pkg
	u, err := url.Parse(rem)
	if err != nil {
		return rem, pkg
	}
	if u.RawQuery == "" {
		u.RawQuery = "go-get=1"
	} else {
		u.RawQuery = u.RawQuery + "&go-get=1"
	}
	checkURL := u.String()
	req, err := http.NewRequest("GET", checkURL, nil)
	if err != nil {
		return rem, pkg
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return rem, pkg
	}
	defer func() {
		_ = res.Body.Close()
	}()

	gotRemote, gotPackage, err := parseImportFromBody(u, res.Body)
	if err != nil || gotRemote == "" || gotPackage == "" {
		return rem, pkg
	}

	return gotRemote, gotPackage
}

func parseImportFromBody(ur *url.URL, r io.ReadCloser) (remote string, pkg string, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var t xml.Token
	for {
		t, err = d.Token()
		if err != nil {
			if err == io.EOF {
				// If we hit the end of the markup and don't have anything
				// we return an error.
				err = errCannotDetectVCS
			}
			return
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {
			// If the prefix supplied by the remote system isn't a prefix to the
			// url we're fetching return continue looking for more go-imports.
			// This will work for exact matches and prefixes. For example,
			// golang.org/x/net as a prefix will match for golang.org/x/net and
			// golang.org/x/net/context.
			vcsURL := ur.Host + ur.Path
			if !strings.HasPrefix(vcsURL, f[0]) {
				continue
			} else {
				pkg = f[0]
				remote = f[2]
				return
			}
		}
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
