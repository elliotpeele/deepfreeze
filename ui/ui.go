/*
 * Copyright (c) Elliot Peele <elliot@bentlogic.net>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Simple package for handling text user inteface interactions.

package ui

import (
	"fmt"
	"io"
	"os"
)

var ui = New()

type UserInterface interface {
	io.Writer
	io.Reader
}

type UI struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

func New() *UI {
	return &UI{
		in:  os.Stdin,
		out: os.Stdout,
		err: os.Stderr,
	}
}

func (ui *UI) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(ui.out, format, a...)
}

func (ui *UI) Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(ui.out, a...)
}

func (ui *UI) Error() *UI {
	return &UI{
		in:  nil,
		out: ui.err,
		err: ui.err,
	}
}

func (ui *UI) Read(p []byte) (n int, err error) {
	return ui.in.Read(p)
}

func (ui *UI) Write(p []byte) (n int, err error) {
	return ui.out.Write(p)
}

func (ui *UI) WriteString(s string) (n int, err error) {
	return io.WriteString(ui.out, s)
}

// Module level UI methods
func Printf(format string, a ...interface{}) (n int, err error) {
	return ui.Printf(format, a...)
}

func Print(a ...interface{}) (n int, err error) {
	return ui.Print(a...)
}

func Error() *UI {
	return ui.Error()
}

func Read(p []byte) (n int, err error) {
	return ui.Read(p)
}

func Write(p []byte) (n int, err error) {
	return ui.Write(p)
}

func WriteString(s string) (n int, err error) {
	return ui.WriteString(s)
}
