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
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"go.uber.org/multierr"
)

// StartServer starts a gg server that can receive commands from connecting
// sockets and respond with the command output.
func (driver *Driver) StartServer(env []string) ([]string, func()) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(driver.err, "Failed to start gg server\n")
		return env, func() {}
	}
	env = append(env, fmt.Sprintf("GG_ADDR=%s", listen.Addr().String()))

	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				return
			}
			if err := driver.handleConnection(conn); err != nil {
				fmt.Fprintf(driver.err, "%s\n", err)
			}
			_ = conn.Close()
		}
	}()

	return env, func() {
		_ = listen.Close()
	}
}

func (driver *Driver) handleConnection(conn io.ReadWriter) error {
	buf := bufio.NewReader(conn)

	// Typical argument length will be 1 or 2.
	args := make([]string, 0, 2)
	for {
		arg, err := buf.ReadString(0)
		if err != nil {
			return err
		}
		arg = strings.TrimSuffix(arg, "\000")
		if arg == "" {
			break
		}
		args = append(args, arg)
	}

	prev := driver.out
	driver.out = conn
	err := driver.ExecuteArguments(context.TODO(), args...)
	driver.out = prev
	return err
}

// RunClient opens a connection to a gg server, sends a command, and pipes the
// response output.
func RunClient(memo *Memo, addr string, args []string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Append(err, conn.Close())
	}()
	for _, arg := range args {
		fmt.Fprintf(conn, "%s\000", arg)
	}
	fmt.Fprintf(conn, "\000")
	_, err = io.Copy(os.Stdout, conn)
	return err
}
