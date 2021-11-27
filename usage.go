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
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
)

func (p *Program) PrintUsage(command *Command) {
	var buf bytes.Buffer

	var programName string
	if command == nil {
		programName = os.Args[0]
	} else {
		programName = os.Args[0] + " " + command.Name
	}

	hasCommands := len(p.commands) > 0

	var arguments []*Argument
	var description string

	if command == nil {
		arguments = p.arguments
		description = p.Description
	} else {
		arguments = command.arguments
		description = command.Description
	}

	hasArguments := len(arguments) > 0

	maxWidth := p.computeMaxWidth(command)

	if command == nil && hasCommands {
		fmt.Fprintf(&buf, "Usage: %s OPTIONS <command>\n", programName)
	} else if hasArguments {
		var argBuf bytes.Buffer

		for _, arg := range arguments {
			if arg.Trailing {
				fmt.Fprintf(&argBuf, " [<%s>...]", arg.Name)
			} else if arg.Optional {
				fmt.Fprintf(&argBuf, " [<%s>]", arg.Name)
			} else {
				fmt.Fprintf(&argBuf, " <%s>", arg.Name)
			}
		}

		fmt.Fprintf(&buf, "Usage: %s OPTIONS%s\n", programName,
			argBuf.String())
	} else {
		fmt.Fprintf(&buf, "Usage: %s OPTIONS\n", programName)
	}

	if description != "" {
		fmt.Fprintf(&buf, "\n%s\n", sentence(description))
	}

	if command == nil && hasCommands {
		p.usageCommands(&buf, maxWidth)
	} else if hasArguments {
		p.usageArguments(&buf, arguments, maxWidth)
	}

	if len(p.options) > 0 {
		if command != nil && len(command.options) > 0 {
			p.usageOptions(&buf, "GLOBAL OPTIONS", p.options, maxWidth)
		} else {
			p.usageOptions(&buf, "OPTIONS", p.options, maxWidth)
		}
	}

	if command != nil && len(command.options) > 0 {
		p.usageOptions(&buf, "COMMAND OPTIONS", command.options, maxWidth)
	}

	io.Copy(os.Stderr, &buf)
}

func (p *Program) computeMaxWidth(command *Command) int {
	max := 0

	for _, cmd := range p.commands {
		if len(cmd.Name) > max {
			max = len(cmd.Name)
		}
	}

	var args []*Argument
	if command == nil {
		args = p.arguments
	} else {
		args = command.arguments
	}

	for _, arg := range args {
		if len(arg.Name) > max {
			max = len(arg.Name)
		}
	}

	f := func(opt *Option) {
		length := 2 + 2 + 2 + len(opt.LongName)
		if opt.ValueName != "" {
			length += 2 + len(opt.ValueName) + 1
		}

		if length > max {
			max = length
		}
	}

	for _, opt := range p.options {
		f(opt)
	}

	if command != nil {
		for _, opt := range command.options {
			f(opt)
		}
	}

	return max
}

func (p *Program) usageCommands(buf *bytes.Buffer, maxWidth int) {
	fmt.Fprintf(buf, "\nCOMMANDS\n\n")

	names := []string{}

	for name := range p.commands {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		command := p.commands[name]
		fmt.Fprintf(buf, "%-*s  %s\n", maxWidth, name, command.Description)
	}
}

func (p *Program) usageArguments(buf *bytes.Buffer, args []*Argument, maxWidth int) {
	fmt.Fprintf(buf, "\nARGUMENTS\n\n")

	for _, arg := range args {
		fmt.Fprintf(buf, "%-*s  %s\n", maxWidth, arg.Name, arg.Description)
	}
}

func (p *Program) usageOptions(buf *bytes.Buffer, label string, options map[string]*Option, maxWidth int) {
	fmt.Fprintf(buf, "\n%s\n\n", label)

	strs := make(map[*Option]string)

	for _, opt := range options {
		if _, found := strs[opt]; found {
			continue
		}

		buf := bytes.NewBuffer([]byte{})

		if opt.ShortName == "" {
			fmt.Fprintf(buf, "  ")
		} else {
			fmt.Fprintf(buf, "-%s", opt.ShortName)
		}

		if opt.LongName != "" {
			if opt.ShortName == "" {
				buf.WriteString("  ")
			} else {
				buf.WriteString(", ")
			}

			fmt.Fprintf(buf, "--%s", opt.LongName)
		}

		if opt.ValueName != "" {
			fmt.Fprintf(buf, " <%s>", opt.ValueName)
		}

		str := buf.String()
		strs[opt] = str
	}

	var opts []*Option
	for opt, _ := range strs {
		opts = append(opts, opt)
	}

	sort.Slice(opts, func(i, j int) bool {
		return opts[i].sortKey() < opts[j].sortKey()
	})

	for _, opt := range opts {
		fmt.Fprintf(buf, "%-*s  %s", maxWidth, strs[opt], opt.Description)

		if opt.DefaultValue != "" {
			fmt.Fprintf(buf, " (default: %s)", opt.DefaultValue)
		}

		fmt.Fprintf(buf, "\n")
	}
}

func (opt *Option) sortKey() string {
	if opt.ShortName != "" {
		return opt.ShortName
	}

	if opt.LongName != "" {
		return opt.LongName
	}

	return ""
}
