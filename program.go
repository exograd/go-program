// Copyright (c) 2021 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
// SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR
// IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package program

import (
	"fmt"
	"os"
)

type Main func(*Program)

type Program struct {
	Name        string
	Description string
	Main        Main

	commands  map[string]*Command
	options   map[string]*Option
	arguments []*Argument

	command *Command

	verbose    bool
	debugLevel int
}

func NewProgram(name, description string) *Program {
	p := &Program{
		Name:        name,
		Description: description,

		commands: make(map[string]*Command),

		options: make(map[string]*Option),
	}

	p.addDefaultOptions()

	return p
}

func (p *Program) SetMain(main Main) {
	if len(p.commands) > 0 {
		panic("cannot have a main function with commands")
	}

	p.Main = main
}

func (p *Program) Run() {
	var main Main
	if p.command == nil {
		main = p.Main
	} else {
		main = p.command.Main
	}

	main(p)
}

func (p *Program) Debug(level int, format string, args ...interface{}) {
	if level > p.debugLevel {
		return
	}

	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (p *Program) Info(format string, args ...interface{}) {
	if !p.verbose {
		return
	}

	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (p *Program) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}

func (p *Program) Fatal(format string, args ...interface{}) {
	p.Error(format, args...)
	os.Exit(1)
}

func (p *Program) fatal(format string, args ...interface{}) {
	p.Error(format, args...)

	fmt.Fprintf(os.Stderr, "\n")

	if p.command == nil {
		p.PrintUsage(nil)
	} else {
		p.PrintUsage(p.command)
	}

	os.Exit(1)
}
