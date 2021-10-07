package program

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
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
	var options map[string]*Option
	var description string

	if command == nil {
		arguments = p.arguments
		options = p.globalOptions
		description = p.Description
	} else {
		arguments = command.arguments
		options = command.options
		description = command.Description
	}

	hasArguments := len(arguments) > 0
	hasOptions := len(options) > 0

	maxWidth := p.computeMaxWidth(command)

	if command == nil && hasCommands {
		fmt.Fprintf(&buf, "Usage: %s OPTIONS <command>\n", programName)
	} else if hasArguments {
		var argBuf bytes.Buffer

		for _, arg := range arguments {
			if arg.Trailing {
				fmt.Fprintf(&argBuf, " [<%s>...]", arg.Name)
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

	if hasOptions {
		p.usageOptions(&buf, options, maxWidth)
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

	var options map[string]*Option
	if command == nil {
		options = p.globalOptions
	} else {
		options = command.options
	}

	for _, opt := range options {
		length := 1 + len(opt.ShortName) + 2 + 2 + len(opt.LongName)
		if opt.ValueName != "" {
			length += 2 + len(opt.ValueName) + 1
		}

		if length > max {
			max = length
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

func (p *Program) usageOptions(buf *bytes.Buffer, options map[string]*Option, maxWidth int) {
	fmt.Fprintf(buf, "\nOPTIONS\n\n")

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

func (p *Program) Start() {
	p.parse()
	p.run()
}

func (p *Program) parse() {
	args := os.Args[1:]

	args = p.parseOptions(args, p.globalOptions)

	if len(p.commands) > 0 {
		args = p.parseCommand(args)

		options := make(map[string]*Option)
		for name, opt := range p.globalOptions {
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
		last := arguments[len(arguments)-1]

		min := len(arguments)
		if last.Trailing {
			min--
		}

		if len(args) < min {
			p.fatal("missing argument(s)")
		}

		for i := 0; i < min; i++ {
			arguments[i].Value = args[i]
		}

		args = args[min:]

		if last.Trailing {
			last.TrailingValues = args

			args = args[len(args):]
		}
	}

	return args
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

func (opt *Option) sortKey() string {
	if opt.ShortName != "" {
		return opt.ShortName
	}

	if opt.LongName != "" {
		return opt.LongName
	}

	return ""
}
