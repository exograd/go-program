package program

import (
	"fmt"
	"os"
)

type Program struct {
	Name        string
	Description string
	Main        func(*Program)

	commands      map[string]*Command
	globalOptions map[string]*Option
	arguments     []*Argument

	command *Command
}

type Command struct {
	Name        string
	Description string
	Main        func(*Program)

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

		globalOptions: make(map[string]*Option),
	}
}

func (p *Program) SetMain(main func(*Program)) {
	if len(p.commands) > 0 {
		panic("cannot have a main function with commands")
	}

	p.Main = main
}

func (p *Program) AddCommand(name, description string, main func(*Program)) {
	if p.Main != nil {
		panic("cannot have a main function with commands")
	}

	p.commands[name] = &Command{
		Name:        name,
		Description: description,
		Main:        main,

		options: make(map[string]*Option),
	}
}

func (p *Program) AddGlobalOption(shortName, longName, valueName, defaultValue, description string) {
	option := &Option{
		ShortName:    shortName,
		LongName:     longName,
		ValueName:    valueName,
		DefaultValue: defaultValue,
		Description:  description,
	}

	p.addOption("", option)
}

func (p *Program) AddGlobalFlag(shortName, longName, description string) {
	p.AddGlobalOption(shortName, longName, "", "", description)
}

func (p *Program) AddCommandOption(commandName, shortName, longName, valueName, defaultValue, description string) {
	option := &Option{
		ShortName:    shortName,
		LongName:     longName,
		ValueName:    valueName,
		DefaultValue: defaultValue,
		Description:  description,
	}

	p.addOption(commandName, option)
}

func (p *Program) AddCommandFlag(commandName, shortName, longName, description string) {
	p.AddCommandOption(commandName, shortName, longName, "", "", description)
}

func (p *Program) addOption(commandName string, option *Option) {
	var m map[string]*Option

	if option.ShortName == "" && option.LongName == "" {
		panic("command has no short or long name")
	}

	if commandName == "" {
		m = p.globalOptions
	} else {
		command, found := p.commands[commandName]
		if !found {
			panicf("unknown command %q", commandName)
		}

		m = command.options
	}

	if option.ShortName != "" {
		if _, found := m[option.ShortName]; found {
			panicf("duplicate option name %q", option.ShortName)
		}

		if _, found := p.globalOptions[option.ShortName]; found {
			panicf("duplicate option name %q", option.ShortName)
		}

		m[option.ShortName] = option
	}

	if option.LongName != "" {
		if _, found := m[option.LongName]; found {
			panicf("duplicate option name %q", option.LongName)
		}

		if _, found := p.globalOptions[option.LongName]; found {
			panicf("duplicate option name %q", option.LongName)
		}

		m[option.LongName] = option
	}
}

func (p *Program) AddArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
	}

	p.addArgument("", arg)
}

func (p *Program) AddTrailingArgument(name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
		Trailing:    true,
	}

	p.addArgument("", arg)
}

func (p *Program) AddCommandArgument(commandName, name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
	}

	p.addArgument(commandName, arg)
}

func (p *Program) AddCommandTrailingArgument(commandName, name, description string) {
	arg := &Argument{
		Name:        name,
		Description: description,
		Trailing:    true,
	}

	p.addArgument(commandName, arg)
}

func (p *Program) addArgument(commandName string, arg *Argument) {
	if commandName == "" {
		p.arguments = append(p.arguments, arg)
	} else {
		command, found := p.commands[commandName]
		if !found {
			panic(fmt.Sprintf("unknown command %q", commandName))
		}

		command.arguments = append(command.arguments, arg)
	}
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

	option, found := p.globalOptions[name]
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
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (p *Program) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}

func (p *Program) BlankLine() {
	fmt.Fprintln(os.Stderr, "")
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

	p.run()
}

func (p *Program) run() {
	var main func(*Program)

	if p.command == nil {
		main = p.Main
	} else {
		main = p.command.Main
	}

	main(p)
}

func (p *Program) addDefaultOptions() {
	p.AddGlobalFlag("h", "help", "print help and exit")
}

func (p *Program) addDefaultCommands() {
	p.AddCommand("help", "print help and exit", cmdHelp)
	p.AddCommandTrailingArgument("help", "command",
		"the name of the command(s)")
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
				p.BlankLine()
				p.BlankLine()
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
	p.BlankLine()

	if p.command == nil {
		p.PrintUsage(nil)
	} else {
		p.PrintUsage(p.command)
	}

	os.Exit(1)
}
