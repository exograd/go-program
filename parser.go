// Copyright (c) 2021 Nicolas Martyanoff <khaelin@gmail.com>
// Copyright (c) 2022 Exograd SAS.
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
	"os"
	"strings"
)

func (p *Program) parse() {
	args := os.Args[1:]

	args = p.parseOptions(args, p.options)

	if p.IsOptionSet("help") {
		return
	}

	if len(p.commands) > 0 {
		args = p.parseCommand(args)

		options := make(map[string]*Option)
		for name, opt := range p.options {
			options[name] = opt
		}
		for name, opt := range p.command.options {
			options[name] = opt
		}

		args = p.parseOptions(args, options)

		args = p.parseArguments(args, p.command.arguments)
	} else {
		args = p.parseArguments(args, p.arguments)
	}
}

func (p *Program) parseOptions(args []string, options map[string]*Option) []string {
	for len(args) > 0 {
		arg := args[0]

		isShort := len(arg) == 2 && arg[0] == '-' && arg[1] != '-'
		isLong := len(arg) > 2 && arg[0:2] == "--"

		if arg == "--" || !(isShort || isLong) {
			break
		}

		key := strings.TrimLeft(arg, "-")

		opt, found := options[key]
		if !found {
			p.fatal("unknown option %q", key)
		}

		opt.Set = true

		if opt.ValueName == "" {
			args = args[1:]
		} else {
			if len(args) < 2 {
				p.fatal("missing value for option %q", key)
			}

			opt.Value = args[1]

			args = args[2:]
		}
	}

	return args
}

func (p *Program) parseCommand(args []string) []string {
	if len(args) == 0 {
		p.fatal("missing command")
	}

	name := args[0]

	command, found := p.commands[name]
	if !found {
		p.fatal("unknown command %q", name)
	}

	p.command = command

	return args[1:]
}

func (p *Program) parseArguments(args []string, arguments []*Argument) []string {
	if len(arguments) > 0 {
		// Mandatory arguments
		min := 0
		for _, argument := range arguments {
			if argument.Optional || argument.Trailing {
				break
			}

			min++
		}

		if len(args) < min {
			p.fatal("missing argument(s)")
		}

		for i := 0; i < min; i++ {
			argument := arguments[i]

			argument.Set = true
			argument.Value = args[i]
		}

		args = args[min:]
		arguments = arguments[min:]

		// Optional arguments
		var trailingArgument *Argument

		for _, argument := range arguments {
			if len(args) == 0 {
				break
			}

			if argument.Trailing {
				trailingArgument = argument
				break
			}

			argument.Set = true
			argument.Value = args[0]

			args = args[1:]
		}

		// Trailing argument
		if trailingArgument != nil {
			trailingArgument.TrailingValues = args
			args = args[len(args):]
		} else {
			if len(args) > 0 {
				p.fatal("too many arguments")
			}
		}
	}

	return args
}
