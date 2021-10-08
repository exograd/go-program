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

	verbose bool
}

type Command struct {
	Name        string
	Description string
	Main        Main

	program *Program

	options   map[string]*Option
	arguments []*Argument
}

type Option struct {
	ShortName    string
	LongName     string
	ValueName    string
	DefaultValue string
	Description  string

	Set   bool
	Value string
}

type Argument struct {
	Name        string
	Description string
	Trailing    bool

	Value          string
	TrailingValues []string
}

func NewProgram(name, description string) *Program {
	return &Program{
		Name:        name,
		Description: description,

		commands: make(map[string]*Command),

		options: make(map[string]*Option),
	}
}

func (p *Program) SetMain(main Main) {
	if len(p.commands) > 0 {
		panic("cannot have a main function with commands")
	}

	p.Main = main
}

func (p *Program) AddCommand(name, description string, main Main) *Command {
	if p.Main != nil {
		panic("cannot have a main function with commands")
	}

	c := &Command{
		Name:        name,
		Description: description,
		Main:        main,

		program: p,

		options: make(map[string]*Option),
	}

	p.commands[name] = c

	return c
}

func (p *Program) AddOption(shortName, longName, valueName, defaultValue, description string) {
	option := &Option{
		ShortName:    shortName,
		LongName:     longName,
		ValueName:    valueName,
		DefaultValue: defaultValue,
		Description:  description,
	}

	p.addOption(nil, option)
}

func (p *Program) AddFlag(shortName, longName, description string) {
	p.AddOption(shortName, longName, "", "", description)
}

func (c *Command) AddOption(shortName, longName, valueName, defaultValue, description string) {
	option := &Option{
		ShortName:    shortName,
		LongName:     longName,
		ValueName:    valueName,
		DefaultValue: defaultValue,
		Description:  description,
	}

	c.program.addOption(c, option)
}

func (c *Command) AddFlag(shortName, longName, description string) {
	c.AddOption(shortName, longName, "", "", description)
}

func (p *Program) addOption(c *Command, option *Option) {
	var m map[string]*Option

	if option.ShortName == "" && option.LongName == "" {
		panic("command has no short or long name")
	}

	if c == nil {
		m = p.options
	} else {
		m = c.options
	}

	if option.ShortName != "" {
		if _, found := m[option.ShortName]; found {
			panicf("duplicate option name %q", option.ShortName)
		}

		if c != nil {
			if _, found := c.program.options[option.ShortName]; found {
				panicf("duplicate option name %q", option.ShortName)
			}
		}

		m[option.ShortName] = option
	}

	if option.LongName != "" {
		if _, found := m[option.LongName]; found {
			panicf("duplicate option name %q", option.LongName)
		}

		if c != nil {
			if _, found := c.program.options[option.LongName]; found {
				panicf("duplicate option name %q", option.LongName)
			}
		}

		m[option.LongName] = option
	}
}

func (p *Program) AddArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
	}

	p.arguments = append(p.arguments, arg)
}

func (p *Program) AddTrailingArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
		Trailing:    true,
	}

	p.arguments = append(p.arguments, arg)
}

func (c *Command) AddArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
	}

	c.arguments = append(c.arguments, arg)
}

func (c *Command) AddTrailingArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
		Trailing:    true,
	}

	c.arguments = append(c.arguments, arg)
}

func (p *Program) CommandName() string {
	if len(p.commands) == 0 {
		panicf("no command defined")
	}

	return p.command.Name
}

func (p *Program) IsOptionSet(name string) bool {
	return p.mustOption(name).Set
}

func (p *Program) OptionValue(name string) string {
	return p.mustOption(name).Value
}

func (p *Program) mustOption(name string) *Option {
	if p.command != nil {
		option, found := p.command.options[name]
		if found {
			return option
		}
	}

	option, found := p.options[name]
	if !found {
		panicf("unknown option %q", name)
	}

	return option
}

func (p *Program) ArgumentValue(name string) string {
	return p.mustArgument(name).Value
}

func (p *Program) TrailingArgumentValues(name string) []string {
	return p.mustArgument(name).TrailingValues
}

func (p *Program) mustArgument(name string) *Argument {
	var arguments []*Argument

	if p.command == nil {
		arguments = p.arguments
	} else {
		arguments = p.command.arguments
	}

	for _, argument := range arguments {
		if name == argument.Name {
			return argument
		}
	}

	panicf("unknown argument %q", name)
	return nil // make the compiler happy
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

func (p *Program) Start() {
	p.addDefaultOptions()

	if len(p.commands) > 0 {
		p.addDefaultCommands()
	}

	p.parse()

	if p.IsOptionSet("help") {
		cmdHelp(p)
		os.Exit(0)
	}

	p.verbose = p.IsOptionSet("verbose")

	p.run()
}

func (p *Program) run() {
	var main Main

	if p.command == nil {
		main = p.Main
	} else {
		main = p.command.Main
	}

	main(p)
}

func (p *Program) addDefaultOptions() {
	p.AddFlag("h", "help", "print help and exit")
	p.AddFlag("v", "verbose", "print status and information messages")
}

func (p *Program) addDefaultCommands() {
	c := p.AddCommand("help", "print help and exit", cmdHelp)
	c.AddTrailingArgument("command", "the name of the command(s)")
}

func cmdHelp(p *Program) {
	var commandNames []string
	if p.command != nil {
		commandNames = p.TrailingArgumentValues("command")
	}

	if len(commandNames) == 0 {
		p.PrintUsage(nil)
	} else {
		for i, commandName := range commandNames {
			if i > 0 {
				fmt.Fprintf(os.Stderr, "\n\n")
			}

			command, found := p.commands[commandName]
			if !found {
				p.Error("unknown command %q", commandName)
				os.Exit(1)
			}

			p.PrintUsage(command)
		}
	}
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
