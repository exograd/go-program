package main

import (
	"fmt"

	"github.com/galdor/go-program"
)

func main() {
	p := program.NewProgram("no-command",
		"an example program without any command")

	p.AddGlobalFlag("", "flag-a", "a long flag")
	p.AddGlobalFlag("b", "", "a short flag")
	p.AddGlobalOption("c", "option-c", "value", "foo",
		"an option with both a short and long name")

	p.AddCommand("foo", "foo command", cmdFoo)
	p.AddCommandFlag("foo", "d", "flag-d", "a command flag")
	p.AddCommandArgument("foo", "arg-1", "the first argument")
	p.AddCommandArgument("foo", "arg-2", "the second argument")
	p.AddCommandTrailingArgument("foo", "arg-3", "all trailing arguments")

	p.AddCommand("bar", "bar command", cmdBar)

	p.Start()
}

func cmdFoo(p *program.Program) {
	fmt.Println("Foo!")
}

func cmdBar(p *program.Program) {
	fmt.Println("Bar!")
}
